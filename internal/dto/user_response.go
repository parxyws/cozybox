package dto

type UserResponse struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

type RegisterResponse struct {
	ReferenceId string `json:"reference_id"`
}

type VerifyEmailResponse struct {
	ReferenceId string `json:"reference_id"`
}

type UserAutheticateResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	TokenType    string       `json:"token_type"`
	User         UserResponse `json:"user"`
}

type UpdateEmailResponse struct {
	ReferenceId string `json:"reference_id"`
}

type CommitUpdateResponse struct {
}
