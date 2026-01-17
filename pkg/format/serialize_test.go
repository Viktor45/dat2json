// pkg/format/serialize_test.go
package format

import (
	"strings"
	"testing"
)

func TestSerializeJSON(t *testing.T) {
	data := map[string][]string{"test": {"a", "b"}}
	out, err := Serialize(data, "json")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), `"test"`) {
		t.Errorf("unexpected JSON: %s", out)
	}
}

func TestSerializeYAML(t *testing.T) {
	data := map[string][]string{"test": {"a", "b"}}
	out, err := Serialize(data, "yaml")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "test:") {
		t.Errorf("unexpected YAML: %s", out)
	}
}

func TestSerializeUnknown(t *testing.T) {
	_, err := Serialize(nil, "xml")
	if err == nil {
		t.Fatal("expected error")
	}
}
