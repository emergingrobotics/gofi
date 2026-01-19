package mock

import (
	"encoding/json"
	"net/http"

	"github.com/unifi-go/gofi/types"
)

// writeJSON writes a JSON response.
func writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// writeAPIResponse writes a successful API response with data.
func writeAPIResponse(w http.ResponseWriter, data interface{}) {
	resp := types.APIResponse[interface{}]{
		Meta: types.ResponseMeta{
			RC: "ok",
		},
		Data: []interface{}{data},
	}

	// If data is already a slice, use it directly
	switch v := data.(type) {
	case []interface{}:
		resp.Data = v
	case []types.Site:
		resp.Data = make([]interface{}, len(v))
		for i, item := range v {
			resp.Data[i] = item
		}
	case []types.Device:
		resp.Data = make([]interface{}, len(v))
		for i, item := range v {
			resp.Data[i] = item
		}
	case []types.Network:
		resp.Data = make([]interface{}, len(v))
		for i, item := range v {
			resp.Data[i] = item
		}
	default:
		// Single item, wrap in slice
		resp.Data = []interface{}{data}
	}

	resp.Meta.Count = len(resp.Data)
	writeJSON(w, http.StatusOK, resp)
}

// writeAPIError writes an API error response.
func writeAPIError(w http.ResponseWriter, statusCode int, rc, message string) {
	resp := struct {
		Meta types.ResponseMeta `json:"meta"`
		Data []interface{}      `json:"data"`
	}{
		Meta: types.ResponseMeta{
			RC:      rc,
			Message: message,
		},
		Data: []interface{}{},
	}

	writeJSON(w, statusCode, resp)
}

// writeUnauthorized writes a 401 Unauthorized response.
func writeUnauthorized(w http.ResponseWriter) {
	writeAPIError(w, http.StatusUnauthorized, "error", "unauthorized")
}

// writeForbidden writes a 403 Forbidden response.
func writeForbidden(w http.ResponseWriter, message string) {
	writeAPIError(w, http.StatusForbidden, "error", message)
}

// writeNotFound writes a 404 Not Found response.
func writeNotFound(w http.ResponseWriter) {
	writeAPIError(w, http.StatusNotFound, "error", "not found")
}

// writeBadRequest writes a 400 Bad Request response.
func writeBadRequest(w http.ResponseWriter, message string) {
	writeAPIError(w, http.StatusBadRequest, "error", message)
}
