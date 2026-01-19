package mock

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/unifi-go/gofi/types"
)

func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"key": "value"}

	writeJSON(w, http.StatusOK, data)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}

	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %s, want application/json", ct)
	}

	var result map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if result["key"] != "value" {
		t.Errorf("key = %s, want value", result["key"])
	}
}

func TestWriteAPIResponse(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"id": "123", "name": "test"}

	writeAPIResponse(w, data)

	var resp types.APIResponse[map[string]interface{}]
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.Meta.RC != "ok" {
		t.Errorf("Meta.RC = %s, want ok", resp.Meta.RC)
	}

	if len(resp.Data) != 1 {
		t.Fatalf("len(Data) = %d, want 1", len(resp.Data))
	}
}

func TestWriteAPIError(t *testing.T) {
	w := httptest.NewRecorder()

	writeAPIError(w, http.StatusNotFound, "error", "Resource not found")

	if w.Code != http.StatusNotFound {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusNotFound)
	}

	var resp struct {
		Meta types.ResponseMeta `json:"meta"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.Meta.RC != "error" {
		t.Errorf("Meta.RC = %s, want error", resp.Meta.RC)
	}

	if resp.Meta.Message != "Resource not found" {
		t.Errorf("Meta.Message = %s, want 'Resource not found'", resp.Meta.Message)
	}
}

func TestWriteUnauthorized(t *testing.T) {
	w := httptest.NewRecorder()
	writeUnauthorized(w)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestWriteForbidden(t *testing.T) {
	w := httptest.NewRecorder()
	writeForbidden(w, "CSRF token invalid")

	if w.Code != http.StatusForbidden {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusForbidden)
	}

	var resp struct {
		Meta types.ResponseMeta `json:"meta"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.Meta.Message != "CSRF token invalid" {
		t.Errorf("Message = %s, want 'CSRF token invalid'", resp.Meta.Message)
	}
}

func TestWriteNotFound(t *testing.T) {
	w := httptest.NewRecorder()
	writeNotFound(w)

	if w.Code != http.StatusNotFound {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestWriteBadRequest(t *testing.T) {
	w := httptest.NewRecorder()
	writeBadRequest(w, "Invalid input")

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}
