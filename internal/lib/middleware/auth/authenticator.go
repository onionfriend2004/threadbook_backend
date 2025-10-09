package auth

import (
	"context"
	"net/http"

	"github.com/onionfriend2004/threadbook_backend/internal/lib"
)

func AuthMiddleware(authenticator AuthenticatorInterface) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("session_id")
			if err != nil || cookie.Value == "" {
				lib.WriteError(w, "unauthorized: missing session_id cookie", http.StatusUnauthorized)
				return
			}

			userID, err := authenticator.Authenticate(cookie.Value)
			if err != nil {
				lib.WriteError(w, err.Error(), http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
