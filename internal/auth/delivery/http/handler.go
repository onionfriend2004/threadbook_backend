package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/onionfriend2004/threadbook_backend/internal/auth/domain"
	"github.com/onionfriend2004/threadbook_backend/internal/auth/service"
)

// Handler обрабатывает HTTP-запросы.
type Handler struct {
	authService *service.AuthService
}

func NewHandler(authService *service.AuthService) *Handler {
	return &Handler{authService: authService}
}

// authErrorToHTTPStatus маппит AuthError на HTTP-статус.
func authErrorToHTTPStatus(err error) int {
	if !domain.IsAuthError(err) {
		return http.StatusInternalServerError
	}

	var authErr *domain.AuthError
	if !errors.As(err, &authErr) {
		return http.StatusInternalServerError
	}

	switch authErr.Code {
	case domain.ErrCodeUserAlreadyExists:
		return http.StatusConflict
	case domain.ErrCodeInvalidCredentials, domain.ErrCodeUserNotFound:
		return http.StatusUnauthorized
	case domain.ErrCodeInvalidEmail, domain.ErrCodeWeakPassword:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	userID, err := h.authService.Register(ctx, req.Email, req.Password)
	if err != nil {
		status := authErrorToHTTPStatus(err)
		http.Error(w, err.Error(), status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"user_id": userID})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	if err := h.authService.Login(ctx, req.Email, req.Password); err != nil {
		status := authErrorToHTTPStatus(err)
		http.Error(w, err.Error(), status)
		return
	}

	w.WriteHeader(http.StatusOK)
}
