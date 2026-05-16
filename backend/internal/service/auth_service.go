package service

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"ai-interview/backend/internal/config"
	"ai-interview/backend/internal/model"
	"ai-interview/backend/internal/repository"
)

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Nickname string `json:"nickname"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UpdateProfileRequest struct {
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatarUrl"`
}

type AuthService struct {
	repo          repository.InterviewRepository
	tokenManager  *TokenManager
	defaultStatus string
}

type TokenManager struct {
	secret      []byte
	expireHours int
}

type tokenPayload struct {
	SubjectType string `json:"subjectType,omitempty"`
	UserID      int64  `json:"userId,omitempty"`
	AdminID     int64  `json:"adminId,omitempty"`
	Email       string `json:"email"`
	ExpiresAt   int64  `json:"expiresAt"`
}

func NewAuthService(repo repository.InterviewRepository, cfg config.AuthConfig) *AuthService {
	expireHours := cfg.TokenExpireHours
	if expireHours <= 0 {
		expireHours = 168
	}

	return &AuthService{
		repo: repo,
		tokenManager: &TokenManager{
			secret:      []byte(cfg.JWTSecret),
			expireHours: expireHours,
		},
		defaultStatus: defaultString(cfg.DefaultUserStatus, "active"),
	}
}

func (s *AuthService) TokenManager() *TokenManager {
	return s.tokenManager
}

func (s *AuthService) Register(ctx context.Context, req RegisterRequest) (model.AuthSession, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	password := strings.TrimSpace(req.Password)
	nickname := strings.TrimSpace(req.Nickname)

	if email == "" || password == "" {
		return model.AuthSession{}, errors.New("email and password are required")
	}
	if len(password) < 6 {
		return model.AuthSession{}, errors.New("password must be at least 6 characters")
	}
	if nickname == "" {
		nickname = deriveNicknameFromEmail(email)
	}

	_, err := s.repo.GetUserByEmail(ctx, email)
	if err == nil {
		return model.AuthSession{}, errors.New("email already exists")
	}
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return model.AuthSession{}, err
	}

	passwordHash, err := hashPassword(password)
	if err != nil {
		return model.AuthSession{}, err
	}

	user, err := s.repo.CreateUser(ctx, model.User{
		Email:        email,
		PasswordHash: passwordHash,
		Nickname:     nickname,
		Status:       s.defaultStatus,
	})
	if err != nil {
		return model.AuthSession{}, err
	}

	token, err := s.tokenManager.SignUser(user.ID, user.Email)
	if err != nil {
		return model.AuthSession{}, err
	}

	user.PasswordHash = ""
	return model.AuthSession{Token: token, User: user}, nil
}

func (s *AuthService) Login(ctx context.Context, req LoginRequest) (model.AuthSession, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	password := strings.TrimSpace(req.Password)
	if email == "" || password == "" {
		return model.AuthSession{}, errors.New("email and password are required")
	}

	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.AuthSession{}, errors.New("email or password is incorrect")
		}
		return model.AuthSession{}, err
	}

	if !verifyPassword(user.PasswordHash, password) {
		return model.AuthSession{}, errors.New("email or password is incorrect")
	}
	if user.Status != "active" {
		return model.AuthSession{}, errors.New("user is disabled")
	}

	token, err := s.tokenManager.SignUser(user.ID, user.Email)
	if err != nil {
		return model.AuthSession{}, err
	}

	user.PasswordHash = ""
	return model.AuthSession{Token: token, User: user}, nil
}

func (s *AuthService) GetCurrentUser(ctx context.Context, userID int64) (model.User, error) {
	if userID == 0 {
		return model.User{}, errors.New("user is not authenticated")
	}

	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.User{}, ErrNotFound
		}
		return model.User{}, err
	}

	user.PasswordHash = ""
	return user, nil
}

func (s *AuthService) UpdateCurrentUser(ctx context.Context, userID int64, req UpdateProfileRequest) (model.User, error) {
	if userID == 0 {
		return model.User{}, errors.New("user is not authenticated")
	}

	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.User{}, ErrNotFound
		}
		return model.User{}, err
	}

	nickname := strings.TrimSpace(req.Nickname)
	if nickname == "" {
		nickname = user.Nickname
	}
	if nickname == "" {
		nickname = deriveNicknameFromEmail(user.Email)
	}

	user.Nickname = nickname
	user.AvatarURL = strings.TrimSpace(req.AvatarURL)

	updatedUser, err := s.repo.UpdateUser(ctx, user)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.User{}, ErrNotFound
		}
		return model.User{}, err
	}

	updatedUser.PasswordHash = ""
	return updatedUser, nil
}

func (m *TokenManager) Sign(userID int64, email string) (string, error) {
	return m.SignUser(userID, email)
}

func (m *TokenManager) SignUser(userID int64, email string) (string, error) {
	return m.sign(tokenPayload{
		SubjectType: "user",
		UserID:      userID,
		Email:       email,
		ExpiresAt:   time.Now().Add(time.Duration(m.expireHours) * time.Hour).Unix(),
	})
}

func (m *TokenManager) SignAdmin(adminID int64, email string) (string, error) {
	return m.sign(tokenPayload{
		SubjectType: "admin",
		AdminID:     adminID,
		Email:       email,
		ExpiresAt:   time.Now().Add(time.Duration(m.expireHours) * time.Hour).Unix(),
	})
}

func (m *TokenManager) Parse(token string) (int64, error) {
	return m.ParseUser(token)
}

func (m *TokenManager) ParseUser(token string) (int64, error) {
	payload, err := m.parsePayload(token)
	if err != nil {
		return 0, err
	}

	if payload.SubjectType != "" && payload.SubjectType != "user" {
		return 0, errors.New("invalid user token")
	}
	if payload.UserID == 0 {
		return 0, errors.New("invalid user token")
	}

	return payload.UserID, nil
}

func (m *TokenManager) ParseAdmin(token string) (int64, error) {
	payload, err := m.parsePayload(token)
	if err != nil {
		return 0, err
	}

	if payload.SubjectType != "admin" || payload.AdminID == 0 {
		return 0, errors.New("invalid admin token")
	}

	return payload.AdminID, nil
}

func (m *TokenManager) sign(payload tokenPayload) (string, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal token payload: %w", err)
	}

	encodedPayload := base64.RawURLEncoding.EncodeToString(raw)
	signature := signTokenValue(m.secret, encodedPayload)
	return encodedPayload + "." + signature, nil
}

func (m *TokenManager) parsePayload(token string) (tokenPayload, error) {
	parts := strings.Split(strings.TrimSpace(token), ".")
	if len(parts) != 2 {
		return tokenPayload{}, errors.New("invalid token")
	}

	expectedSignature := signTokenValue(m.secret, parts[0])
	if subtle.ConstantTimeCompare([]byte(parts[1]), []byte(expectedSignature)) != 1 {
		return tokenPayload{}, errors.New("invalid token")
	}

	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return tokenPayload{}, errors.New("invalid token")
	}

	var payload tokenPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return tokenPayload{}, errors.New("invalid token")
	}
	if payload.ExpiresAt <= time.Now().Unix() {
		return tokenPayload{}, errors.New("token expired")
	}

	return payload, nil
}

func signTokenValue(secret []byte, payload string) string {
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func hashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate password salt: %w", err)
	}

	hash := sha256.Sum256(append(salt, []byte(password)...))
	return hex.EncodeToString(salt) + ":" + hex.EncodeToString(hash[:]), nil
}

func verifyPassword(stored string, password string) bool {
	parts := strings.Split(stored, ":")
	if len(parts) != 2 {
		return false
	}

	salt, err := hex.DecodeString(parts[0])
	if err != nil {
		return false
	}

	hash := sha256.Sum256(append(salt, []byte(password)...))
	return subtle.ConstantTimeCompare([]byte(parts[1]), []byte(hex.EncodeToString(hash[:]))) == 1
}

func deriveNicknameFromEmail(email string) string {
	prefix := strings.TrimSpace(strings.Split(email, "@")[0])
	if prefix == "" {
		return "新用户"
	}
	return prefix
}
