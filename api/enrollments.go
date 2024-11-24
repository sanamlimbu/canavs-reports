package api

import (
	"canvas-report/canvas"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/guregu/null/v5"
)

type EnrollmentResult struct {
	SISUserID       string      `json:"sis_user_id"`
	StudentName     string      `json:"student_name"`
	AccountName     string      `json:"account_name"`
	CourseName      string      `json:"course_name"`
	SectionName     string      `json:"section_name"`
	EnrollmentState string      `json:"enrollment_state"`
	CourseState     string      `json:"course_state"`
	CurrentGrade    null.String `json:"current_grade"`
	CurrentScore    null.Float  `json:"current_score"`
	EnrollmentRole  string      `json:"enrollment_role"`
	GradesURL       string      `json:"grades_url"`
}

// GetStudentEnrollmentsResultByUserID returns enrollments result of given user ID.
// Only student enrollments of the user is retrieved.
func (c *APIController) GetStudentEnrollmentsResultByUserID(w http.ResponseWriter, r *http.Request) {
	userIDParam := chi.URLParam(r, "user_id")
	if userIDParam == "" {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	userID, err := strconv.Atoi(userIDParam)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	states := canvas.GetOnlyValidEnrollmentState(r.URL.Query()["state[]"])

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	enrollments, code, err := c.canvasClient.GetEnrollmentsByUserID(ctx, userID, states)
	if err != nil {
		http.Error(w, fmt.Sprintf("error fetching enrollments of user: %d", userID), code)
		return
	}

	courses, code, err := c.canvasClient.GetCoursesByUserID(ctx, userID)
	if err != nil {
		http.Error(w, fmt.Sprintf("error fetching courses of user: %d", userID), code)
		return
	}

	results := make([]*EnrollmentResult, len(enrollments))

	courseByCourseID := make(map[int]*canvas.Course, len(courses))

	for _, course := range courses {
		courseByCourseID[course.ID] = course
	}

	for i, enrollment := range enrollments {
		result := &EnrollmentResult{
			SISUserID:       enrollment.User.SISUserID,
			StudentName:     enrollment.User.Name,
			CurrentGrade:    enrollment.Grades.CurrentGrade,
			CurrentScore:    enrollment.Grades.CurrentScore,
			GradesURL:       enrollment.Grades.HtmlUrl,
			EnrollmentState: enrollment.EnrollmentState,
			EnrollmentRole:  enrollment.Role,
			SectionName:     enrollment.SISSectionID.String,
		}

		if course, ok := courseByCourseID[enrollment.CourseID]; ok {
			result.CourseName = course.Name
			result.CourseState = course.WorkflowState
			result.AccountName = course.Account.Name

		} else {
			course, code, err := c.canvasClient.GetCourseByID(enrollment.CourseID)
			if err != nil {
				http.Error(w, fmt.Sprintf("error fetching course: %d", enrollment.CourseID), code)
				return
			}

			courseByCourseID[enrollment.CourseID] = &course

			result.CourseName = course.Name
			result.CourseState = course.WorkflowState
			result.AccountName = course.Account.Name
		}

		if result.SectionName == "" {
			section, code, err := c.canvasClient.GetSectionByID(enrollment.CourseSectionID)
			if err != nil {
				http.Error(w, fmt.Sprintf("error fetching section: %d", enrollment.CourseSectionID), code)
				return
			}

			result.SectionName = section.Name
		}

		results[i] = result
	}

	if err := json.NewEncoder(w).Encode(&results); err != nil {
		http.Error(w, "error encoding json response", http.StatusInternalServerError)
	}
}
