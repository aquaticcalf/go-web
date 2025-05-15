package goweb

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Context represents the request context
type Context struct {
	Writer  http.ResponseWriter
	Request *http.Request
	Params  map[string]string
	app     *App
	store   map[string]interface{}
	aborted bool
}

// Set stores a value in the context
func (c *Context) Set(key string, value interface{}) {
	c.store[key] = value
}

// Get retrieves a value from the context
func (c *Context) Get(key string) (interface{}, bool) {
	val, exists := c.store[key]
	return val, exists
}

// MustGet retrieves a value from the context and panics if not found
func (c *Context) MustGet(key string) interface{} {
	if val, exists := c.Get(key); exists {
		return val
	}
	panic(fmt.Sprintf("Key %s does not exist in context", key))
}

// GetParam gets a URL parameter by name
func (c *Context) GetParam(name string) string {
	return c.Params[name]
}

// Query gets a query parameter by name
func (c *Context) Query(name string) string {
	return c.Request.URL.Query().Get(name)
}

// QueryDefault gets a query parameter by name with a default value
func (c *Context) QueryDefault(name, defaultValue string) string {
	if value := c.Query(name); value != "" {
		return value
	}
	return defaultValue
}

// Abort stops the chain execution
func (c *Context) Abort() {
	c.aborted = true
}

// IsAborted returns whether the context was aborted
func (c *Context) IsAborted() bool {
	return c.aborted
}

// Bind binds request body to a struct
func (c *Context) Bind(obj interface{}) error {
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(obj); err != nil {
		return err
	}
	return nil
}

// BindAndValidate binds and validates request body
func (c *Context) BindAndValidate(obj interface{}) error {
	if err := c.Bind(obj); err != nil {
		return err
	}

	// Add validation logic here if needed
	return nil
}

// JSON sends a JSON response
func (c *Context) JSON(status int, data interface{}) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(status)

	if err := json.NewEncoder(c.Writer).Encode(data); err != nil {
		c.app.logger.Error("Error encoding JSON: %v", err)
	}
}

// Success sends a successful JSON response
func (c *Context) Success(data interface{}) {
	response := Response{
		Success: true,
		Data:    data,
	}
	c.JSON(http.StatusOK, response)
}

// Error sends an error JSON response
func (c *Context) Error(status int, message string, args ...interface{}) {
	formattedMessage := fmt.Sprintf(message, args...)

	response := Response{
		Success: false,
		Error: &APIError{
			Code:    status,
			Message: formattedMessage,
		},
	}

	c.JSON(status, response)
}

// HandleError handles an error with the registered error handler
func (c *Context) HandleError(err error) {
	c.app.errorHandler.Handle(c, err)
}

// Logger returns the application logger
func (c *Context) Logger() Logger {
	return c.app.logger
}

// Redirect performs an HTTP redirect
func (c *Context) Redirect(status int, location string) {
	http.Redirect(c.Writer, c.Request, location, status)
}

// Cookie gets a cookie by name
func (c *Context) Cookie(name string) (*http.Cookie, error) {
	return c.Request.Cookie(name)
}

// SetCookie sets a cookie
func (c *Context) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.Writer, cookie)
}

// File sends a file as the response
func (c *Context) File(filepath string) {
	http.ServeFile(c.Writer, c.Request, filepath)
}

// String sends a string response
func (c *Context) String(status int, format string, values ...interface{}) {
	c.Writer.Header().Set("Content-Type", "text/plain")
	c.Writer.WriteHeader(status)
	fmt.Fprintf(c.Writer, format, values...)
}

// HTML sends an HTML response
func (c *Context) HTML(status int, html string) {
	c.Writer.Header().Set("Content-Type", "text/html")
	c.Writer.WriteHeader(status)
	c.Writer.Write([]byte(html))
}

// LogRequest logs information about the current request
func (c *Context) LogRequest() {
	c.app.logger.Info("%s %s", c.Request.Method, c.Request.URL.Path)
}

// WithMeta adds metadata to a success response
func (c *Context) WithMeta(data interface{}, meta interface{}) {
	response := Response{
		Success: true,
		Data:    data,
		Meta:    meta,
	}
	c.JSON(http.StatusOK, response)
}
