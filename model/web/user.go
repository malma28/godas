package web

type UserCreateRequest struct {
	Name     string `json:"name" validate:"required,min=1,max=128"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type UserUpdateRequest struct {
	Name string `json:"name" validate:"min=1,max=128"`
}

type UserResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
