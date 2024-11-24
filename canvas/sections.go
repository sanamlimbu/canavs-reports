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

type Section struct {
	ID            int         `json:"id"`
	CourseID      int         `json:"course_id"`
	Name          string      `json:"name"`
	StartAt       null.Time   `json:"start_at"`
	EndAt         null.Time   `json:"end_at"`
	CreatedAt     time.Time   `json:"created_at"`
	SISSectionID  null.String `json:"sis_section_id"`
	SISCourseID   null.String `json:"sis_course_id"`
	IntegrationID null.String `json:"integration_id"`
	SISImportID   null.String `json:"sis_import_id"`
	TotalStudents null.Int    `json:"total_students"` // total number of active and invited students in the section
}

// GetSectionsByCourseID retrieves sections of the given courseID.
// Total number of active and invited students in the section is included.
func (c *CanvasClient) GetSectionsByCourseID(ctx context.Context, courseID int) ([]*Section, int, error) {
	params := url.Values{}

	params.Add("page", "1")
	params.Add("per_page", strconv.Itoa(c.pageSize))
	params.Add("include[]", "total_students")

	requestUrl := fmt.Sprintf("%s/courses/%d/sections?%s", c.baseUrl, courseID, params.Encode())

	result := make([]*Section, 0)

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
					return nil, res.StatusCode, fmt.Errorf("error fetching sections of course: %d", courseID)
				}

				body, err := io.ReadAll(res.Body)
				res.Body.Close()
				if err != nil {
					return nil, http.StatusInternalServerError, err
				}

				var sections []*Section

				if err := json.Unmarshal(body, &sections); err != nil {
					return nil, http.StatusInternalServerError, err
				}

				result = append(result, sections...)

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

// GetSectionByID retrieves section with given ID.
func (c *CanvasClient) GetSectionByID(sectionID int) (Section, int, error) {
	requestUrl := fmt.Sprintf("%s/sections/%d", c.baseUrl, sectionID)

	req, err := http.NewRequest(http.MethodGet, requestUrl, nil)
	if err != nil {
		return Section{}, http.StatusInternalServerError, err
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return Section{}, http.StatusInternalServerError, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return Section{}, http.StatusInternalServerError, err
	}

	if res.StatusCode != http.StatusOK {
		return Section{}, res.StatusCode, fmt.Errorf("error fetching section: %d", sectionID)
	}

	var section Section

	if err := json.Unmarshal(body, &section); err != nil {
		return Section{}, http.StatusInternalServerError, err
	}

	return section, http.StatusInternalServerError, nil
}
