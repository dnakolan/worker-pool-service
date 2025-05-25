package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/dnakolan/worker-pool-service/internal/model"
	"github.com/dnakolan/worker-pool-service/internal/service"
	"github.com/google/uuid"
)

type JobsHandler struct {
	service service.JobsService
}

func NewJobsHandler(service service.JobsService) *JobsHandler {
	return &JobsHandler{service: service}
}

func (h *JobsHandler) CreateJobsHandler(w http.ResponseWriter, r *http.Request) {
	var req model.CreateJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	payload, err := req.ParsePayload()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	now := time.Now()
	job := &model.Job{
		UID:       uuid.New(),
		Type:      req.Type,
		Payload:   payload,
		Status:    model.JobStatusPending,
		CreatedAt: &now,
	}

	if err := h.service.CreateJobs(r.Context(), job); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(job)
}

func (h *JobsHandler) ListJobsHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("Not implemented"))
}

func (h *JobsHandler) GetJobsHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("Not implemented"))
}
