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

type CourseEnrollmentType string

const (
	TeacherCourseEnrollment  CourseEnrollmentType = "teacher"
	StudentCourseEnrollment  CourseEnrollmentType = "student"
	TaCourseEnrollment       CourseEnrollmentType = "ta"
	ObserverCourseEnrollment CourseEnrollmentType = "observer"
	DesignerCourseEnrollment CourseEnrollmentType = "designer"
)

type CourseWorkflowState string

const (
	ClaimedCourseWorkflowState   CourseWorkflowState = "claimed"
	AvailableCourseWorkflowState CourseWorkflowState = "available"
	DeletedCourseWorkflowState   CourseWorkflowState = "deleted"
)

type Course struct {
	ID                int         `json:"id"`
	CourseCode        string      `json:"course_code"`
	Name              string      `json:"name"`
	SISCourseID       null.String `json:"sis_course_id"`
	GradingStandardID null.Int    `json:"grading_standard_id"`
	AccountID         int         `json:"account_id"`
	RootAccountID     int         `json:"root_account_id"`
	FriendlyName      null.String `json:"friendly_name"`
	WorkflowState     string      `json:"workflow_state"`
	StartAt           null.Time   `json:"start_at"`
	EndAt             null.Time   `json:"end_at"`
	IsPublic          bool        `json:"is_public"`
	EnrollmentTermID  int         `json:"enrollment_term_id"`
	Account           Account     `json:"account"`
	Sections          []Section   `json:"sections"`
}

// GetCourseID retrieves course with given ID.
// Account information of the course is included.
func (c *CanvasClient) GetCourseByID(courseID int) (Course, int, error) {
	params := url.Values{}

	params.Add("include[]", "account")

	requestUrl := fmt.Sprintf("%s/courses/%d?%s", c.baseUrl, courseID, params.Encode())

	req, err := http.NewRequest(http.MethodGet, requestUrl, nil)
	if err != nil {
		return Course{}, http.StatusInternalServerError, err
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return Course{}, http.StatusInternalServerError, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return Course{}, res.StatusCode, fmt.Errorf("error fetching course: %d", courseID)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return Course{}, http.StatusInternalServerError, err
	}

	var course Course

	if err := json.Unmarshal(body, &course); err != nil {
		return course, http.StatusInternalServerError, err
	}

	return course, http.StatusOK, nil
}

// GetCoursesByAccountID retrieves courses for a given account ID.
// Account information of the course is included.
// If "types" is provided, only include courses with at least one user enrolled under one of the specified enrollment types.
func (c *CanvasClient) GetCoursesByAccountID(ctx context.Context, accountID int, courseSearchTerm string, types []CourseEnrollmentType) ([]*Course, int, error) {
	params := url.Values{}

	length := len(courseSearchTerm)

	switch length {
	case 0:
	case 1:
		return nil, http.StatusBadRequest, fmt.Errorf("course search term is less than 2 characters")
	default:
		params.Add("search_term", courseSearchTerm)
	}

	params.Add("per_page", strconv.Itoa(c.pageSize))
	params.Add("include[]", "account")

	for _, t := range types {
		params.Add("enrollment_type[]", string(t))
	}

	requestUrl := fmt.Sprintf("%s/accounts/%d/courses?%s", c.baseUrl, accountID, params.Encode())

	result := make([]*Course, 0)

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
					return nil, res.StatusCode, fmt.Errorf("error fetching courses of account: %d", accountID)
				}

				body, err := io.ReadAll(res.Body)
				res.Body.Close()
				if err != nil {
					return nil, http.StatusInternalServerError, err
				}

				var courses []*Course

				if err := json.Unmarshal(body, &courses); err != nil {
					return nil, http.StatusInternalServerError, err
				}

				result = append(result, courses...)

				nextUrl := getNextUrl(res.Header.Get("Link"))

				if nextUrl == "" {
					break loop
				}

				requestUrl = nextUrl
			}
		}
	}

	return result, http.StatusOK, nil
}

// GetCoursesByUserID retrieves active courses for a given user ID.
// Account and section information are included.
func (c *CanvasClient) GetCoursesByUserID(ctx context.Context, userID int) ([]*Course, int, error) {
	params := url.Values{}

	params.Add("per_page", strconv.Itoa(c.pageSize))
	params.Add("include[]", "account")
	params.Add("include[]", "sections")

	requestUrl := fmt.Sprintf("%s/users/%d/courses?%s", c.baseUrl, userID, params.Encode())

	result := make([]*Course, 0)

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
					return nil, res.StatusCode, fmt.Errorf("error fetching courses of user: %d", userID)
				}

				body, err := io.ReadAll(res.Body)
				res.Body.Close()
				if err != nil {
					return nil, http.StatusInternalServerError, err
				}

				var courses []*Course
				if err := json.Unmarshal(body, &courses); err != nil {
					return nil, http.StatusInternalServerError, err
				}

				result = append(result, courses...)

				nextUrl := getNextUrl(res.Header.Get("Link"))

				if nextUrl == "" {
					break loop
				}

				requestUrl = nextUrl
			}
		}
	}

	return result, http.StatusOK, nil
}
