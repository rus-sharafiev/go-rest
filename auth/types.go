package auth

import "time"

// -- User ------------------------------------------------------------------------
type UserData struct {
	ID        *int       `json:"id"`
	Email     *string    `json:"email"`
	FirstName *string    `json:"firstName"`
	LastName  *string    `json:"lastName"`
	Phone     *string    `json:"phone"`
	Avatar    *string    `json:"avatar"`
	Access    *string    `json:"access"`
	CreatedAt *time.Time `json:"createdAt"`
	UpdatedAt *time.Time `json:"updatedAt"`
}

// -- Sign Up ---------------------------------------------------------------------

type SignUpDto struct {
	Email      string  `json:"email"`
	Password   string  `json:"password"`
	Grecaptcha *string `json:"grecaptcha"`
}

type SignUpCode struct {
	Code int `json:"code"`
}

type SignUpData struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Code     int    `json:"code"`
}

type SignUpResult struct {
	User        UserData `json:"user"`
	AccessToken string   `json:"accessToken"`
}

type ReCaptchaResponse struct {
	Success     bool         `json:"success"`
	ChallengeTs string       `json:"challenge_ts"`
	Hostname    string       `json:"hostname"`
	ErrorCodes  *interface{} `json:"error-codes"`
}

// -- Log In ----------------------------------------------------------------------

type LogInDto struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserPswd struct {
	PasswordHash string
	UserId       int
}

type LoginResult struct {
	User        UserData `json:"user"`
	AccessToken string   `json:"accessToken"`
}

// -- Refresh access token --------------------------------------------------------

type RefreshResult struct {
	AccessToken string `json:"accessToken"`
}

// -- Reset password --------------------------------------------------------------

type ResetPasswordDto struct {
	Email string `json:"email"`
}

// -- Update password -------------------------------------------------------------

type UpdatePasswordDto struct {
	Password string `json:"password"`
	Token    string `json:"token"`
}

// Status message of a sent mail
type Message struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
}
