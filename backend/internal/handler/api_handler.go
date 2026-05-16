package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"ai-interview/backend/internal/service"
	"ai-interview/backend/pkg/response"
)

type APIHandler struct {
	service      *service.InterviewService
	authService  *service.AuthService
	adminService *service.AdminService
	resumeParser *service.ResumeParser
}

func NewAPIHandler(
	service *service.InterviewService,
	authService *service.AuthService,
	adminService *service.AdminService,
	resumeParser *service.ResumeParser,
) *APIHandler {
	return &APIHandler{
		service:      service,
		authService:  authService,
		adminService: adminService,
		resumeParser: resumeParser,
	}
}

func (h *APIHandler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", h.handleHealth)
	mux.HandleFunc("/api/v1/auth/register", h.handleAuthRegister)
	mux.HandleFunc("/api/v1/auth/login", h.handleAuthLogin)
	mux.HandleFunc("/api/v1/auth/logout", h.handleAuthLogout)
	mux.HandleFunc("/api/v1/auth/me", h.handleAuthMe)
	mux.HandleFunc("/api/v1/admin/auth/login", h.handleAdminAuthLogin)
	mux.HandleFunc("/api/v1/admin/auth/logout", h.handleAdminAuthLogout)
	mux.HandleFunc("/api/v1/admin/auth/me", h.handleAdminAuthMe)
	mux.HandleFunc("/api/v1/admin/overview", h.handleAdminOverview)
	mux.HandleFunc("/api/v1/admin/users", h.handleAdminUsers)
	mux.HandleFunc("/api/v1/admin/users/", h.handleAdminUserActions)
	mux.HandleFunc("/api/v1/admin/interview-sessions", h.handleAdminInterviewSessions)
	mux.HandleFunc("/api/v1/admin/interview-sessions/", h.handleAdminInterviewSessionActions)
	mux.HandleFunc("/api/v1/admin/interview-reports/", h.handleAdminInterviewReports)
	mux.HandleFunc("/api/v1/admin/resumes", h.handleAdminResumes)
	mux.HandleFunc("/api/v1/admin/resumes/", h.handleAdminResumeActions)
	mux.HandleFunc("/api/v1/admin/ai-configs", h.handleAdminAIConfigs)
	mux.HandleFunc("/api/v1/interview-plans", h.handleInterviewPlans)
	mux.HandleFunc("/api/v1/interview-sessions", h.handleInterviewSessions)
	mux.HandleFunc("/api/v1/interview-sessions/", h.handleInterviewSessionActions)
	mux.HandleFunc("/api/v1/interview-answers", h.handleInterviewAnswers)
	mux.HandleFunc("/api/v1/interview-reports/", h.handleInterviewReports)
	mux.HandleFunc("/api/v1/history", h.handleHistory)
	mux.HandleFunc("/api/v1/ai-config", h.handleAIConfig)
	mux.HandleFunc("/api/v1/ai-config/test", h.handleAIConfigTest)
	mux.HandleFunc("/api/v1/resume/parse", h.handleResumeParse)

	return withMiddleware(mux, h.authService, h.adminService)
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

func (h *APIHandler) handleAuthRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.MethodNotAllowed(w)
		return
	}

	var req service.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	result, err := h.authService.Register(r.Context(), req)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, result)
}

func (h *APIHandler) handleAuthLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.MethodNotAllowed(w)
		return
	}

	var req service.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	result, err := h.authService.Login(r.Context(), req)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, result)
}

func (h *APIHandler) handleAuthLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.MethodNotAllowed(w)
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"ok": true,
	})
}

func (h *APIHandler) handleAuthMe(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		user, err := h.authService.GetCurrentUser(r.Context(), currentUserID(r.Context()))
		if err != nil {
			handleServiceError(w, err)
			return
		}

		response.JSON(w, http.StatusOK, user)
	case http.MethodPut:
		var req service.UpdateProfileRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.BadRequest(w, "invalid request body")
			return
		}

		user, err := h.authService.UpdateCurrentUser(r.Context(), currentUserID(r.Context()), req)
		if err != nil {
			handleServiceError(w, err)
			return
		}

		response.JSON(w, http.StatusOK, user)
	default:
		response.MethodNotAllowed(w)
	}
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
	req.UserID = currentUserID(r.Context())

	plan, err := h.service.CreatePlan(r.Context(), req)
	if err != nil {
		handleServiceError(w, err)
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
	req.UserID = currentUserID(r.Context())

	detail, err := h.service.CreateSession(r.Context(), req)
	if err != nil {
		handleServiceError(w, err)
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
		detail, getErr := h.service.GetSession(r.Context(), currentUserID(r.Context()), sessionID)
		if getErr != nil {
			handleServiceError(w, getErr)
			return
		}

		response.JSON(w, http.StatusOK, detail)
		return
	}

	if len(parts) == 2 && parts[1] == "answers" && r.Method == http.MethodGet {
		items, getErr := h.service.GetSessionAnswers(r.Context(), currentUserID(r.Context()), sessionID)
		if getErr != nil {
			handleServiceError(w, getErr)
			return
		}

		response.JSON(w, http.StatusOK, items)
		return
	}

	if len(parts) == 2 && parts[1] == "question-reviews" && r.Method == http.MethodGet {
		items, getErr := h.service.GetQuestionReviews(r.Context(), currentUserID(r.Context()), sessionID)
		if getErr != nil {
			handleServiceError(w, getErr)
			return
		}

		response.JSON(w, http.StatusOK, items)
		return
	}

	if len(parts) == 3 && parts[1] == "questions" && parts[2] == "next" && r.Method == http.MethodPost {
		question, getErr := h.service.GetNextQuestion(r.Context(), currentUserID(r.Context()), sessionID)
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

	answer, feedback, err := h.service.SubmitAnswer(r.Context(), currentUserID(r.Context()), req)
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

	report, getErr := h.service.GetReport(r.Context(), currentUserID(r.Context()), sessionID)
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

	items, getErr := h.service.ListHistory(r.Context(), currentUserID(r.Context()))
	if getErr != nil {
		handleServiceError(w, getErr)
		return
	}

	response.JSON(w, http.StatusOK, items)
}

func (h *APIHandler) handleAIConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		cfg, getErr := h.service.GetUserAIConfig(r.Context(), currentUserID(r.Context()))
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
		req.UserID = currentUserID(r.Context())

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
	req.UserID = currentUserID(r.Context())

	result, testErr := h.service.TestUserAIConfig(r.Context(), req)
	if testErr != nil {
		handleServiceError(w, testErr)
		return
	}

	response.JSON(w, http.StatusOK, result)
}

func (h *APIHandler) handleResumeParse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.MethodNotAllowed(w)
		return
	}

	if h.resumeParser == nil {
		response.BadRequest(w, "resume parser is not available")
		return
	}

	if err := r.ParseMultipartForm(16 << 20); err != nil {
		response.BadRequest(w, "invalid multipart form")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		response.BadRequest(w, "resume file is required")
		return
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		response.BadRequest(w, "failed to read resume file")
		return
	}

	result, err := h.resumeParser.Parse(r.Context(), header.Filename, fileBytes)
	if err != nil {
		response.BadRequest(w, err.Error())
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
