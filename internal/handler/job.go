package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/dnakolan/worker-pool-service/internal/model"
	"github.com/dnakolan/worker-pool-service/internal/service"
	"github.com/go-chi/chi/v5"
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
	filter, err := parseFilter(r.URL.Query())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	jobs, err := h.service.ListJobs(r.Context(), filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(jobs)
}

func (h *JobsHandler) GetJobsHandler(w http.ResponseWriter, r *http.Request) {
	uid, err := uuid.Parse(chi.URLParam(r, "uid"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	job, err := h.service.GetJobs(r.Context(), uid.String())
	if err != nil {
		if err.Error() == "job not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}

func parseFilter(query url.Values) (*model.JobFilter, error) {
	var jobType *string
	var jobStatus *model.JobStatus

	// Handle job type
	if typeStr := query.Get("type"); typeStr != "" {
		jobType = &typeStr
	}

	// Handle job status
	if statusStr := query.Get("status"); statusStr != "" {
		if !model.IsValidJobStatus(statusStr) {
			return nil, fmt.Errorf("invalid status: %s", statusStr)
		}
		status := model.JobStatus(statusStr)
		jobStatus = &status
	}

	filter := &model.JobFilter{
		Type:   jobType,
		Status: jobStatus,
	}

	if err := filter.Validate(); err != nil {
		return nil, err
	}

	return filter, nil
}
