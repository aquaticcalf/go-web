package goweb

import "net/http"

// ErrorHandler is responsible for handling errors
type ErrorHandler interface {
	Handle(ctx *Context, err error)
}

// defaultErrorHandler implements the ErrorHandler interface
type defaultErrorHandler struct{}

func (eh *defaultErrorHandler) Handle(ctx *Context, err error) {
	ctx.Error(http.StatusInternalServerError, "%v", err)
}
