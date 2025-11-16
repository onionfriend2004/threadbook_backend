package deliveryHTTP

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/goccy/go-json"
	"github.com/onionfriend2004/threadbook_backend/internal/apperrors"
	"github.com/onionfriend2004/threadbook_backend/internal/lib"
	"github.com/onionfriend2004/threadbook_backend/internal/lib/middleware/auth"
	"github.com/onionfriend2004/threadbook_backend/internal/thread/delivery/dto"
	"github.com/onionfriend2004/threadbook_backend/internal/thread/usecase"
	"go.uber.org/zap"
)

func (h *ThreadHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "thread_id")
	threadID64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		lib.WriteError(w, "invalid thread id", http.StatusBadRequest)
		return
	}
	threadID := uint(threadID64)

	// 1. Парсим multipart
	if err := r.ParseMultipartForm(h.fileConfig.GetMaxSize("message_files")); err != nil {
		lib.WriteError(w, "failed to parse form data", http.StatusBadRequest)
		return
	}
	defer func() {
		if r.MultipartForm != nil {
			r.MultipartForm.RemoveAll()
		}
	}()

	// 2. Получаем userID и username
	userID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		lib.WriteError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	username, err := auth.GetUsernameFromContext(r.Context())
	if err != nil {
		lib.WriteError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// 3. Текст сообщения
	content := strings.TrimSpace(r.FormValue("content"))
	if content == "" && len(r.MultipartForm.File["files"]) == 0 {
		lib.WriteError(w, "message must contain text or files", http.StatusBadRequest)
		return
	}

	// 4. Обработка файлов (может быть много)
	var payloads []*usecase.FilePayload
	files := r.MultipartForm.File["files"]

	for _, fileHeader := range files {
		if !h.fileConfig.ValidateSize("message_files", fileHeader.Size) {
			maxSizeMB := h.fileConfig.GetMaxSize("message_files") >> 20
			lib.WriteError(w, fmt.Sprintf("file %s exceeds %dMB", fileHeader.Filename, maxSizeMB), http.StatusBadRequest)
			return
		}

		if !h.fileConfig.IsAllowedFormat(fileHeader.Filename) {
			allowed := strings.Join(h.fileConfig.GetAllowedFormats(), ", ")
			lib.WriteError(w, fmt.Sprintf("allowed formats: %s", allowed), http.StatusBadRequest)
			return
		}

		file, err := fileHeader.Open()
		if err != nil {
			lib.WriteError(w, fmt.Sprintf("failed to open file %s", fileHeader.Filename), http.StatusBadRequest)
			return
		}
		defer file.Close()

		payloads = append(payloads, &usecase.FilePayload{
			File:        file,
			Size:        fileHeader.Size,
			Filename:    fileHeader.Filename,
			ContentType: h.fileConfig.GetContentTypeByExtension(fileHeader.Filename),
		})
	}

	// 5. Вызов usecase
	msg, err := h.messageUsecase.SendMessage(r.Context(), usecase.SendMessageInput{
		ThreadID: threadID,
		UserID:   userID,
		Username: username,
		Content:  content,
		Payloads: payloads,
	})
	if err != nil {
		code, clientErr := apperrors.GetErrAndCodeToSend(err)
		h.logger.Error("failed to send message", zap.Error(err))
		lib.WriteError(w, clientErr.Error(), code)
		return
	}

	// 6. Формируем DTO-ответ
	payloadLinks := make([]string, len(msg.Payloads))
	for i, p := range msg.Payloads {
		payloadLinks[i] = p.FileLink
	}

	resp := dto.SendMessageResponse{
		Message: dto.MessageResponse{
			ID:        msg.ID,
			ThreadID:  msg.ThreadID,
			Username:  username,
			Content:   msg.Content,
			Payloads:  payloadLinks,
			CreatedAt: msg.CreatedAt,
			UpdatedAt: msg.UpdatedAt,
		},
	}

	// 7. Отправляем JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Warn("failed to encode send message response", zap.Error(err))
	}
}
