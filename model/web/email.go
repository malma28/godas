package web

type EmailVerificationCreateRequest struct {
	Email string `json:"email" validate:"required,email"`
	Code  string `json:"code" validate:"required,min=6,max=6"`
}

type EmailVerificationRecreateRequest struct {
	Email string `json:"email" validate:"required,email"`
}
