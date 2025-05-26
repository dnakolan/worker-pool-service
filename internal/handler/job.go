package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
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

// extractLastPathSegment returns the last segment of the URL path
func extractLastPathSegment(path string) string {
	segments := strings.Split(path, "/")
	if len(segments) == 0 {
		return ""
	}
	return segments[len(segments)-1]
}

func (h *JobsHandler) GetJobsHandler(w http.ResponseWriter, r *http.Request) {
	jobID := extractLastPathSegment(r.URL.Path)

	if jobID == "" {
		http.Error(w, "invalid job ID: invalid UUID length: 0", http.StatusBadRequest)
		return
	}

	// Validate UUID format before calling service
	_, err := uuid.Parse(jobID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	job, err := h.service.GetJobs(r.Context(), jobID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
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
