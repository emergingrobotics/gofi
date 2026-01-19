package internal

import (
	"encoding/json"
	"testing"
)

func TestParseAPIResponse_Success(t *testing.T) {
	input := `{
		"meta": {
			"rc": "ok",
			"count": 2
		},
		"data": [
			{"id": "1", "name": "test1"},
			{"id": "2", "name": "test2"}
		]
	}`

	type TestData struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	resp, err := ParseAPIResponse[TestData]([]byte(input))
	if err != nil {
		t.Fatalf("ParseAPIResponse() error = %v", err)
	}

	if resp.Meta.RC != "ok" {
		t.Errorf("Meta.RC = %s, want ok", resp.Meta.RC)
	}

	if len(resp.Data) != 2 {
		t.Fatalf("len(Data) = %d, want 2", len(resp.Data))
	}

	if resp.Data[0].Name != "test1" {
		t.Errorf("Data[0].Name = %s, want test1", resp.Data[0].Name)
	}
}

func TestParseAPIResponse_Error(t *testing.T) {
	input := `{
		"meta": {
			"rc": "error",
			"msg": "Permission denied"
		},
		"data": []
	}`

	type TestData struct {
		ID string `json:"id"`
	}

	_, err := ParseAPIResponse[TestData]([]byte(input))
	if err == nil {
		t.Fatal("ParseAPIResponse() expected error, got nil")
	}

	expectedMsg := "API error: Permission denied (rc=error)"
	if err.Error() != expectedMsg {
		t.Errorf("error message = %q, want %q", err.Error(), expectedMsg)
	}
}

func TestParseAPIResponse_MalformedJSON(t *testing.T) {
	input := `{invalid json`

	type TestData struct {
		ID string `json:"id"`
	}

	_, err := ParseAPIResponse[TestData]([]byte(input))
	if err == nil {
		t.Fatal("ParseAPIResponse() expected error for malformed JSON, got nil")
	}
}

func TestIsErrorResponse(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			"success response",
			`{"meta": {"rc": "ok"}, "data": []}`,
			false,
		},
		{
			"error response",
			`{"meta": {"rc": "error", "msg": "Failed"}, "data": []}`,
			true,
		},
		{
			"no meta",
			`{"data": []}`,
			false,
		},
		{
			"empty rc is ok",
			`{"meta": {"rc": ""}, "data": []}`,
			false,
		},
		{
			"malformed json",
			`{invalid}`,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsErrorResponse([]byte(tt.input))
			if got != tt.want {
				t.Errorf("IsErrorResponse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractErrorMessage(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			"with message",
			`{"meta": {"rc": "error", "msg": "Permission denied"}}`,
			"Permission denied",
		},
		{
			"without message",
			`{"meta": {"rc": "error"}}`,
			"error (rc=error)",
		},
		{
			"malformed json",
			`{invalid}`,
			"unknown error",
		},
		{
			"no meta",
			`{"data": []}`,
			"error (rc=)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractErrorMessage([]byte(tt.input))
			if got != tt.want {
				t.Errorf("ExtractErrorMessage() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseSingleResult(t *testing.T) {
	input := `{
		"meta": {"rc": "ok"},
		"data": [
			{"id": "1", "name": "test1"},
			{"id": "2", "name": "test2"}
		]
	}`

	type TestData struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	result, err := ParseSingleResult[TestData]([]byte(input))
	if err != nil {
		t.Fatalf("ParseSingleResult() error = %v", err)
	}

	if result.ID != "1" {
		t.Errorf("ID = %s, want 1", result.ID)
	}

	if result.Name != "test1" {
		t.Errorf("Name = %s, want test1", result.Name)
	}
}

func TestParseSingleResult_Empty(t *testing.T) {
	input := `{
		"meta": {"rc": "ok"},
		"data": []
	}`

	type TestData struct {
		ID string `json:"id"`
	}

	_, err := ParseSingleResult[TestData]([]byte(input))
	if err == nil {
		t.Fatal("ParseSingleResult() expected error for empty data, got nil")
	}
}

func TestMarshalCommand(t *testing.T) {
	data, err := MarshalCommand("restart", "aa:bb:cc:dd:ee:ff")
	if err != nil {
		t.Fatalf("MarshalCommand() error = %v", err)
	}

	// Parse it back to verify
	var cmd struct {
		Cmd string `json:"cmd"`
		MAC string `json:"mac"`
	}

	if err := ParseJSON(data, &cmd); err != nil {
		t.Fatalf("Failed to parse marshaled command: %v", err)
	}

	if cmd.Cmd != "restart" {
		t.Errorf("Cmd = %s, want restart", cmd.Cmd)
	}

	if cmd.MAC != "aa:bb:cc:dd:ee:ff" {
		t.Errorf("MAC = %s, want aa:bb:cc:dd:ee:ff", cmd.MAC)
	}
}

// Helper function for tests
func ParseJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
