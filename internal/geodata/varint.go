// Package geodata provides utilities for decoding binary geolocation data formats.
// internal/geodata/varint.go
package geodata

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// ReadVarintString reads a varint-prefixed string from the reader.
// Returns error if length is invalid or exceeds remaining bytes.
func ReadVarintString(r *bytes.Reader) (string, error) {
	length, err := binary.ReadUvarint(r)
	if err != nil {
		return "", err
	}
	if length > uint64(r.Len()) {
		return "", fmt.Errorf("string too long")
	}
	buf := make([]byte, length)
	if _, err := r.Read(buf); err != nil {
		return "", err
	}
	return string(buf), nil
}
