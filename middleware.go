package goweb

import (
	"context"
	"net/http"
	"time"
)

// Middleware is a function that processes requests before they reach the handler
type Middleware func(HandlerFunc) HandlerFunc

// HandlerFunc is a function that processes a request with a context
type HandlerFunc func(*Context)

// LoggerMiddleware logs all requests
func LoggerMiddleware() Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(c *Context) {
			start := time.Now()

			c.LogRequest()
			next(c)

			c.Logger().Debug("Completed in %v", time.Since(start))
		}
	}
}

// RecoverMiddleware recovers from panics
func RecoverMiddleware() Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(c *Context) {
			defer func() {
				if r := recover(); r != nil {
					c.Logger().Error("Panic recovered: %v", r)
					c.Error(http.StatusInternalServerError, "Internal Server Error")
				}
			}()
			next(c)
		}
	}
}

// TimeoutMiddleware adds a timeout to the request
func TimeoutMiddleware(timeout time.Duration) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(c *Context) {
			ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
			defer cancel()

			c.Request = c.Request.WithContext(ctx)

			done := make(chan struct{})
			go func() {
				next(c)
				close(done)
			}()

			select {
			case <-done:
				return
			case <-ctx.Done():
				if ctx.Err() == context.DeadlineExceeded {
					c.Error(http.StatusRequestTimeout, "Request timeout")
				}
				return
			}
		}
	}
}
