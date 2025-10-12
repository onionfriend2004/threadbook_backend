package deliveryHTTP

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/goccy/go-json"
	"github.com/onionfriend2004/threadbook_backend/internal/lib"
	"github.com/onionfriend2004/threadbook_backend/internal/lib/middleware/auth"
	"github.com/onionfriend2004/threadbook_backend/internal/spool/delivery/dto"
	"github.com/onionfriend2004/threadbook_backend/internal/spool/usecase"
	"go.uber.org/zap"
)

func (h *SpoolHandler) GetSpoolMembers(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("GetSpoolMembers called")

	// --- 1. Получаем userID из контекста ---
	userID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		h.logger.Warn("unauthorized: no user id in context", zap.Error(err))
		lib.WriteError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	h.logger.Debug("user id from context", zap.Uint("user_id", userID))

	// --- 2. Читаем spoolID из URL ---
	spoolIDStr := chi.URLParam(r, "spoolID")
	spoolIDInt, err := strconv.Atoi(spoolIDStr)
	if err != nil || spoolIDInt < 0 {
		h.logger.Warn("invalid spool_id", zap.String("spool_id_raw", spoolIDStr), zap.Error(err))
		lib.WriteError(w, "invalid spool_id", http.StatusBadRequest)
		return
	}
	spoolID := uint(spoolIDInt)
	h.logger.Debug("parsed spool id", zap.Uint("spool_id", spoolID))

	// --- 3. Вызываем usecase ---
	users, err := h.usecase.GetSpoolMembers(r.Context(), usecase.GetSpoolMembersInput{
		UserID:  userID,
		SpoolID: spoolID,
	})
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrForbidden):
			h.logger.Warn("user not a member of spool", zap.Uint("user_id", userID), zap.Uint("spool_id", spoolID))
			lib.WriteError(w, "forbidden", http.StatusForbidden)
			return

		case errors.Is(err, usecase.ErrNotFound):
			h.logger.Warn("spool not found", zap.Uint("spool_id", spoolID))
			lib.WriteError(w, "spool not found", http.StatusNotFound)
			return

		default:
			h.logger.Error("failed to get spool members", zap.Error(err))
			lib.WriteError(w, "internal error", http.StatusInternalServerError)
			return
		}
	}

	h.logger.Info("successfully fetched spool members",
		zap.Uint("spool_id", spoolID),
		zap.Uint("user_id", userID),
		zap.Int("members_count", len(users)),
	)

	// --- 4. Формируем ответ ---
	resp := dto.GetSpoolMembersResponse{}
	for _, u := range users {
		resp.Members = append(resp.Members, dto.MemberShortInfo{
			Username: u.Username,
			// Avatar:   u.AvatarLink,
		})
	}

	// --- 5. Отправляем JSON ---
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(lib.StatusOK)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
		return
	}

	h.logger.Info("response sent successfully",
		zap.Uint("spool_id", spoolID),
		zap.Int("members_count", len(resp.Members)),
	)
}
