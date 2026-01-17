// pkg/format/serialize.go
package format

import (
	"bytes"
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"
)

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