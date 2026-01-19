package transport

import (
	"net/http"
	"testing"
)

func TestResponse_IsSuccess(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		want       bool
	}{
		{"200 OK", 200, true},
		{"201 Created", 201, true},
		{"204 No Content", 204, true},
		{"299 (last 2xx)", 299, true},
		{"300 Redirect", 300, false},
		{"400 Bad Request", 400, false},
		{"404 Not Found", 404, false},
		{"500 Server Error", 500, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &Response{StatusCode: tt.statusCode}
			got := resp.IsSuccess()
			if got != tt.want {
				t.Errorf("IsSuccess() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResponse_Parse(t *testing.T) {
	tests := []struct {
		name    string
		body    []byte
		target  interface{}
		wantErr bool
	}{
		{
			"valid JSON",
			[]byte(`{"key": "value", "count": 42}`),
			&map[string]interface{}{},
			false,
		},
		{
			"empty body",
			[]byte{},
			&map[string]interface{}{},
			false,
		},
		{
			"invalid JSON",
			[]byte(`{invalid json}`),
			&map[string]interface{}{},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &Response{Body: tt.body}
			err := resp.Parse(tt.target)

			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && len(tt.body) > 0 {
				// Verify parsing worked
				result, ok := tt.target.(*map[string]interface{})
				if !ok {
					t.Fatal("Type assertion failed")
				}
				if len(*result) == 0 {
					t.Error("Parse() should have populated the map")
				}
			}
		})
	}
}

func TestResponse_String(t *testing.T) {
	body := []byte("test response body")
	resp := &Response{Body: body}

	got := resp.String()
	want := "test response body"

	if got != want {
		t.Errorf("String() = %s, want %s", got, want)
	}
}

func TestResponse_Headers(t *testing.T) {
	headers := http.Header{
		"Content-Type": []string{"application/json"},
		"X-Custom":     []string{"custom-value"},
	}

	resp := &Response{
		StatusCode: 200,
		Body:       []byte(`{}`),
		Headers:    headers,
	}

	if resp.Headers.Get("Content-Type") != "application/json" {
		t.Error("Content-Type header not preserved")
	}

	if resp.Headers.Get("X-Custom") != "custom-value" {
		t.Error("X-Custom header not preserved")
	}
}
