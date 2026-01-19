// internal/geoip/decode.go
package geoip

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"

	"dat2json/internal/geodata"
	"dat2json/internal/geodata/router"

	"google.golang.org/protobuf/proto"
)

const (
	magicHeaderGeoIP = "GEOI"
	magicHeaderSize  = 4
)

var ErrInvalidFormat = fmt.Errorf("not a valid geoip.dat file")

func Decode(data []byte) (map[string][]string, error) {
	if len(data) >= magicHeaderSize && string(data[:magicHeaderSize]) == magicHeaderGeoIP {
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
		countryCode, err := geodata.ReadVarintString(r)
		if err != nil {
			return nil, fmt.Errorf("read country code: %w", err)
		}

		count, err := binary.ReadUvarint(r)
		if err != nil {
			return nil, fmt.Errorf("read CIDR count: %w", err)
		}

		var cidrs []string
		for i := uint64(0); i < count; i++ {
			ip4 := make([]byte, 4)
			if _, err := r.Read(ip4); err != nil {
				return nil, fmt.Errorf("read IP prefix: %w", err)
			}

			mask, err := r.ReadByte()
			if err != nil {
				return nil, fmt.Errorf("read mask: %w", err)
			}

			if mask <= 32 {
				ip := net.IP(ip4)
				cidrs = append(cidrs, fmt.Sprintf("%s/%d", ip, mask))
			} else {
				ip6Suffix := make([]byte, 12)
				if _, err := r.Read(ip6Suffix); err != nil {
					return nil, fmt.Errorf("read IPv6 suffix: %w", err)
				}
				ip := net.IP(append(ip4, ip6Suffix...))
				cidrs = append(cidrs, fmt.Sprintf("%s/%d", ip, mask))
			}
		}

		result[countryCode] = cidrs
	}

	return result, nil
}

func decodeProtobuf(data []byte) (map[string][]string, error) {
	var list router.GeoIPList
	if err := proto.Unmarshal(data, &list); err != nil {
		return nil, ErrInvalidFormat
	}

	result := make(map[string][]string)
	for _, geoip := range list.Entry {
		var cidrs []string
		for _, cidr := range geoip.Cidr {
			ipStr := net.IP(cidr.Ip).String()
			cidrs = append(cidrs, fmt.Sprintf("%s/%d", ipStr, cidr.Prefix))
		}
		result[geoip.CountryCode] = cidrs
	}

	return result, nil
}

func IsValid(data []byte) bool {
	if len(data) >= magicHeaderSize && string(data[:magicHeaderSize]) == magicHeaderGeoIP {
		return true
	}
	_, err := decodeProtobuf(data)
	return err == nil
}
