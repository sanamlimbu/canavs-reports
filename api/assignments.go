package api

import (
	"canvas-report/canvas"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/guregu/null/v5"
)

type sectionWithTeachers struct {
	sectionID    int
	sisSectionID string
	teachers     []string
}

type UngradedAssignment struct {
	Account               string    `json:"account"`
	CourseName            string    `json:"course_name"`
	Name                  string    `json:"name"`
	SectionName           string    `json:"section_name"`
	CourseID              int       `json:"course_id"`
	NeedingGradingSection int       `json:"needs_grading_section"`
	Teachers              []string  `json:"teachers"`
	DueAt                 time.Time `json:"due_at"`
	UnlockAt              time.Time `json:"unlock_at"`
	LockAt                time.Time `json:"lock_at"`
	Published             bool      `json:"published"`
	GradebookURL          string    `json:"gradebook_url"`
}

// GetUngradedAssignmentsByCourseID retrieves ungraded assignments in the given course ID.
// Ungraded assignments are organised by each section within the course.
func (c *APIController) GetUngradedAssignmentsByCourseID(w http.ResponseWriter, r *http.Request) {
	courseIDParam := chi.URLParam(r, "course_id")
	if courseIDParam == "" {
		http.Error(w, "course not found", http.StatusNotFound)
		return
	}

	courseID, err := strconv.Atoi(courseIDParam)
	if err != nil {
		http.Error(w, "course not found", http.StatusNotFound)
		return
	}

	if courseID <= 0 {
		http.Error(w, "course not found", http.StatusNotFound)
		return
	}

	course, code, err := c.canvasClient.GetCourseByID(courseID)
	if err != nil {
		http.Error(w, fmt.Sprintf("error fetching course: %d", courseID), code)
		return
	}

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	assignments, code, err := c.canvasClient.GetAssignmentsByCourseID(ctx, courseID, "", canvas.UngradedAssignmentBucket, true)
	if err != nil {
		http.Error(w, fmt.Sprintf("error fetching assignments of course: %d", courseID), code)
	}

	sectionWithTeachersBySectionID := make(map[int]sectionWithTeachers)

	results := make([]*UngradedAssignment, 0)

	for _, assignment := range assignments {
		select {
		case <-ctx.Done():
			http.Error(w, ctx.Err().Error(), http.StatusRequestTimeout)
			return
		default:
			{
				for _, section := range assignment.NeedsGradingCountBySection {

					assignmentDateBySectionID := make(map[int]canvas.AssignmentDate)

					for _, date := range assignment.AllDates {
						if date.SetID.Valid && date.SetType == "CourseSection" {
							assignmentDateBySectionID[int(date.SetID.Int64)] = date // in this case set id is section id
						}
					}

					// no section information at the moment
					if _, ok := sectionWithTeachersBySectionID[section.SectionID]; !ok {

						enrollments, code, err := c.canvasClient.GetEnrollmentsBySectionID(ctx, section.SectionID, nil, []canvas.EnrollmentType{canvas.TeacherEnrollmentType})
						if err != nil {
							http.Error(w, fmt.Sprintf("error fetching enrollments of section: %d", section.SectionID), code)
							return
						}

						teachers := []string{}

						for _, enrollment := range enrollments {
							teachers = append(teachers, enrollment.User.Name)
						}

						st := sectionWithTeachers{
							sectionID: section.SectionID,
							teachers:  teachers,
						}

						// there are teachers in the section
						if len(enrollments) != 0 {
							st.sisSectionID = enrollments[0].SISSectionID.String
						}

						// get section when there is no sis section id
						if st.sisSectionID == "" {
							_section, code, err := c.canvasClient.GetSectionByID(section.SectionID)
							if err != nil {
								http.Error(w, fmt.Sprintf("error fetching section: %d", section.SectionID), code)
							}

							st.sisSectionID = _section.Name
						}

						sectionWithTeachersBySectionID[section.SectionID] = st
					}

					result := &UngradedAssignment{
						Name:                  assignment.Name,
						CourseID:              assignment.CourseID,
						NeedingGradingSection: section.NeedsGradingCount,
						Published:             assignment.Published,
						Account:               course.Account.Name,
						CourseName:            course.Name,
						GradebookURL:          fmt.Sprintf(`%s/courses/%d/gradebook`, c.canvasClient.WebUrl, courseID),
					}

					// now we have section information
					if st, ok := sectionWithTeachersBySectionID[section.SectionID]; ok {
						result.SectionName = st.sisSectionID
						result.Teachers = st.teachers
					}

					// section has date
					if date, ok := assignmentDateBySectionID[section.SectionID]; ok {
						result.DueAt = date.DueAt
						result.LockAt = date.LockAt
						result.UnlockAt = date.UnlockAt
					}

					results = append(results, result)
				}
			}
		}
	}

	if err := json.NewEncoder(w).Encode(&results); err != nil {
		http.Error(w, "error encoding json response", http.StatusInternalServerError)
	}
}

type AssignmentResult struct {
	UserSisID       string     `json:"user_sis_id"`
	Name            string     `json:"name"`
	Acccount        string     `json:"account"`
	CourseName      string     `json:"course_name"`
	Section         string     `json:"section"`
	Title           string     `json:"title"`
	PointsPossible  null.Float `json:"points_possible"`
	Score           null.Float `json:"score"`
	Discrepancy     string     `json:"discrepancy"`
	SubmittedAt     string     `json:"submitted_at"`
	Status          string     `json:"status"`
	DueAt           string     `json:"due_at"`
	CourseState     string     `json:"course_state"`
	EnrollmentRole  string     `json:"enrollment_role"`
	EnrollmentState string     `json:"enrollment_state"`
}

// GetStudentAssignmentsResultByUserID retrieves assignments result of the given student in respective enrolled courses.
func (c *APIController) GetStudentAssignmentsResultByUserID(w http.ResponseWriter, r *http.Request) {
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

	user, code, err := c.canvasClient.GetUserByID(userID)
	if err != nil {
		http.Error(w, fmt.Sprintf("error fetching user: %d", userID), code)
	}

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	courses, code, err := c.canvasClient.GetCoursesByUserID(ctx, userID)
	if err != nil {
		http.Error(w, fmt.Sprintf("error fetching courses of user: %d", userID), code)
		return
	}

	courseByCourseID := make(map[int]*canvas.Course, len(courses))

	for _, course := range courses {
		courseByCourseID[course.ID] = course
	}

	results := make([]*AssignmentResult, 0)

	// for "invited", "rejected", and "deleted", GetAssignmentsDataOfUserByCourseID return 404 error
	// so skip those enrollments
	states := []canvas.EnrollmentState{canvas.ActiveEnrollmentState, canvas.CompletedEnrollmentState}

	enrollments, code, err := c.canvasClient.GetEnrollmentsByUserID(ctx, userID, states)
	if err != nil {
		http.Error(w, fmt.Sprintf("error fetching enrollments of user: %d", userID), code)
		return
	}

loop:
	for _, enrollment := range enrollments {
		select {
		case <-ctx.Done():
			http.Error(w, ctx.Err().Error(), http.StatusRequestTimeout)
			return
		default:
			{
				if enrollment.Role != string(canvas.StudentEnrollmentType) {
					continue loop
				}

				data, code, err := c.canvasClient.GetAssignmentsDataOfUserByCourseID(ctx, userID, enrollment.CourseID)
				if err != nil {
					http.Error(w, fmt.Sprintf("error fetching assignment results of user: %d and course: %d", userID, enrollment.CourseID), code)
					return
				}

				for _, ad := range data {
					result := &AssignmentResult{
						Title:           ad.Title,
						PointsPossible:  ad.PointsPossible,
						DueAt:           ad.DueAt,
						Score:           ad.Submission.Score,
						SubmittedAt:     ad.Submission.SubmittedAt,
						UserSisID:       user.SISUserID,
						Name:            user.Name,
						Section:         enrollment.SISSectionID.String,
						EnrollmentRole:  enrollment.Role,
						EnrollmentState: enrollment.EnrollmentState,
						Status:          ad.Status,
					}

					// Check for situation where student got more marks than possible
					if ad.Submission.Score.Float64 > ad.PointsPossible.Float64 {
						result.Discrepancy = "ERROR"
					}

					if course, ok := courseByCourseID[enrollment.CourseID]; ok {
						result.Acccount = course.Account.Name
						result.CourseName = course.Name
						result.CourseState = course.WorkflowState

					} else {
						course, code, err := c.canvasClient.GetCourseByID(enrollment.CourseID)
						if err != nil {
							http.Error(w, fmt.Sprintf("error fetching course: %d", enrollment.CourseID), code)
							return
						}

						courseByCourseID[enrollment.CourseID] = &course

						result.Acccount = course.Account.Name
						result.CourseName = course.Name
						result.CourseState = course.WorkflowState
					}

					results = append(results, result)
				}
			}
		}
	}

	if err := json.NewEncoder(w).Encode(&results); err != nil {
		http.Error(w, "error encoding json response", http.StatusInternalServerError)
	}
}
