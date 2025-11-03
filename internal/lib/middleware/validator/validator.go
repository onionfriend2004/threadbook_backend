package validator

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func ValidateJSONMiddleware(prototype any) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			obj := reflect.New(reflect.TypeOf(prototype)).Interface()

			if err := json.NewDecoder(r.Body).Decode(obj); err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": "invalid JSON"})
				return
			}

			if err := validate.Struct(obj); err != nil {
				var msgs []string
				for _, e := range err.(validator.ValidationErrors) {
					msgs = append(msgs, strings.ToLower(e.Field())+": failed "+e.Tag())
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": strings.Join(msgs, ", ")})
				return
			}

			ctx := context.WithValue(r.Context(), "validatedBody", obj)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetValidatedBody[T any](r *http.Request) *T {
	return r.Context().Value("validatedBody").(*T)
}
