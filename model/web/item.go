package web

type ItemRequest struct {
	Name string `json:"name" validate:"required,min=1,max=128"`
}

type ItemResponse struct {
	Index uint64 `json:"index"`
	Name  string `json:"name"`
}
