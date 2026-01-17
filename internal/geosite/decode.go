// internal/geosite/decode.go
package geosite

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"dat2json/internal/geodata"
	"dat2json/internal/geodata/router"

	"google.golang.org/protobuf/proto"
)

var ErrInvalidFormat = fmt.Errorf("not a valid geosite.dat file")

func Decode(data []byte) (map[string][]string, error) {
	if len(data) >= 4 && string(data[:4]) == "GEOS" {
		return decodeBinary(data)
	}
	return decodeProtobuf(data)
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

			prefix := ""
			switch domainType {
			case 0:
				prefix = "domain:"
			case 1:
				prefix = "full:"
			case 2:
				prefix = "regexp:"
			case 3:
				prefix = "keyword:"
			default:
				prefix = fmt.Sprintf("type%d:", domainType)
			}
			domains = append(domains, prefix+value)
		}

		result[tagName] = domains
	}

	return result, nil
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
			prefix := ""
			switch d.GetType() {
			case router.Domain_Domain:
				prefix = "domain:"
			case router.Domain_Full:
				prefix = "full:"
			case router.Domain_Regex:
				prefix = "regexp:"
			case router.Domain_Plain:
				prefix = "keyword:"
			default:
				prefix = fmt.Sprintf("type%d:", d.GetType())
			}
			domains = append(domains, prefix+d.GetValue())
		}
		result[site.CountryCode] = domains
	}

	return result, nil
}

func IsValid(data []byte) bool {
	if len(data) >= 4 {
		if string(data[:4]) == "GEOI" || string(data[:4]) == "GEOS" {
			return true
		}
	}
	_, err := decodeProtobuf(data)
	return err == nil
}
