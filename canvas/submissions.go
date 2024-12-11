package canvas

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/guregu/null/v5"
)

type SubmissionWorkflowState string

const (
	GradedSubmissionWorkflowState        SubmissionWorkflowState = "graded"
	SubmittedSubmissionWorkflowState     SubmissionWorkflowState = "submitted"
	UnsubmittedSubmissionWorkflowState   SubmissionWorkflowState = "unsubmitted"
	PendingReviewSubmissionWorkflowState SubmissionWorkflowState = "pending_review"
)

type Submission struct {
	ID                            int         `json:"id"`
	UserID                        int         `json:"user_id"`
	AssignmentID                  int         `json:"assignment_id"`
	Grade                         null.String `json:"grade"`
	Score                         null.Float  `json:"score"`
	SubmittedAt                   null.String `json:"submitted_at"`
	Attempt                       null.Int    `json:"attempt"`
	WorkflowState                 string      `json:"workflow_state"`
	GradeMatchesCurrentSubmission bool        `json:"grade_matches_current_submission"`
	GradedAt                      null.String `json:"graded_at"`
	GraderID                      null.Int    `json:"grader_id"`
	Late                          bool        `json:"late"`
	Excused                       null.Bool   `json:"excused"`
	Assignment                    struct {
		ID             int        `json:"id"`
		PointsPossible null.Float `json:"points_possible"`
		Name           string     `json:"name"`
	} `json:"assignment"`
}

func (c *CanvasClient) GetSubmissionsByCourseID(ctx context.Context, courseID int, studentID int, submissionWorkflowState SubmissionWorkflowState) ([]*Submission, int, error) {
	params := url.Values{}

	params.Add("page", "1")
	params.Add("per_page", strconv.Itoa(c.pageSize))
	params.Add("student_ids[]", strconv.Itoa(studentID))
	params.Add("include[]", "assignment")
	params.Add("workflow_state", string(submissionWorkflowState))

	requestUrl := fmt.Sprintf("%s/courses/%d/students/submissions?%s", c.baseUrl, courseID, params.Encode())

	results := make([]*Submission, 0)

loop:
	for {
		select {
		case <-ctx.Done():
			return nil, http.StatusRequestTimeout, ctx.Err()
		default:
			{
				req, err := http.NewRequest(http.MethodGet, requestUrl, nil)
				if err != nil {
					return nil, http.StatusInternalServerError, err
				}

				res, err := c.httpClient.Do(req)
				if err != nil {
					return nil, http.StatusInternalServerError, err
				}

				if res.StatusCode != http.StatusOK {
					return nil, res.StatusCode, fmt.Errorf("error fetching submissions of course: %d and student: %d", courseID, studentID)
				}

				body, err := io.ReadAll(res.Body)
				res.Body.Close()
				if err != nil {
					return nil, http.StatusInternalServerError, err
				}

				var submissions []*Submission

				if err := json.Unmarshal(body, &submissions); err != nil {
					return nil, http.StatusInternalServerError, err
				}

				results = append(results, submissions...)

				nextUrl := getNextUrl(res.Header.Get("Link"))
				if nextUrl == "" {
					break loop
				}

				requestUrl = nextUrl
			}
		}
	}

	return results, http.StatusOK, nil
}
