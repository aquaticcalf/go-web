package goweb

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

// App is the main web application structure
type App struct {
	router       *mux.Router
	middlewares  []Middleware
	logger       Logger
	errorHandler ErrorHandler
}

// AppConfig holds configuration for the App
type AppConfig struct {
	CorsOptions  cors.Options
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	Logger       Logger
	ErrorHandler ErrorHandler
}

// DefaultAppConfig returns a default configuration
func DefaultAppConfig() AppConfig {
	return AppConfig{
		CorsOptions: cors.Options{
			AllowedOrigins:   []string{"*"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
			AllowedHeaders:   []string{"Content-Type", "Authorization", "X-Requested-With"},
			AllowCredentials: true,
		},
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		Logger:       &defaultLogger{},
		ErrorHandler: &defaultErrorHandler{},
	}
}

// New creates a new App instance with default configuration
func New() *App {
	return NewWithConfig(DefaultAppConfig())
}

// NewWithConfig creates a new App with custom configuration
func NewWithConfig(config AppConfig) *App {
	app := &App{
		router:       mux.NewRouter().StrictSlash(true),
		middlewares:  []Middleware{},
		logger:       config.Logger,
		errorHandler: config.ErrorHandler,
	}
	return app
}

// Use adds middleware to the application
func (a *App) Use(middleware ...Middleware) {
	a.middlewares = append(a.middlewares, middleware...)
}

// UseFunc adds handler function as middleware
func (a *App) UseFunc(handler func(next http.Handler) http.Handler) {
	a.router.Use(handler)
}

// SetLogger sets a custom logger
func (a *App) SetLogger(logger Logger) {
	a.logger = logger
}

// SetErrorHandler sets a custom error handler
func (a *App) SetErrorHandler(handler ErrorHandler) {
	a.errorHandler = handler
}

// Group creates a new route group
func (a *App) Group(path string) *Group {
	return &Group{
		router:      a.router.PathPrefix(path).Subrouter(),
		parent:      a,
		middlewares: []Middleware{},
	}
}

// Add registers routes for a group
func (a *App) Add(path string, setup func(*Group)) {
	group := a.Group(path)
	setup(group)
}

// Static serves static files from the given directory
func (a *App) Static(path, dir string) {
	fileServer := http.FileServer(http.Dir(dir))
	a.router.PathPrefix(path).Handler(http.StripPrefix(path, fileServer))
}

// Routes returns all registered routes for debugging
func (a *App) Routes() []string {
	var routes []string
	a.router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		path, _ := route.GetPathTemplate()
		methods, _ := route.GetMethods()
		routes = append(routes, fmt.Sprintf("%s [%s]", path, strings.Join(methods, ",")))
		return nil
	})
	return routes
}

// Run starts the HTTP server
func (a *App) Run(address string) error {
	config := DefaultAppConfig()
	return a.RunWithConfig(address, config)
}

// RunWithConfig starts the HTTP server with the given configuration
func (a *App) RunWithConfig(address string, config AppConfig) error {
	corsHandler := cors.New(config.CorsOptions)
	handler := corsHandler.Handler(a.router)

	server := &http.Server{
		Addr:         ":" + address,
		Handler:      handler,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
	}

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	go func() {
		a.logger.Info("Server started at http://localhost:%s", address)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Error("Error starting server: %v", err)
		}
	}()

	<-stop
	a.logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		a.logger.Error("Server shutdown error: %v", err)
		return err
	}

	a.logger.Info("Server gracefully stopped")
	return nil
}

// ServeHTTP allows the App to implement the http.Handler interface
func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.router.ServeHTTP(w, r)
}
