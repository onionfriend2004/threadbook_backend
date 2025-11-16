package deliveryHTTP

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/goccy/go-json"
	"github.com/onionfriend2004/threadbook_backend/internal/apperrors"
	"github.com/onionfriend2004/threadbook_backend/internal/lib"
	"github.com/onionfriend2004/threadbook_backend/internal/lib/middleware/auth"
	"github.com/onionfriend2004/threadbook_backend/internal/thread/delivery/dto"
	"github.com/onionfriend2004/threadbook_backend/internal/thread/usecase"
	"go.uber.org/zap"
)

func (h *ThreadHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "thread_id")
	threadID64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		lib.WriteError(w, "invalid thread id", lib.StatusBadRequest)
		return
	}
	threadID := uint(threadID64)

	// --- 2. Курсор и лимит ---
	cursorID := uint(0)
	if cStr := r.URL.Query().Get("cursor_id"); cStr != "" {
		if c, err := strconv.ParseUint(cStr, 10, 64); err == nil && c > 0 {
			cursorID = uint(c)
		}
	}

	limit := 15 // дефолтный лимит
	if lStr := r.URL.Query().Get("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 {
			limit = l
		}
	}

	forward := true // направление загрузки (true — новые после курсора, false — старые)
	if fStr := r.URL.Query().Get("forward"); fStr != "" {
		forward = fStr == "true"
	}

	userID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		lib.WriteError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	input := usecase.GetMessagesInput{
		UserID:   userID,
		ThreadID: threadID,
		CursorID: cursorID,
		Limit:    limit,
		Forward:  forward,
	}

	msgs, err := h.messageUsecase.GetMessages(r.Context(), input)
	if err != nil {
		code, clientErr := apperrors.GetErrAndCodeToSend(err)
		h.logger.Warn("failed to get messages", zap.Error(err))
		lib.WriteError(w, clientErr.Error(), code)
		return
	}

	// --- 4. Формируем ответ ---
	resp := make([]dto.MessageResponse, 0, len(msgs))
	for _, m := range msgs {
		payloadLinks := make([]string, len(m.Payloads))
		for i, p := range m.Payloads {
			payloadLinks[i] = p.FileLink
		}

		resp = append(resp, dto.MessageResponse{
			ID:        m.ID,
			ThreadID:  m.ThreadID,
			Username:  m.User.Username,
			Content:   m.Content,
			Payloads:  payloadLinks,
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
		})
	}

	// --- 5. Отправка JSON ---
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(lib.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Warn("failed to encode response", zap.Error(err))
	}
}
