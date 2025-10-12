package auth

import (
	"context"
	"net/http"

	"github.com/onionfriend2004/threadbook_backend/internal/lib"
)

func AuthMiddleware(authenticator AuthenticatorInterface) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("sid")
			if err != nil || cookie.Value == "" {
				lib.WriteError(w, "unauthorized: missing sid cookie", http.StatusUnauthorized)
				return
			}

			userID, username, err := authenticator.Authenticate(cookie.Value)
			if err != nil {
				lib.WriteError(w, err.Error(), http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			ctx = context.WithValue(ctx, UsernameKey, username)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
