package canvas

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/guregu/null/v5"
)

type AssignmentBucket string

const (
	PastAssignmentBucket        AssignmentBucket = "past"
	OverAssignmentdueBucket     AssignmentBucket = "overdue"
	UndatedAssignmentBucket     AssignmentBucket = "undated"
	UngradedAssignmentBucket    AssignmentBucket = "ungraded"
	UnsubmittedAssignmentBucket AssignmentBucket = "unsubmitted"
	UpcomingAssignmentBucket    AssignmentBucket = "upcoming"
	FutureAssignmentBucket      AssignmentBucket = "future"
	AllAssignmentBucket         AssignmentBucket = "all"
)

type Assignment struct {
	ID                         int       `json:"id"`
	CourseID                   int       `json:"course_id"`
	Name                       string    `json:"name"`
	DueAt                      null.Time `json:"due_at"`
	UnlockAt                   null.Time `json:"unlock_at"`
	LockAt                     null.Time `json:"lock_at"`
	NeedsGradingCount          int       `json:"needs_grading_count"`
	Published                  bool      `json:"published"`
	HtmlUrl                    string    `json:"html_url"`
	NeedsGradingCountBySection []struct {
		SectionID         int `json:"section_id"`
		NeedsGradingCount int `json:"needs_grading_count"`
	} `json:"needs_grading_count_by_section"`
	AllDates           []AssignmentDate `json:"all_dates"`
	GradingStandardID  null.Int         `json:"grading_standard_id"`
	GradingType        string           `json:"grading_type"`
	OmitFromFinalGrade bool             `json:"omit_from_final_grade"`
	WorkflowState      string           `json:"workflow_state"`
}

type SetType string

const (
	GroupSetType         SetType = "Group"
	CourseSectionSetType SetType = "CourseSection"
	AdhocSetType         SetType = "ADHOC"
)

type AssignmentDate struct {
	ID       null.Int  `json:"id"`
	DueAt    time.Time `json:"due_at"`
	UnlockAt time.Time `json:"unlock_at"`
	LockAt   time.Time `json:"lock_at"`
	Title    string    `json:"title"`
	SetType  string    `json:"set_type"` // "Group", "CourseSection", "ADHOC", "Noop"
	SetID    null.Int  `json:"set_id"`   // set_id is null when set_type is "ADHOC"
	Base     bool      `json:"base"`
}

// GetAssignmentsByCourseID retrieves assignments int the given course ID.
// Search term and assignment bucket: past, ungraded, overdue, etc. are used to filter assignments.
// Needs grading count by section information is included.
func (c *CanvasClient) GetAssignmentsByCourseID(ctx context.Context, courseID int, assignmentSearchTerm string, bucket AssignmentBucket, needsGradingCountBySection bool) ([]*Assignment, int, error) {
	params := url.Values{}

	length := len(assignmentSearchTerm)

	switch length {
	case 0:
	case 1:
		return nil, http.StatusBadRequest, fmt.Errorf("assignments search term is less than 2 characters")
	default:
		params.Add("search_term", assignmentSearchTerm)
	}

	params.Add("page_size", strconv.Itoa(c.pageSize))

	if bucket != AllAssignmentBucket {
		params.Add("bucket", string(bucket))
	}

	if needsGradingCountBySection {
		params.Add("needs_grading_count_by_section", "true")
		params.Add("include[]", "all_dates")
	}

	requestUrl := fmt.Sprintf("%s/courses/%d/assignments?%s", c.baseUrl, courseID, params.Encode())

	result := make([]*Assignment, 0)

loop:
	for {
		select {
		case <-ctx.Done():
			return nil, http.StatusRequestTimeout, ctx.Err()
		default:
			req, err := http.NewRequest(http.MethodGet, requestUrl, nil)
			if err != nil {
				return nil, http.StatusInternalServerError, err
			}

			res, err := c.httpClient.Do(req)
			if err != nil {
				return nil, http.StatusInternalServerError, err
			}

			if res.StatusCode != http.StatusOK {
				return nil, res.StatusCode, fmt.Errorf("error fetching assignments of course: %d", courseID)
			}

			body, err := io.ReadAll(res.Body)
			res.Body.Close()
			if err != nil {
				return nil, http.StatusInternalServerError, err
			}

			var assignments []*Assignment

			if err := json.Unmarshal(body, &assignments); err != nil {
				return nil, http.StatusInternalServerError, err
			}

			result = append(result, assignments...)

			nextUrl := getNextUrl(res.Header.Get("Link"))

			if nextUrl == "" {
				break loop
			}

			requestUrl = nextUrl
		}
	}

	return result, http.StatusOK, nil
}

type AssignmentData struct {
	AssignmentID   int        `json:"assignment_id"`
	Title          string     `json:"title"`
	MaxScore       null.Float `json:"max_score"`
	MinScore       null.Float `json:"min_score"`
	PointsPossible null.Float `json:"points_possible"`
	DueAt          string     `json:"due_at"`
	UnlockAt       string     `json:"unlock_at"`
	Submission     struct {
		Score       null.Float  `json:"score"`
		SubmittedAt string      `json:"submitted_at"`
		PostedAt    null.String `json:"posted_at"`
	} `json:"submission"`
	Status string `json:"status"`
}

func (c *CanvasClient) GetAssignmentsDataOfUserByCourseID(ctx context.Context, userID, courseID int) ([]*AssignmentData, int, error) {
	params := url.Values{}

	params.Add("per_page", strconv.Itoa(c.pageSize))

	requestUrl := fmt.Sprintf("%s/courses/%d/analytics/users/%d/assignments?%s", c.baseUrl, courseID, userID, params.Encode())

	results := make([]*AssignmentData, 0)

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
					return nil, res.StatusCode, fmt.Errorf("error fetching assignment results of user: %d and course: %d", userID, courseID)
				}

				body, err := io.ReadAll(res.Body)
				res.Body.Close()
				if err != nil {
					return nil, http.StatusInternalServerError, err
				}

				var ad []*AssignmentData
				if err := json.Unmarshal(body, &ad); err != nil {
					return nil, http.StatusInternalServerError, err
				}

				results = append(results, ad...)

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
