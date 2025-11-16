package deliveryHTTP

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/onionfriend2004/threadbook_backend/internal/apperrors"
	"github.com/onionfriend2004/threadbook_backend/internal/lib"
	"github.com/onionfriend2004/threadbook_backend/internal/lib/middleware/auth"
	"github.com/onionfriend2004/threadbook_backend/internal/lib/middleware/validator"
	"github.com/onionfriend2004/threadbook_backend/internal/spool/delivery/dto"
	"github.com/onionfriend2004/threadbook_backend/internal/spool/usecase"
	"go.uber.org/zap"
)

func (h *SpoolHandler) AccessLevel(w http.ResponseWriter, r *http.Request) {
	editorID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		lib.WriteError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	spoolIDStr := chi.URLParam(r, "spoolID")
	spoolIDInt, err := strconv.Atoi(spoolIDStr)
	if err != nil || spoolIDInt <= 0 {
		lib.WriteError(w, "invalid spool_id", http.StatusBadRequest)
		return
	}
	spoolID := uint(spoolIDInt)

	username := chi.URLParam(r, "username")
	req := validator.GetValidatedBody[dto.AccessLevelRequest](r)
	err = h.usecase.AccessLevel(r.Context(), usecase.AccessLevelInput{
		EditorID:    editorID,
		SpoolID:     spoolID,
		Username:    username,
		AccessLevel: req.AccessLevel,
	})

	if err != nil {
		code, clientErr := apperrors.GetErrAndCodeToSend(err)
		h.logger.Warn("failed change user Access Level", zap.Error(err))
		lib.WriteError(w, clientErr.Error(), code)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(lib.StatusOK)
}
