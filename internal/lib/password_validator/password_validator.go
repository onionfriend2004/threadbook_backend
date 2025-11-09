package passwordvalidator

import (
	"fmt"
	"regexp"
)

func ValidatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	// Хотя бы одна заглавная, одна строчная, одна цифра, один спецсимвол
	patterns := []string{
		`[A-Z]`, // заглавная
		`[a-z]`, // строчная
		`[0-9]`, // цифра
	}

	for _, pattern := range patterns {
		matched, _ := regexp.MatchString(pattern, password)
		if !matched {
			return fmt.Errorf("password must contain at least one uppercase letter, one lowercase letter, one number and one special character")
		}
	}

	return nil
}
