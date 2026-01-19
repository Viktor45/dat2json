// Package geosite provides decoding functionality for geosite.dat files in binary or Protobuf format.
package geosite

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"dat2json/internal/geodata"
	"dat2json/internal/geodata/router"

	"google.golang.org/protobuf/proto"
)

const (
	magicHeaderGeoSite = "GEOS"
	magicHeaderSize    = 4
)

// ErrInvalidFormat is returned when the data does not represent a valid geosite.dat file.
var ErrInvalidFormat = fmt.Errorf("not a valid geosite.dat file")

// Decode decodes binary or Protobuf geosite data into a map of tags to domain lists.
func Decode(data []byte) (map[string][]string, error) {
	if len(data) >= magicHeaderSize && string(data[:magicHeaderSize]) == magicHeaderGeoSite {
		return decodeBinary(data)
	}
	return decodeProtobuf(data)
}

// domainTypePrefix returns the prefix string for a given domain type byte.
func domainTypePrefix(domainType byte) string {
	switch domainType {
	case 0:
		return "domain:"
	case 1:
		return "full:"
	case 2:
		return "regexp:"
	case 3:
		return "keyword:"
	default:
		return fmt.Sprintf("type%d:", domainType)
	}
}

func decodeBinary(data []byte) (map[string][]string, error) {
	if len(data) < 5 {
		return nil, ErrInvalidFormat
	}
	r := bytes.NewReader(data[5:])
	result := make(map[string][]string)

	for r.Len() > 0 {
		tagName, err := geodata.ReadVarintString(r)
		if err != nil {
			return nil, fmt.Errorf("read tag name: %w", err)
		}

		count, err := binary.ReadUvarint(r)
		if err != nil {
			return nil, fmt.Errorf("read domain count: %w", err)
		}

		var domains []string
		for i := uint64(0); i < count; i++ {
			domainType, err := r.ReadByte()
			if err != nil {
				return nil, fmt.Errorf("read domain type: %w", err)
			}

			value, err := geodata.ReadVarintString(r)
			if err != nil {
				return nil, fmt.Errorf("read domain value: %w", err)
			}

			domains = append(domains, domainTypePrefix(domainType)+value)
		}

		result[tagName] = domains
	}

	return result, nil
}

// protobufDomainTypePrefix returns the prefix for Protobuf router.Domain types.
func protobufDomainTypePrefix(t router.Domain_Type) string {
	switch t {
	case router.Domain_Domain:
		return "domain:"
	case router.Domain_Full:
		return "full:"
	case router.Domain_Regex:
		return "regexp:"
	case router.Domain_Plain:
		return "keyword:"
	default:
		return fmt.Sprintf("type%d:", t)
	}
}

func decodeProtobuf(data []byte) (map[string][]string, error) {
	var list router.GeoSiteList
	if err := proto.Unmarshal(data, &list); err != nil {
		return nil, ErrInvalidFormat
	}

	result := make(map[string][]string)
	for _, site := range list.Entry {
		var domains []string
		for _, d := range site.Domain {
			domains = append(domains, protobufDomainTypePrefix(d.GetType())+d.GetValue())
		}
		result[site.CountryCode] = domains
	}

	return result, nil
}

// IsValid checks if the data is a valid geosite.dat file (binary or Protobuf format).
func IsValid(data []byte) bool {
	if len(data) >= magicHeaderSize && string(data[:magicHeaderSize]) == magicHeaderGeoSite {
		return true
	}
	_, err := decodeProtobuf(data)
	return err == nil
}
