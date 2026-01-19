package internal

import (
	"encoding/json"
	"fmt"

	"github.com/unifi-go/gofi/types"
)

// ParseAPIResponse parses a UniFi API response into the generic APIResponse type.
func ParseAPIResponse[T any](data []byte) (*types.APIResponse[T], error) {
	var resp types.APIResponse[T]
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	// Check if the response indicates an error
	if resp.Meta.RC != "ok" && resp.Meta.RC != "" {
		return nil, fmt.Errorf("API error: %s (rc=%s)", resp.Meta.Message, resp.Meta.RC)
	}

	return &resp, nil
}

// IsErrorResponse checks if the response data indicates an error.
func IsErrorResponse(data []byte) bool {
	var meta struct {
		Meta types.ResponseMeta `json:"meta"`
	}

	if err := json.Unmarshal(data, &meta); err != nil {
		return false
	}

	return meta.Meta.RC != "" && meta.Meta.RC != "ok"
}

// ExtractErrorMessage extracts the error message from an API error response.
func ExtractErrorMessage(data []byte) string {
	var meta struct {
		Meta types.ResponseMeta `json:"meta"`
	}

	if err := json.Unmarshal(data, &meta); err != nil {
		return "unknown error"
	}

	if meta.Meta.Message != "" {
		return meta.Meta.Message
	}

	return fmt.Sprintf("error (rc=%s)", meta.Meta.RC)
}

// ParseSingleResult parses an API response that contains a single item.
// If the response contains multiple items, returns the first one.
func ParseSingleResult[T any](data []byte) (*T, error) {
	resp, err := ParseAPIResponse[T](data)
	if err != nil {
		return nil, err
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no data in response")
	}

	return &resp.Data[0], nil
}

// MarshalCommand marshals a command request with the given command name and MAC.
func MarshalCommand(cmd, mac string) ([]byte, error) {
	cmdReq := types.CommandRequest{
		Cmd: cmd,
		MAC: mac,
	}
	return json.Marshal(cmdReq)
}
