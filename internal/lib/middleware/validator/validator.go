package validator

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// var validate = validator.New()

type ValidationErrorResponse struct {
	Error string `json:"error"`
}

func ValidateJSONMiddleware(prototype any) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			obj := reflect.New(reflect.TypeOf(prototype)).Interface()

			// Декодируем JSON
			if err := json.NewDecoder(r.Body).Decode(obj); err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(ValidationErrorResponse{
					Error: "invalid JSON format",
				})
				return
			}

			if err := validate.Struct(obj); err != nil {
				var validationErrors []string

				if ve, ok := err.(validator.ValidationErrors); ok {
					for _, e := range ve {
						msg := getValidationMessage(e)
						validationErrors = append(validationErrors, msg)
					}
				} else {
					validationErrors = append(validationErrors, "validation failed")
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(ValidationErrorResponse{
					Error: strings.Join(validationErrors, "; "),
				})
				return
			}

			// Сохраняем в контекст
			ctx := context.WithValue(r.Context(), "validatedBody", obj)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Функция для получения понятных сообщений об ошибках
func getValidationMessage(e validator.FieldError) string {
	field := strings.ToLower(e.Field())

	switch e.Tag() {
	case "required":
		return field + " is required"
	case "email":
		return field + " must be a valid email"
	case "password":
		return field + " must contain uppercase, lowercase, numbers and be at least 8 characters long"
	case "username":
		return field + " can only contain letters, numbers, underscore and hyphen"
	case "min":
		return field + " must be at least " + e.Param() + " characters"
	case "max":
		return field + " must be at most " + e.Param() + " characters"
	default:
		return field + " failed " + e.Tag() + " validation"
	}
}

func GetValidatedBody[T any](r *http.Request) *T {
	if body := r.Context().Value("validatedBody"); body != nil {
		return body.(*T)
	}
	return nil
}
