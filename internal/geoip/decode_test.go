// internal/geoip/decode_test.go
package geoip

import (
	"net"
	"testing"

	"dat2json/internal/geodata/router"

	"google.golang.org/protobuf/proto"
)

func TestDecodeBinary(t *testing.T) {
	data := []byte("GEOI\x01")
	data = append(data, 0x02, 'U', 'S') // "US"
	data = append(data, 0x01)           // 1 CIDR
	data = append(data, 1, 2, 3, 4, 24) // 1.2.3.4/24

	result, err := Decode(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 country, got %d", len(result))
	}
	cidrs := result["US"]
	if len(cidrs) != 1 || cidrs[0] != "1.2.3.4/24" {
		t.Errorf("expected ['1.2.3.4/24'], got %v", cidrs)
	}
}

func TestDecodeProtobuf(t *testing.T) {
	geoipList := &router.GeoIPList{
		Entry: []*router.GeoIP{
			{
				CountryCode: "CN",
				Cidr: []*router.CIDR{
					{Ip: net.IP{1, 0, 0, 0}.To4(), Prefix: 24},
				},
			},
		},
	}
	data, _ := proto.Marshal(geoipList)
	result, err := Decode(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 country, got %d", len(result))
	}
	cidrs := result["CN"]
	if len(cidrs) != 1 || cidrs[0] != "1.0.0.0/24" {
		t.Errorf("expected ['1.0.0.0/24'], got %v", cidrs)
	}
}

func TestDecodeInvalid(t *testing.T) {
	_, err := Decode([]byte("INVALID"))
	if err == nil {
		t.Fatal("expected error")
	}
}
