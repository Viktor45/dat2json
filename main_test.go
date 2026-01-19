// main_test.go
package main

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"dat2json/internal/geodata/router"

	"google.golang.org/protobuf/proto"
)

// resetFlags clears all flag state to ensure clean test isolation.
func resetFlags() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	// Recreate all flags with their default values.
	inputFile = flag.String("i", "", "")
	outputFile = flag.String("o", "", "")
	outputDir = flag.String("output-dir", "", "")
	tagFilter = flag.String("tag", "", "")
	countryFilter = flag.String("country", "", "")
	listTags = flag.Bool("list-tags", false, "")
	sortKeys = flag.Bool("sort", false, "")
	formatFlag = flag.String("format", "", "")
	ipMode = flag.Bool("ip", false, "")
	siteMode = flag.Bool("site", false, "")
	help = flag.Bool("h", false, "")
}

func TestIntegrationGeoIPBinary(t *testing.T) {
	resetFlags()
	inputFile := filepath.Join(t.TempDir(), "geoip.dat")
	outputFile := filepath.Join(t.TempDir(), "output.json")

	data := []byte("GEOI\x01\x02US\x01\x01\x02\x03\x04\x18")
	os.WriteFile(inputFile, data, 0o644)

	os.Args = []string{"dat2json", "-i", inputFile, "--ip", "-o", outputFile}

	defer func() { recover() }()

	func() {
		defer func() {
			if r := recover(); r != nil {
				return
			}
		}()
		main()
	}()

	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), `"US"`) {
		t.Errorf("unexpected output: %s", string(content))
	}
}

func TestIntegrationGeoSiteProtobuf(t *testing.T) {
	resetFlags()
	inputFile := filepath.Join(t.TempDir(), "geosite.pb")
	outputFile := filepath.Join(t.TempDir(), "output.yaml")

	geositeList := &router.GeoSiteList{
		Entry: []*router.GeoSite{
			{
				CountryCode: "test",
				Domain: []*router.Domain{
					{Type: router.Domain_Domain, Value: ".example.com"},
				},
			},
		},
	}
	data, _ := proto.Marshal(geositeList)
	os.WriteFile(inputFile, data, 0o644)

	// Только --site
	os.Args = []string{"dat2json", "-i", inputFile, "--site", "-o", outputFile} // ← ИСПРАВЛЕНО

	defer func() { recover() }()

	func() {
		defer func() {
			if r := recover(); r != nil {
				return
			}
		}()
		main()
	}()

	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "test:") {
		t.Errorf("unexpected output: %s", string(content))
	}
}
