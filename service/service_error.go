package service

import "errors"

var ErrBadRequest = errors.New("bad request")
var ErrDuplicate = errors.New("duplicate")
var ErrNotFound = errors.New("not found")
var ErrUnauthorized = errors.New("unauthorized")
