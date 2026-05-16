package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"ai-interview/backend/internal/model"
	"ai-interview/backend/internal/service"
	"ai-interview/backend/pkg/response"
)

func (h *APIHandler) handleAdminAuthLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.MethodNotAllowed(w)
		return
	}

	var req service.AdminLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	result, err := h.adminService.Login(r.Context(), req)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, result)
}

func (h *APIHandler) handleAdminAuthLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.MethodNotAllowed(w)
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (h *APIHandler) handleAdminAuthMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.MethodNotAllowed(w)
		return
	}

	admin, err := h.adminService.GetCurrentAdmin(r.Context(), currentAdminID(r.Context()))
	if err != nil {
		handleServiceError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, admin)
}

func (h *APIHandler) handleAdminOverview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.MethodNotAllowed(w)
		return
	}

	data, err := h.adminService.GetOverview(r.Context())
	if err != nil {
		handleServiceError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, data)
}

func (h *APIHandler) handleAdminUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.MethodNotAllowed(w)
		return
	}

	items, err := h.adminService.ListUsers(r.Context(), parseAdminListQuery(r))
	if err != nil {
		handleServiceError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, items)
}

func (h *APIHandler) handleAdminUserActions(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/admin/users/")
	path = strings.Trim(path, "/")
	if path == "" {
		response.NotFound(w, "route not found")
		return
	}

	parts := strings.Split(path, "/")
	userID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		response.BadRequest(w, "invalid user id")
		return
	}

	if len(parts) == 1 && r.Method == http.MethodGet {
		detail, getErr := h.adminService.GetUserDetail(r.Context(), userID)
		if getErr != nil {
			handleServiceError(w, getErr)
			return
		}

		response.JSON(w, http.StatusOK, detail)
		return
	}

	if len(parts) == 2 && parts[1] == "status" && r.Method == http.MethodPut {
		var req service.AdminUpdateUserStatusRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.BadRequest(w, "invalid request body")
			return
		}

		user, updateErr := h.adminService.UpdateUserStatus(r.Context(), userID, req)
		if updateErr != nil {
			handleServiceError(w, updateErr)
			return
		}

		response.JSON(w, http.StatusOK, user)
		return
	}

	if len(parts) == 2 && parts[1] == "reset-password" && r.Method == http.MethodPost {
		var req service.AdminResetUserPasswordRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.BadRequest(w, "invalid request body")
			return
		}

		if err := h.adminService.ResetUserPassword(r.Context(), userID, req); err != nil {
			handleServiceError(w, err)
			return
		}

		response.JSON(w, http.StatusOK, map[string]any{"ok": true})
		return
	}

	response.NotFound(w, "route not found")
}

func (h *APIHandler) handleAdminInterviewSessions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.MethodNotAllowed(w)
		return
	}

	items, err := h.adminService.ListInterviewSessions(r.Context(), parseAdminListQuery(r))
	if err != nil {
		handleServiceError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, items)
}

func (h *APIHandler) handleAdminInterviewSessionActions(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/admin/interview-sessions/")
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
		detail, getErr := h.adminService.GetInterviewSessionDetail(r.Context(), sessionID)
		if getErr != nil {
			handleServiceError(w, getErr)
			return
		}

		response.JSON(w, http.StatusOK, detail)
		return
	}

	if len(parts) == 2 && parts[1] == "answers" && r.Method == http.MethodGet {
		items, getErr := h.adminService.GetInterviewSessionAnswers(r.Context(), sessionID)
		if getErr != nil {
			handleServiceError(w, getErr)
			return
		}

		response.JSON(w, http.StatusOK, items)
		return
	}

	response.NotFound(w, "route not found")
}

func (h *APIHandler) handleAdminInterviewReports(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.MethodNotAllowed(w)
		return
	}

	sessionID, err := parseTrailingID(r.URL.Path, "/api/v1/admin/interview-reports/")
	if err != nil {
		response.BadRequest(w, "invalid session id")
		return
	}

	report, getErr := h.adminService.GetInterviewReport(r.Context(), sessionID)
	if getErr != nil {
		handleServiceError(w, getErr)
		return
	}

	response.JSON(w, http.StatusOK, report)
}

func (h *APIHandler) handleAdminResumes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.MethodNotAllowed(w)
		return
	}

	items, err := h.adminService.ListResumes(r.Context(), parseAdminListQuery(r))
	if err != nil {
		handleServiceError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, items)
}

func (h *APIHandler) handleAdminResumeActions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.MethodNotAllowed(w)
		return
	}

	resumeID, err := parseTrailingID(r.URL.Path, "/api/v1/admin/resumes/")
	if err != nil {
		response.BadRequest(w, "invalid resume id")
		return
	}

	detail, getErr := h.adminService.GetResumeDetail(r.Context(), resumeID)
	if getErr != nil {
		handleServiceError(w, getErr)
		return
	}

	response.JSON(w, http.StatusOK, detail)
}

func (h *APIHandler) handleAdminAIConfigs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.MethodNotAllowed(w)
		return
	}

	items, err := h.adminService.ListAIConfigs(r.Context(), parseAdminListQuery(r))
	if err != nil {
		handleServiceError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, items)
}

func parseAdminListQuery(r *http.Request) model.AdminListQuery {
	query := r.URL.Query()
	page, _ := strconv.Atoi(query.Get("page"))
	pageSize, _ := strconv.Atoi(query.Get("pageSize"))

	return model.AdminListQuery{
		Page:     page,
		PageSize: pageSize,
		Keyword:  query.Get("keyword"),
		Status:   query.Get("status"),
	}
}
