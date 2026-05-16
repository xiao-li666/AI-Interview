package handler

import (
	"log"
	"net/http"
	"strings"
	"time"

	"ai-interview/backend/internal/service"
	"ai-interview/backend/pkg/response"
)

func withMiddleware(next http.Handler, authService *service.AuthService, adminService *service.AdminService) http.Handler {
	return recoverMiddleware(loggingMiddleware(corsMiddleware(authMiddleware(next, authService, adminService))))
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startedAt := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(startedAt))
	})
}

func recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func authMiddleware(next http.Handler, authService *service.AuthService, adminService *service.AdminService) http.Handler {
	userPublicPrefixes := []string{
		"/healthz",
		"/api/v1/auth/register",
		"/api/v1/auth/login",
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/v1/admin/") {
			if r.URL.Path == "/api/v1/admin/auth/login" {
				next.ServeHTTP(w, r)
				return
			}

			if adminService == nil {
				response.Unauthorized(w, "admin service is not available")
				return
			}

			token := extractBearerToken(r.Header.Get("Authorization"))
			if token == "" {
				response.Unauthorized(w, "admin authorization token is required")
				return
			}

			adminID, err := adminService.TokenManager().ParseAdmin(token)
			if err != nil {
				response.Unauthorized(w, err.Error())
				return
			}

			admin, err := adminService.GetCurrentAdmin(r.Context(), adminID)
			if err != nil {
				response.Unauthorized(w, "admin authorization is invalid")
				return
			}
			if admin.Status != "active" {
				response.Unauthorized(w, "admin is disabled")
				return
			}

			next.ServeHTTP(w, r.WithContext(withAuthAdminID(r.Context(), adminID)))
			return
		}

		for _, prefix := range userPublicPrefixes {
			if strings.HasPrefix(r.URL.Path, prefix) {
				next.ServeHTTP(w, r)
				return
			}
		}

		if authService == nil {
			response.Unauthorized(w, "auth service is not available")
			return
		}

		token := extractBearerToken(r.Header.Get("Authorization"))
		if token == "" {
			response.Unauthorized(w, "authorization token is required")
			return
		}

		userID, err := authService.TokenManager().ParseUser(token)
		if err != nil {
			response.Unauthorized(w, err.Error())
			return
		}

		user, err := authService.GetCurrentUser(r.Context(), userID)
		if err != nil {
			response.Unauthorized(w, "authorization is invalid")
			return
		}
		if user.Status != "active" {
			response.Unauthorized(w, "user is disabled")
			return
		}

		next.ServeHTTP(w, r.WithContext(withAuthUserID(r.Context(), userID)))
	})
}

func extractBearerToken(headerValue string) string {
	const prefix = "Bearer "
	if !strings.HasPrefix(headerValue, prefix) {
		return ""
	}

	return strings.TrimSpace(strings.TrimPrefix(headerValue, prefix))
}
