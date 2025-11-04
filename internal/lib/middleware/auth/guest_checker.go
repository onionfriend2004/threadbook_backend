package auth

import (
	"net/http"

	"github.com/onionfriend2004/threadbook_backend/internal/lib"
)

func GuestMiddleware(authenticator AuthenticatorInterface) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("sid")
			if err != nil || cookie.Value == "" {
				lib.WriteError(w, "unauthorized: missing sid cookie", http.StatusUnauthorized)
				return
			}

			userID, _, err := authenticator.Authenticate(cookie.Value)
			if err != nil {
				lib.WriteError(w, err.Error(), http.StatusUnauthorized)
				return
			}

			user, err := authenticator.GetUserByID(userID)
			if err != nil {
				lib.WriteError(w, "user not found", http.StatusUnauthorized)
				return
			}

			if user.IsGuest {
				lib.WriteError(w, "this endpoint is not for guests", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
