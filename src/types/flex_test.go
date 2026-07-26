package types

import (
	"encoding/json"
	"testing"
)

func TestFlexInt_UnmarshalJSON_Numeric(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantVal  float64
		wantInt  int
		wantInt64 int64
	}{
		{"integer", `123`, 123, 123, 123},
		{"float", `123.45`, 123.45, 123, 123},
		{"zero", `0`, 0, 0, 0},
		{"negative", `-42`, -42, -42, -42},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var f FlexInt
			if err := json.Unmarshal([]byte(tt.input), &f); err != nil {
				t.Fatalf("UnmarshalJSON() error = %v", err)
			}
			if f.Val != tt.wantVal {
				t.Errorf("Val = %v, want %v", f.Val, tt.wantVal)
			}
			if f.Int() != tt.wantInt {
				t.Errorf("Int() = %v, want %v", f.Int(), tt.wantInt)
			}
			if f.Int64() != tt.wantInt64 {
				t.Errorf("Int64() = %v, want %v", f.Int64(), tt.wantInt64)
			}
		})
	}
}

func TestFlexInt_UnmarshalJSON_String(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantVal float64
		wantTxt string
	}{
		{"numeric string", `"123"`, 123, "123"},
		{"zero string", `"0"`, 0, "0"},
		{"empty string", `""`, 0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var f FlexInt
			if err := json.Unmarshal([]byte(tt.input), &f); err != nil {
				t.Fatalf("UnmarshalJSON() error = %v", err)
			}
			if f.Val != tt.wantVal {
				t.Errorf("Val = %v, want %v", f.Val, tt.wantVal)
			}
			if f.Txt != tt.wantTxt {
				t.Errorf("Txt = %v, want %v", f.Txt, tt.wantTxt)
			}
		})
	}
}

func TestFlexInt_MarshalJSON(t *testing.T) {
	tests := []struct {
		name  string
		val   float64
		txt   string
		want  string
	}{
		{"numeric", 123, "", `123`},
		{"zero", 0, "", `0`},
		{"float", 123.45, "", `123.45`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := FlexInt{Val: tt.val, Txt: tt.txt}
			got, err := json.Marshal(f)
			if err != nil {
				t.Fatalf("MarshalJSON() error = %v", err)
			}
			if string(got) != tt.want {
				t.Errorf("MarshalJSON() = %s, want %s", string(got), tt.want)
			}
		})
	}
}

func TestFlexBool_UnmarshalJSON_Bool(t *testing.T) {
	tests := []struct {
		name string
		input string
		want bool
	}{
		{"true", `true`, true},
		{"false", `false`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var f FlexBool
			if err := json.Unmarshal([]byte(tt.input), &f); err != nil {
				t.Fatalf("UnmarshalJSON() error = %v", err)
			}
			if f.Bool() != tt.want {
				t.Errorf("Bool() = %v, want %v", f.Bool(), tt.want)
			}
		})
	}
}

func TestFlexBool_UnmarshalJSON_String(t *testing.T) {
	tests := []struct {
		name string
		input string
		want bool
	}{
		{"true string", `"true"`, true},
		{"false string", `"false"`, false},
		{"one string", `"1"`, true},
		{"zero string", `"0"`, false},
		{"other string", `"other"`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var f FlexBool
			if err := json.Unmarshal([]byte(tt.input), &f); err != nil {
				t.Fatalf("UnmarshalJSON() error = %v", err)
			}
			if f.Bool() != tt.want {
				t.Errorf("Bool() = %v, want %v", f.Bool(), tt.want)
			}
		})
	}
}

func TestFlexBool_UnmarshalJSON_Numeric(t *testing.T) {
	tests := []struct {
		name string
		input string
		want bool
	}{
		{"one", `1`, true},
		{"zero", `0`, false},
		{"positive", `42`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var f FlexBool
			if err := json.Unmarshal([]byte(tt.input), &f); err != nil {
				t.Fatalf("UnmarshalJSON() error = %v", err)
			}
			if f.Bool() != tt.want {
				t.Errorf("Bool() = %v, want %v", f.Bool(), tt.want)
			}
		})
	}
}

func TestFlexBool_MarshalJSON(t *testing.T) {
	tests := []struct {
		name string
		val  bool
		want string
	}{
		{"true", true, `true`},
		{"false", false, `false`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := FlexBool{Val: tt.val}
			got, err := json.Marshal(f)
			if err != nil {
				t.Fatalf("MarshalJSON() error = %v", err)
			}
			if string(got) != tt.want {
				t.Errorf("MarshalJSON() = %s, want %s", string(got), tt.want)
			}
		})
	}
}

func TestFlexString_UnmarshalJSON_String(t *testing.T) {
	var f FlexString
	input := `"hello"`
	if err := json.Unmarshal([]byte(input), &f); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}
	if f.Val != "hello" {
		t.Errorf("Val = %v, want hello", f.Val)
	}
	if len(f.Arr) != 1 || f.Arr[0] != "hello" {
		t.Errorf("Arr = %v, want [hello]", f.Arr)
	}
}

func TestFlexString_UnmarshalJSON_Array(t *testing.T) {
	var f FlexString
	input := `["hello", "world"]`
	if err := json.Unmarshal([]byte(input), &f); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}
	if f.Val != "hello" {
		t.Errorf("Val = %v, want hello", f.Val)
	}
	if len(f.Arr) != 2 {
		t.Errorf("Arr length = %d, want 2", len(f.Arr))
	}
}

func TestFlexString_MarshalJSON_Single(t *testing.T) {
	f := FlexString{Val: "hello", Arr: []string{"hello"}}
	got, err := json.Marshal(f)
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}
	want := `"hello"`
	if string(got) != want {
		t.Errorf("MarshalJSON() = %s, want %s", string(got), want)
	}
}

func TestFlexString_MarshalJSON_Multiple(t *testing.T) {
	f := FlexString{Val: "hello", Arr: []string{"hello", "world"}}
	got, err := json.Marshal(f)
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}
	want := `["hello","world"]`
	if string(got) != want {
		t.Errorf("MarshalJSON() = %s, want %s", string(got), want)
	}
}
