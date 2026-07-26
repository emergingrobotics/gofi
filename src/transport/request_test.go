package transport

import (
	"testing"
)

func TestNewRequest(t *testing.T) {
	req := NewRequest("GET", "/api/test")

	if req.Method != "GET" {
		t.Errorf("Method = %s, want GET", req.Method)
	}

	if req.Path != "/api/test" {
		t.Errorf("Path = %s, want /api/test", req.Path)
	}

	if req.Headers == nil {
		t.Error("Headers should be initialized")
	}
}

func TestRequest_WithBody(t *testing.T) {
	body := map[string]string{"key": "value"}
	req := NewRequest("POST", "/api/test").WithBody(body)

	if req.Body == nil {
		t.Error("Body should be set")
	}

	bodyMap, ok := req.Body.(map[string]string)
	if !ok {
		t.Fatal("Body type assertion failed")
	}

	if bodyMap["key"] != "value" {
		t.Errorf("Body[key] = %s, want value", bodyMap["key"])
	}
}

func TestRequest_WithHeader(t *testing.T) {
	req := NewRequest("GET", "/api/test").
		WithHeader("X-Test", "test-value")

	if req.Headers["X-Test"] != "test-value" {
		t.Errorf("Headers[X-Test] = %s, want test-value", req.Headers["X-Test"])
	}
}

func TestRequest_WithHeaders(t *testing.T) {
	headers := map[string]string{
		"X-Header-1": "value1",
		"X-Header-2": "value2",
	}

	req := NewRequest("GET", "/api/test").WithHeaders(headers)

	for k, v := range headers {
		if req.Headers[k] != v {
			t.Errorf("Headers[%s] = %s, want %s", k, req.Headers[k], v)
		}
	}
}

func TestRequest_Chaining(t *testing.T) {
	body := map[string]int{"count": 42}
	req := NewRequest("POST", "/api/test").
		WithBody(body).
		WithHeader("Content-Type", "application/json").
		WithHeader("X-Custom", "custom-value")

	if req.Body == nil {
		t.Error("Body should be set")
	}

	if req.Headers["Content-Type"] != "application/json" {
		t.Error("Content-Type header not set")
	}

	if req.Headers["X-Custom"] != "custom-value" {
		t.Error("X-Custom header not set")
	}
}
