package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dnakolan/worker-pool-service/internal/model"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockJobsService is a mock implementation of service.JobsService
type MockJobsService struct {
	mock.Mock
}

func (m *MockJobsService) CreateJobs(ctx context.Context, wp *model.Job) error {
	args := m.Called(ctx, wp)
	return args.Error(0)
}

func (m *MockJobsService) ListJobs(ctx context.Context, filter *model.JobFilter) ([]*model.Job, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Job), args.Error(1)
}

func (m *MockJobsService) GetJobs(ctx context.Context, uid string) (*model.Job, error) {
	args := m.Called(ctx, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Job), args.Error(1)
}

func TestCreateJobsHandler(t *testing.T) {
	mockService := new(MockJobsService)
	handler := NewJobsHandler(mockService)

	tests := []struct {
		name           string
		request        model.CreateJobRequest
		setupMock      func()
		expectedStatus int
	}{
		{
			name: "successful creation",
			request: model.CreateJobRequest{
				Type:    "sleep",
				Payload: json.RawMessage(`{"duration":"1s"}`),
			},
			setupMock: func() {
				mockService.On("CreateJobs", mock.Anything, mock.MatchedBy(func(j *model.Job) bool {
					if j.Type != "sleep" {
						return false
					}
					payload, ok := j.Payload.(model.SleepJobPayload)
					return ok && payload.Duration == "1s"
				})).Return(nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "invalid job type",
			request: model.CreateJobRequest{
				Type:    "invalid",
				Payload: json.RawMessage(`{}`),
			},
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid sleep payload - missing duration",
			request: model.CreateJobRequest{
				Type:    "sleep",
				Payload: json.RawMessage(`{}`),
			},
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid payload format",
			request: model.CreateJobRequest{
				Type:    "sleep",
				Payload: json.RawMessage(`"invalid"`),
			},
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid payload format - string instead of object",
			request: model.CreateJobRequest{
				Type:    "sleep",
				Payload: json.RawMessage(`"invalid"`),
			},
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid payload format - array instead of object",
			request: model.CreateJobRequest{
				Type:    "sleep",
				Payload: json.RawMessage(`[]`),
			},
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid payload format - number instead of object",
			request: model.CreateJobRequest{
				Type:    "sleep",
				Payload: json.RawMessage(`123`),
			},
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid duration format",
			request: model.CreateJobRequest{
				Type:    "sleep",
				Payload: json.RawMessage(`{"duration":123}`), // number instead of string
			},
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			reqBody, err := json.Marshal(tt.request)
			assert.NoError(t, err)
			req := httptest.NewRequest(http.MethodPost, "/jobs", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Create a new chi context with the URL parameter
			rctx := chi.NewRouteContext()
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			handler.CreateJobsHandler(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusCreated {
				var response model.Job
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.NotEmpty(t, response.UID)
				assert.Equal(t, tt.request.Type, response.Type)
				assert.Equal(t, model.JobStatusPending, response.Status)
				payload, ok := response.Payload.(model.SleepJobPayload)
				assert.True(t, ok)
				assert.Equal(t, "1s", payload.Duration)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestGetJobsHandler(t *testing.T) {
	mockService := new(MockJobsService)
	handler := NewJobsHandler(mockService)
	testUID := uuid.New()

	tests := []struct {
		name           string
		uid            string
		setupMock      func()
		expectedStatus int
	}{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			req := httptest.NewRequest(http.MethodGet, "/jobs/"+tt.uid, nil)
			w := httptest.NewRecorder()

			// Create a new chi context with the URL parameter
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("uid", tt.uid)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			handler.GetJobsHandler(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response model.Job
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, testUID, response.UID)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestListJobsHandler(t *testing.T) {
	mockService := new(MockJobsService)
	handler := NewJobsHandler(mockService)
	testUID := uuid.New()
	now := time.Now()

	tests := []struct {
		name           string
		queryParams    map[string]string
		setupMock      func()
		expectedStatus int
		expectedLen    int
	}{{
		name:        "successful list - no filters",
		queryParams: map[string]string{},
		setupMock: func() {
			mockService.On("ListJobs", mock.Anything, mock.MatchedBy(func(f *model.JobFilter) bool {
				return f.Type == nil && f.Status == nil
			})).Return([]*model.Job{
				{
					UID:       testUID,
					Type:      "sleep",
					Payload:   model.SleepJobPayload{Duration: "1s"},
					CreatedAt: &now,
				},
			}, nil)
		},
		expectedStatus: http.StatusOK,
		expectedLen:    1,
	},
		{
			name: "successful list - with filters",
			queryParams: map[string]string{
				"status": "pending",
			},
			setupMock: func() {
				mockService.On("ListJobs", mock.Anything, mock.MatchedBy(func(f *model.JobFilter) bool {
					return *f.Status == "pending" && f.Type == nil
				})).Return([]*model.Job{
					{
						UID:       testUID,
						Type:      "sleep",
						Payload:   model.SleepJobPayload{Duration: "1s"},
						CreatedAt: &now,
					},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedLen:    1,
		},
		{
			name: "invalid filter values",
			queryParams: map[string]string{
				"type": "invalid",
			},
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedLen:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			// Build URL with query parameters
			req := httptest.NewRequest(http.MethodGet, "/jobs", nil)
			q := req.URL.Query()
			for key, value := range tt.queryParams {
				q.Add(key, value)
			}
			req.URL.RawQuery = q.Encode()

			w := httptest.NewRecorder()

			handler.ListJobsHandler(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response []*model.Job
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Len(t, response, tt.expectedLen)
			}

			mockService.AssertExpectations(t)
		})
	}
}
