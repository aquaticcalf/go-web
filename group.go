package goweb

import (
	"net/http"

	"github.com/gorilla/mux"
)

// Group represents a group of routes
type Group struct {
	router      *mux.Router
	parent      *App
	middlewares []Middleware
}

// Use adds middleware to the group
func (g *Group) Use(middleware ...Middleware) {
	g.middlewares = append(g.middlewares, middleware...)
}

// Route registers a route with the given HTTP methods
func (g *Group) Route(path string, action HandlerFunc, methods ...string) {
	handler := g.wrapHandler(action)

	if len(methods) == 0 {
		g.router.HandleFunc(path, handler)
	} else {
		g.router.HandleFunc(path, handler).Methods(methods...)
	}
}

// GET registers a route with GET method
func (g *Group) GET(path string, action HandlerFunc) {
	g.Route(path, action, http.MethodGet)
}

// POST registers a route with POST method
func (g *Group) POST(path string, action HandlerFunc) {
	g.Route(path, action, http.MethodPost)
}

// PUT registers a route with PUT method
func (g *Group) PUT(path string, action HandlerFunc) {
	g.Route(path, action, http.MethodPut)
}

// DELETE registers a route with DELETE method
func (g *Group) DELETE(path string, action HandlerFunc) {
	g.Route(path, action, http.MethodDelete)
}

// PATCH registers a route with PATCH method
func (g *Group) PATCH(path string, action HandlerFunc) {
	g.Route(path, action, http.MethodPatch)
}

// OPTIONS registers a route with OPTIONS method
func (g *Group) OPTIONS(path string, action HandlerFunc) {
	g.Route(path, action, http.MethodOptions)
}

// wrapHandler applies all middleware to the handler
func (g *Group) wrapHandler(handler HandlerFunc) http.HandlerFunc {
	// Combine app middlewares and group middlewares
	middlewares := append(g.parent.middlewares, g.middlewares...)

	h := handler
	// Apply middlewares in reverse order
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := &Context{
			Writer:  w,
			Request: r,
			Params:  mux.Vars(r),
			app:     g.parent,
			store:   make(map[string]interface{}),
		}

		defer func() {
			if r := recover(); r != nil {
				g.parent.logger.Error("Panic recovered: %v", r)
				ctx.Error(http.StatusInternalServerError, "Internal Server Error")
			}
		}()

		h(ctx)
	}
}
