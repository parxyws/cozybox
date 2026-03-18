package dto

type RegisterUserRequest struct {
	Name     string `json:"name"     validate:"required,min=2,max=100"`
	Username string `json:"username" validate:"required,min=3,max=30,alphanum"`
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=128"`
}

type VerifyEmailRequest struct {
	ReferenceId string `json:"reference_id" validate:"required"`
	Otp         string `json:"otp"          validate:"required,len=6,numeric"`
}

type UpdateUserRequest struct {
	Name     string `json:"name"     validate:"omitempty,min=2,max=100"`
	Username string `json:"username" validate:"omitempty,min=3,max=30,alphanum"`
	Email    string `json:"email"    validate:"omitempty,email"`
	Password string `json:"password" validate:"omitempty,min=8,max=128"`
}

type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type RefreshTokenRequest struct {
	SessionID    string `json:"session_id"    validate:"required"`
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type UpdateProfileRequest struct {
	Name     string `json:"name"     validate:"omitempty,min=2,max=100"`
	Username string `json:"username" validate:"omitempty,min=3,max=30,alphanum"`
}

type UpdatePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required,min=8,max=128"`
	NewPassword     string `json:"new_password"     validate:"required,min=8,max=128"`
}

type UpdateEmailRequest struct {
	CurrentPassword string `json:"current_password" validate:"required,min=8,max=128"`
	NewEmail        string `json:"new_email"        validate:"required,email"`
}

type DeleteAccountRequest struct {
	CurrentPassword string `json:"current_password" validate:"required,min=8,max=128"`
}
