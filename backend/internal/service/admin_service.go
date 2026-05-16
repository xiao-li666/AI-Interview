package service

import (
	"context"
	"errors"
	"strings"

	"ai-interview/backend/internal/config"
	"ai-interview/backend/internal/model"
	"ai-interview/backend/internal/repository"
)

type AdminLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AdminUpdateUserStatusRequest struct {
	Status string `json:"status"`
}

type AdminResetUserPasswordRequest struct {
	NewPassword string `json:"newPassword"`
}

type AdminService struct {
	repo         repository.AdminRepository
	tokenManager *TokenManager
}

func NewAdminService(repo repository.AdminRepository, cfg config.AuthConfig) *AdminService {
	expireHours := cfg.TokenExpireHours
	if expireHours <= 0 {
		expireHours = 168
	}

	return &AdminService{
		repo: repo,
		tokenManager: &TokenManager{
			secret:      []byte(cfg.JWTSecret),
			expireHours: expireHours,
		},
	}
}

func (s *AdminService) TokenManager() *TokenManager {
	return s.tokenManager
}

func (s *AdminService) EnsureSeedAdmin(ctx context.Context, email string, password string, nickname string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	password = strings.TrimSpace(password)
	nickname = strings.TrimSpace(nickname)
	if email == "" || password == "" {
		return nil
	}
	if nickname == "" {
		nickname = "系统管理员"
	}

	_, err := s.repo.GetAdminByEmail(ctx, email)
	if err == nil {
		return nil
	}
	if !errors.Is(err, repository.ErrNotFound) {
		return err
	}

	passwordHash, err := hashPassword(password)
	if err != nil {
		return err
	}

	_, err = s.repo.CreateAdmin(ctx, model.Admin{
		Email:        email,
		PasswordHash: passwordHash,
		Nickname:     nickname,
		Status:       "active",
	})
	return err
}

func (s *AdminService) Login(ctx context.Context, req AdminLoginRequest) (model.AdminSession, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	password := strings.TrimSpace(req.Password)
	if email == "" || password == "" {
		return model.AdminSession{}, errors.New("email and password are required")
	}

	admin, err := s.repo.GetAdminByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.AdminSession{}, errors.New("email or password is incorrect")
		}
		return model.AdminSession{}, err
	}
	if !verifyPassword(admin.PasswordHash, password) {
		return model.AdminSession{}, errors.New("email or password is incorrect")
	}
	if admin.Status != "active" {
		return model.AdminSession{}, errors.New("admin is disabled")
	}

	token, err := s.tokenManager.SignAdmin(admin.ID, admin.Email)
	if err != nil {
		return model.AdminSession{}, err
	}

	admin.PasswordHash = ""
	return model.AdminSession{Token: token, Admin: admin}, nil
}

func (s *AdminService) GetCurrentAdmin(ctx context.Context, adminID int64) (model.Admin, error) {
	if adminID == 0 {
		return model.Admin{}, errors.New("admin is not authenticated")
	}

	admin, err := s.repo.GetAdminByID(ctx, adminID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.Admin{}, ErrNotFound
		}
		return model.Admin{}, err
	}

	admin.PasswordHash = ""
	return admin, nil
}

func (s *AdminService) GetOverview(ctx context.Context) (model.AdminOverview, error) {
	return s.repo.GetAdminOverview(ctx)
}

func (s *AdminService) ListUsers(ctx context.Context, query model.AdminListQuery) (model.PagedResult[model.AdminUserListItem], error) {
	return s.repo.ListAdminUsers(ctx, normalizeAdminListQuery(query))
}

func (s *AdminService) GetUserDetail(ctx context.Context, userID int64) (model.AdminUserDetail, error) {
	if userID == 0 {
		return model.AdminUserDetail{}, errors.New("user id is required")
	}

	detail, err := s.repo.GetAdminUserDetail(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.AdminUserDetail{}, ErrNotFound
		}
		return model.AdminUserDetail{}, err
	}
	detail.User.PasswordHash = ""
	return detail, nil
}

func (s *AdminService) UpdateUserStatus(ctx context.Context, userID int64, req AdminUpdateUserStatusRequest) (model.User, error) {
	if userID == 0 {
		return model.User{}, errors.New("user id is required")
	}

	status := strings.TrimSpace(req.Status)
	if status != "active" && status != "disabled" {
		return model.User{}, errors.New("status must be active or disabled")
	}

	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.User{}, ErrNotFound
		}
		return model.User{}, err
	}

	user.Status = status
	updated, err := s.repo.UpdateUser(ctx, user)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.User{}, ErrNotFound
		}
		return model.User{}, err
	}

	updated.PasswordHash = ""
	return updated, nil
}

func (s *AdminService) ResetUserPassword(ctx context.Context, userID int64, req AdminResetUserPasswordRequest) error {
	if userID == 0 {
		return errors.New("user id is required")
	}

	password := strings.TrimSpace(req.NewPassword)
	if len(password) < 6 {
		return errors.New("newPassword must be at least 6 characters")
	}

	passwordHash, err := hashPassword(password)
	if err != nil {
		return err
	}

	err = s.repo.UpdateUserPassword(ctx, userID, passwordHash)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNotFound
		}
	}
	return err
}

func (s *AdminService) ListInterviewSessions(ctx context.Context, query model.AdminListQuery) (model.PagedResult[model.AdminInterviewSessionListItem], error) {
	return s.repo.ListAdminInterviewSessions(ctx, normalizeAdminListQuery(query))
}

func (s *AdminService) GetInterviewSessionDetail(ctx context.Context, sessionID int64) (model.AdminInterviewSessionDetail, error) {
	if sessionID == 0 {
		return model.AdminInterviewSessionDetail{}, errors.New("session id is required")
	}

	detail, err := s.repo.GetSession(ctx, sessionID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.AdminInterviewSessionDetail{}, ErrNotFound
		}
		return model.AdminInterviewSessionDetail{}, err
	}

	user, err := s.repo.GetUserByID(ctx, detail.Session.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.AdminInterviewSessionDetail{}, ErrNotFound
		}
		return model.AdminInterviewSessionDetail{}, err
	}

	return model.AdminInterviewSessionDetail{
		User: model.AdminUserSummary{
			ID:       user.ID,
			Email:    user.Email,
			Nickname: user.Nickname,
			Status:   user.Status,
		},
		Session:   detail.Session,
		Questions: detail.Questions,
	}, nil
}

func (s *AdminService) GetInterviewSessionAnswers(ctx context.Context, sessionID int64) ([]model.SessionAnswerItem, error) {
	if sessionID == 0 {
		return nil, errors.New("session id is required")
	}

	items, err := s.repo.ListSessionAnswers(ctx, sessionID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return items, nil
}

func (s *AdminService) GetInterviewReport(ctx context.Context, sessionID int64) (model.InterviewReport, error) {
	if sessionID == 0 {
		return model.InterviewReport{}, errors.New("session id is required")
	}

	report, err := s.repo.GetReport(ctx, sessionID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.InterviewReport{}, ErrNotFound
		}
		return model.InterviewReport{}, err
	}
	return report, nil
}

func (s *AdminService) ListResumes(ctx context.Context, query model.AdminListQuery) (model.PagedResult[model.AdminResumeListItem], error) {
	return s.repo.ListAdminResumes(ctx, normalizeAdminListQuery(query))
}

func (s *AdminService) GetResumeDetail(ctx context.Context, resumeID int64) (model.AdminResumeDetail, error) {
	if resumeID == 0 {
		return model.AdminResumeDetail{}, errors.New("resume id is required")
	}

	detail, err := s.repo.GetAdminResumeDetail(ctx, resumeID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.AdminResumeDetail{}, ErrNotFound
		}
		return model.AdminResumeDetail{}, err
	}
	return detail, nil
}

func (s *AdminService) ListAIConfigs(ctx context.Context, query model.AdminListQuery) (model.PagedResult[model.AdminAIConfigListItem], error) {
	return s.repo.ListAdminAIConfigs(ctx, normalizeAdminListQuery(query))
}

func normalizeAdminListQuery(query model.AdminListQuery) model.AdminListQuery {
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}
	if query.PageSize > 100 {
		query.PageSize = 100
	}
	query.Keyword = strings.TrimSpace(query.Keyword)
	query.Status = strings.TrimSpace(query.Status)
	return query
}
