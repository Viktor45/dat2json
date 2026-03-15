// Package geodata provides utilities for decoding binary geolocation data formats.
// internal/geodata/varint.go
package geodata

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

const maxVarintStringLen = 1 << 20 // 1 MiB

// ReadVarintString reads a varint-prefixed string from the reader.
// It includes bounds checks to avoid unbounded memory allocation from untrusted data.
func ReadVarintString(r *bytes.Reader) (string, error) {
	length, err := binary.ReadUvarint(r)
	if err != nil {
		return "", err
	}
	if length > uint64(r.Len()) {
		return "", fmt.Errorf("string length exceeds remaining bytes")
	}
	if length > maxVarintStringLen {
		return "", fmt.Errorf("string length too large")
	}
	buf := make([]byte, length)
	if _, err := r.Read(buf); err != nil {
		return "", err
	}
	return string(buf), nil
}
