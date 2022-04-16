package web

import "net/http"

type Payload struct {
	Code    int    `json:"code"`
	Status  string `json:"status"`
	Success bool   `json:"success"`
	Data    any    `json:"data"`
}

func NewFailPayload(statusCode int) Payload {
	return Payload{
		Code:    statusCode,
		Status:  http.StatusText(statusCode),
		Success: false,
		Data:    nil,
	}
}
