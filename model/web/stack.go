package web

import "godas/model/domain"

type StackCreateRequest struct {
}

type StackResponse struct {
	ID    string        `json:"id"`
	Owner string        `json:"owner"`
	Items []domain.Item `json:"items"`
}
