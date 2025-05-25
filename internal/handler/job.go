package handler

import (
	"net/http"

	"github.com/dnakolan/worker-pool-service/internal/service"
)

type JobsHandler struct {
	service service.JobsService
}

func NewJobsHandler(service service.JobsService) *JobsHandler {
	return &JobsHandler{service: service}
}

func (h *JobsHandler) CreateJobsHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("Not implemented"))
}

func (h *JobsHandler) ListJobsHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("Not implemented"))
}

func (h *JobsHandler) GetJobsHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("Not implemented"))
}
