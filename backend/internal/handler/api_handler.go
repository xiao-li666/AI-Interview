package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"ai-interview/backend/internal/service"
	"ai-interview/backend/pkg/response"
)

type APIHandler struct {
	service *service.InterviewService
}

func NewAPIHandler(service *service.InterviewService) *APIHandler {
	return &APIHandler{service: service}
}

func (h *APIHandler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", h.handleHealth)
	mux.HandleFunc("/api/v1/interview-plans", h.handleInterviewPlans)
	mux.HandleFunc("/api/v1/interview-sessions", h.handleInterviewSessions)
	mux.HandleFunc("/api/v1/interview-sessions/", h.handleInterviewSessionActions)
	mux.HandleFunc("/api/v1/interview-answers", h.handleInterviewAnswers)
	mux.HandleFunc("/api/v1/interview-reports/", h.handleInterviewReports)
	mux.HandleFunc("/api/v1/history", h.handleHistory)
	mux.HandleFunc("/api/v1/ai-config", h.handleAIConfig)
	mux.HandleFunc("/api/v1/ai-config/test", h.handleAIConfigTest)

	return withMiddleware(mux)
}

func (h *APIHandler) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.MethodNotAllowed(w)
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"status": "ok",
	})
}

func (h *APIHandler) handleInterviewPlans(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.MethodNotAllowed(w)
		return
	}

	var req service.CreatePlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	plan, err := h.service.CreatePlan(r.Context(), req)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, plan)
}

func (h *APIHandler) handleInterviewSessions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.MethodNotAllowed(w)
		return
	}

	var req service.CreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	detail, err := h.service.CreateSession(r.Context(), req)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, detail)
}

func (h *APIHandler) handleInterviewSessionActions(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/interview-sessions/")
	path = strings.Trim(path, "/")

	if path == "" {
		response.NotFound(w, "route not found")
		return
	}

	parts := strings.Split(path, "/")
	sessionID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		response.BadRequest(w, "invalid session id")
		return
	}

	if len(parts) == 1 && r.Method == http.MethodGet {
		detail, getErr := h.service.GetSession(r.Context(), sessionID)
		if getErr != nil {
			handleServiceError(w, getErr)
			return
		}

		response.JSON(w, http.StatusOK, detail)
		return
	}

	if len(parts) == 2 && parts[1] == "answers" && r.Method == http.MethodGet {
		items, getErr := h.service.GetSessionAnswers(r.Context(), sessionID)
		if getErr != nil {
			handleServiceError(w, getErr)
			return
		}

		response.JSON(w, http.StatusOK, items)
		return
	}

	if len(parts) == 3 && parts[1] == "questions" && parts[2] == "next" && r.Method == http.MethodPost {
		question, getErr := h.service.GetNextQuestion(r.Context(), sessionID)
		if getErr != nil {
			handleServiceError(w, getErr)
			return
		}

		response.JSON(w, http.StatusOK, question)
		return
	}

	response.NotFound(w, "route not found")
}

func (h *APIHandler) handleInterviewAnswers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.MethodNotAllowed(w)
		return
	}

	var req service.SubmitAnswerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	answer, feedback, err := h.service.SubmitAnswer(r.Context(), req)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, map[string]any{
		"answer":   answer,
		"feedback": feedback,
	})
}

func (h *APIHandler) handleInterviewReports(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.MethodNotAllowed(w)
		return
	}

	sessionID, err := parseTrailingID(r.URL.Path, "/api/v1/interview-reports/")
	if err != nil {
		response.BadRequest(w, "invalid session id")
		return
	}

	report, getErr := h.service.GetReport(r.Context(), sessionID)
	if getErr != nil {
		handleServiceError(w, getErr)
		return
	}

	response.JSON(w, http.StatusOK, report)
}

func (h *APIHandler) handleHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.MethodNotAllowed(w)
		return
	}

	userIDValue := r.URL.Query().Get("userId")
	userID, err := strconv.ParseInt(userIDValue, 10, 64)
	if err != nil {
		response.BadRequest(w, "invalid userId")
		return
	}

	items, getErr := h.service.ListHistory(r.Context(), userID)
	if getErr != nil {
		handleServiceError(w, getErr)
		return
	}

	response.JSON(w, http.StatusOK, items)
}

func (h *APIHandler) handleAIConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		userIDValue := r.URL.Query().Get("userId")
		userID, err := strconv.ParseInt(userIDValue, 10, 64)
		if err != nil {
			response.BadRequest(w, "invalid userId")
			return
		}

		cfg, getErr := h.service.GetUserAIConfig(r.Context(), userID)
		if getErr != nil {
			handleServiceError(w, getErr)
			return
		}

		response.JSON(w, http.StatusOK, cfg)
	case http.MethodPut:
		var req service.SaveAIConfigRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.BadRequest(w, "invalid request body")
			return
		}

		cfg, saveErr := h.service.SaveUserAIConfig(r.Context(), req)
		if saveErr != nil {
			handleServiceError(w, saveErr)
			return
		}

		response.JSON(w, http.StatusOK, cfg)
	default:
		response.MethodNotAllowed(w)
	}
}

func (h *APIHandler) handleAIConfigTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.MethodNotAllowed(w)
		return
	}

	var req service.TestAIConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	result, testErr := h.service.TestUserAIConfig(r.Context(), req)
	if testErr != nil {
		handleServiceError(w, testErr)
		return
	}

	response.JSON(w, http.StatusOK, result)
}

func parseTrailingID(path string, prefix string) (int64, error) {
	value := strings.TrimPrefix(path, prefix)
	value = strings.Trim(value, "/")
	if value == "" {
		return 0, errors.New("missing id")
	}

	return strconv.ParseInt(value, 10, 64)
}

func handleServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrNotFound):
		response.NotFound(w, err.Error())
	default:
		response.BadRequest(w, err.Error())
	}
}
