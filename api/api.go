package api

import (
	"canvas-report/canvas"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type APIController struct {
	canvasClient *canvas.CanvasClient
	auther       *auther
}

func NewAPIController(canvasClient *canvas.CanvasClient, auther *auther) (*APIController, error) {
	if canvasClient == nil {
		return nil, fmt.Errorf("missing canvas client")
	}

	controller := &APIController{
		canvasClient: canvasClient,
		auther:       auther,
	}

	return controller, nil
}

func NewRouter(c *APIController, allowedOrigins []string) *chi.Mux {
	r := chi.NewRouter()

	if len(allowedOrigins) == 0 {
		allowedOrigins = []string{"*"}
	}

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "OPTIONS"},
		AllowedHeaders:   []string{"Origin", "X-Requested-With", "Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/", func(r chi.Router) {
		r.Get("/courses/{course_id}/ungraded-assignments", c.GetUngradedAssignmentsByCourseID)
		r.Get("/users/{user_id}/student-enrollments-result", c.GetStudentEnrollmentsResultByUserID)
		r.Get("/users/{user_id}/student-assignments-result", c.GetStudentAssignmentsResultByUserID)
		r.Get("/users/{user_id}/ungraded-assignments", c.GetUngradedAssignmentsByUserID)

		r.Get("/health", healthCheck)
	})

	return r
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"message":"Hello World","time":"%s"}`, time.Now())))
}
