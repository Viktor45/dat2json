// Package format provides serialization utilities for converting data to JSON or YAML formats.
package format

import (
	"bytes"
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"
)

// Serialize converts a map of strings to JSON or YAML bytes based on the specified format.
func Serialize(data map[string][]string, format string) ([]byte, error) {
	switch format {
	case "json":
		return json.MarshalIndent(data, "", "  ")
	case "yaml":
		buf := &bytes.Buffer{}
		enc := yaml.NewEncoder(buf)
		enc.SetIndent(2)
		if err := enc.Encode(data); err != nil {
			return nil, err
		}
		out := buf.Bytes()
		if len(out) > 0 && out[len(out)-1] == '\n' {
			out = out[:len(out)-1]
		}
		return out, nil
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}
