// internal/geosite/decode_test.go
package geosite

import (
	"testing"

	"dat2json/internal/geodata/router"

	"google.golang.org/protobuf/proto"
)

func TestDecodeBinary(t *testing.T) {
	data := []byte("GEOS\x01")
	data = append(data, 0x06, 'g', 'o', 'o', 'g', 'l', 'e')                    // "google" (6)
	data = append(data, 0x01)                                                  // 1 domain
	data = append(data, 0x00)                                                  // type=domain
	data = append(data, 0x0b)                                                  // len=11 ← ИСПРАВЛЕНО
	data = append(data, '.', 'g', 'o', 'o', 'g', 'l', 'e', '.', 'c', 'o', 'm') // 11 bytes

	result, err := Decode(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 tag, got %d", len(result))
	}
	domains := result["google"]
	if len(domains) != 1 || domains[0] != "domain:.google.com" {
		t.Errorf("expected ['domain:.google.com'], got %v", domains)
	}
}

func TestDecodeProtobuf(t *testing.T) {
	geositeList := &router.GeoSiteList{
		Entry: []*router.GeoSite{
			{
				CountryCode: "youtube",
				Domain: []*router.Domain{
					{Type: router.Domain_Domain, Value: ".youtube.com"},
					{Type: router.Domain_Plain, Value: "googlevideo"}, // ← ИСПРАВЛЕНО
				},
			},
		},
	}
	data, _ := proto.Marshal(geositeList)
	result, err := Decode(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 tag, got %d", len(result))
	}
	domains := result["youtube"]
	expected := []string{"domain:.youtube.com", "keyword:googlevideo"}
	for i, e := range expected {
		if domains[i] != e {
			t.Errorf("domain[%d]: expected %q, got %q", i, e, domains[i])
		}
	}
}

func TestDecodeInvalid(t *testing.T) {
	_, err := Decode([]byte("INVALID"))
	if err == nil {
		t.Fatal("expected error")
	}
}
