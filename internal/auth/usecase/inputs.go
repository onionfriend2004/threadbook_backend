package usecase

// SignUpInput contains the required data to register a new user.
type SignUpInput struct {
	Email    string
	Username string
	Password string
}

// SignInInput contains the credentials for user authentication.
type SignInInput struct {
	Email    string
	Password string
}
