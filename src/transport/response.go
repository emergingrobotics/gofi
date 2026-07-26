package transport

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Response represents an HTTP response.
type Response struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
}

// IsSuccess returns true if the response indicates success (2xx status code).
func (r *Response) IsSuccess() bool {
	return r.StatusCode >= 200 && r.StatusCode < 300
}

// Parse parses the response body into the given value.
func (r *Response) Parse(v interface{}) error {
	if len(r.Body) == 0 {
		return nil
	}

	if err := json.Unmarshal(r.Body, v); err != nil {
		return fmt.Errorf("failed to parse response body: %w", err)
	}

	return nil
}

// String returns the response body as a string.
func (r *Response) String() string {
	return string(r.Body)
}
