package canvas

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"time"

	"github.com/guregu/null/v5"
)

type Enrollment struct {
	ID              int       `json:"id"`
	UserID          int       `json:"user_id"`
	CourseID        int       `json:"course_id"`
	Type            string    `json:"type"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	CourseSectionID int       `json:"course_section_id"`
	RootAccountID   int       `json:"root_account_id"`
	EnrollmentState string    `json:"enrollment_state"`
	Role            string    `json:"role"`
	RoleID          int       `json:"role_id"`
	SISImportID     int       `json:"sis_import_id"`
	Grades          struct {
		HtmlUrl      string      `json:"html_url"`
		CurrentScore null.Float  `json:"current_score"`
		CurrentGrade null.String `json:"current_grade"`
		FinalScore   null.Float  `json:"final_score"`
		FinalGrade   null.String `json:"final_grade"`
	} `json:"grades"`
	SISAccountID null.String `json:"sis_account_id"`
	SISCourseID  null.String `json:"sis_course_id"`
	SISSectionID null.String `json:"sis_section_id"`
	User         User        `json:"user"`
}

type EnrollmentType string

const (
	TeacherEnrollmentType  EnrollmentType = "TeacherEnrollment"
	StudentEnrollmentType  EnrollmentType = "StudentEnrollment"
	TaEnrollmentType       EnrollmentType = "TaEnrollment"
	DesignerEnrollmentType EnrollmentType = "DesignerEnrollment"
	ObserverEnrollmentType EnrollmentType = "ObserverEnrollment"
)

type EnrollmentState string

const (
	ActiveEnrollmentState    EnrollmentState = "active"
	InactiveEnrollmentState  EnrollmentState = "inactive"
	CompletedEnrollmentState EnrollmentState = "completed"
	InvitedEnrollmentState   EnrollmentState = "invited"
	RejectedEnrollmentState  EnrollmentState = "rejected"
	DeletedEnrollmentState   EnrollmentState = "deleted"
)

func GetAllEnrollmentState() []EnrollmentState {
	return []EnrollmentState{
		ActiveEnrollmentState, InactiveEnrollmentState, CompletedEnrollmentState, InvitedEnrollmentState, RejectedEnrollmentState, DeletedEnrollmentState,
	}
}

func GetOnlyValidEnrollmentState(states []string) []EnrollmentState {
	result := make([]EnrollmentState, 0, len(states))

	allEnrollmentState := GetAllEnrollmentState()

	for _, state := range states {
		enrollmentState := EnrollmentState(state)

		if slices.Contains(allEnrollmentState, enrollmentState) {
			result = append(result, enrollmentState)
		}
	}

	return result
}

// GetEnrollmentsBySectionID retrieves enrollments in the given section ID.
// Enrollments are filtered based on enrollment states and enrollment type parameters.
func (c *CanvasClient) GetEnrollmentsBySectionID(ctx context.Context, sectionID int, states []EnrollmentState, types []EnrollmentType) ([]*Enrollment, int, error) {
	params := url.Values{}

	params.Add("per_page", strconv.Itoa(c.pageSize))

	for _, state := range states {
		params.Add("state[]", string(state))
	}

	for _, t := range types {
		params.Add("type[]", string(t))
	}

	requestUrl := fmt.Sprintf("%s/sections/%d/enrollments?%s", c.baseUrl, sectionID, params.Encode())

	result := make([]*Enrollment, 0)

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
					return nil, res.StatusCode, fmt.Errorf("error fetching enrollments of course section: %d", sectionID)
				}

				body, err := io.ReadAll(res.Body)
				res.Body.Close()
				if err != nil {
					return nil, http.StatusInternalServerError, err
				}

				var enrollments []*Enrollment

				if err := json.Unmarshal(body, &enrollments); err != nil {
					return nil, http.StatusInternalServerError, err
				}

				result = append(result, enrollments...)

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

// GetEnrollmentsByUserID retrieves enrollments of given user ID.
// Enrollments are filtered based on enrollment states parameters.
func (c *CanvasClient) GetEnrollmentsByUserID(ctx context.Context, userID int, states []EnrollmentState) ([]*Enrollment, int, error) {
	params := url.Values{}

	params.Add("per_page", strconv.Itoa(c.pageSize))

	for _, state := range states {
		params.Add("state[]", string(state))
	}

	requestUrl := fmt.Sprintf("%s/users/%d/enrollments?%s", c.baseUrl, userID, params.Encode())

	result := make([]*Enrollment, 0)

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
					return nil, res.StatusCode, fmt.Errorf("error fetching enrollments of user: %d", userID)
				}

				body, err := io.ReadAll(res.Body)
				res.Body.Close()
				if err != nil {
					return nil, http.StatusInternalServerError, err
				}

				var enrollments []*Enrollment

				if err := json.Unmarshal(body, &enrollments); err != nil {
					return nil, http.StatusInternalServerError, err
				}

				result = append(result, enrollments...)

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
