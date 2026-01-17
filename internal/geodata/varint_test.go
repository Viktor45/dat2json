// internal/geodata/varint_test.go
package geodata

import (
	"bytes"
	"testing"
)

func TestReadVarintString(t *testing.T) {
	// "hello" (len=5)
	data := []byte{0x05, 'h', 'e', 'l', 'l', 'o'}
	r := bytes.NewReader(data)
	s, err := ReadVarintString(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s != "hello" {
		t.Errorf("expected 'hello', got %q", s)
	}
}

func TestReadVarintStringTooLong(t *testing.T) {
	data := []byte{0xff, 0xff} // huge length
	r := bytes.NewReader(data)
	_, err := ReadVarintString(r)
	if err == nil {
		t.Fatal("expected error")
	}
}
