package main

import (
	"encoding/json"
	"io"
)

// writeJSON emits v as indented JSON. Indented because this output is read
// by people at least as often as by jq (C-GLOBAL-011).
func writeJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
