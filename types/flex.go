package types

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// FlexInt handles JSON fields that may be either numbers or strings.
// UniFi API sometimes returns numbers as strings (e.g., "123" instead of 123).
type FlexInt struct {
	Val float64
	Txt string
}

// UnmarshalJSON implements json.Unmarshaler.
func (f *FlexInt) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as number first
	var num float64
	if err := json.Unmarshal(data, &num); err == nil {
		f.Val = num
		f.Txt = fmt.Sprintf("%.0f", num)
		return nil
	}

	// Try as string
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	f.Txt = str
	if str != "" {
		// Try to parse string as number
		if num, err := strconv.ParseFloat(str, 64); err == nil {
			f.Val = num
		}
	}

	return nil
}

// MarshalJSON implements json.Marshaler.
func (f FlexInt) MarshalJSON() ([]byte, error) {
	// Prefer numeric representation if we have a valid value
	if f.Val != 0 || f.Txt == "" || f.Txt == "0" {
		return json.Marshal(f.Val)
	}
	return json.Marshal(f.Txt)
}

// Int returns the value as an int.
func (f FlexInt) Int() int {
	return int(f.Val)
}

// Int64 returns the value as an int64.
func (f FlexInt) Int64() int64 {
	return int64(f.Val)
}

// Float64 returns the value as a float64.
func (f FlexInt) Float64() float64 {
	return f.Val
}

// String returns the string representation.
func (f FlexInt) String() string {
	if f.Txt != "" {
		return f.Txt
	}
	return fmt.Sprintf("%.0f", f.Val)
}

// FlexBool handles JSON fields that may be booleans, strings, or numbers.
// UniFi API sometimes returns:
//   - true/false
//   - "true"/"false"
//   - 1/0
type FlexBool struct {
	Val bool
	Txt string
}

// UnmarshalJSON implements json.Unmarshaler.
func (f *FlexBool) UnmarshalJSON(data []byte) error {
	// Try boolean first
	var b bool
	if err := json.Unmarshal(data, &b); err == nil {
		f.Val = b
		f.Txt = strconv.FormatBool(b)
		return nil
	}

	// Try string
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		f.Txt = str
		f.Val = str == "true" || str == "1"
		return nil
	}

	// Try number
	var num int
	if err := json.Unmarshal(data, &num); err == nil {
		f.Val = num != 0
		f.Txt = strconv.FormatBool(f.Val)
		return nil
	}

	return fmt.Errorf("cannot unmarshal %s as FlexBool", string(data))
}

// MarshalJSON implements json.Marshaler.
func (f FlexBool) MarshalJSON() ([]byte, error) {
	return json.Marshal(f.Val)
}

// Bool returns the boolean value.
func (f FlexBool) Bool() bool {
	return f.Val
}

// String returns the string representation.
func (f FlexBool) String() string {
	return f.Txt
}

// FlexString handles JSON fields that may be either a string or an array of strings.
type FlexString struct {
	Val string
	Arr []string
}

// UnmarshalJSON implements json.Unmarshaler.
func (f *FlexString) UnmarshalJSON(data []byte) error {
	// Try string first
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		f.Val = str
		f.Arr = []string{str}
		return nil
	}

	// Try array
	var arr []string
	if err := json.Unmarshal(data, &arr); err != nil {
		return err
	}

	f.Arr = arr
	if len(arr) > 0 {
		f.Val = arr[0]
	}

	return nil
}

// MarshalJSON implements json.Marshaler.
func (f FlexString) MarshalJSON() ([]byte, error) {
	// If we have multiple values, marshal as array
	if len(f.Arr) > 1 {
		return json.Marshal(f.Arr)
	}
	// Otherwise marshal as string
	return json.Marshal(f.Val)
}

// String returns the first string value.
func (f FlexString) String() string {
	return f.Val
}

// Strings returns all string values.
func (f FlexString) Strings() []string {
	return f.Arr
}
