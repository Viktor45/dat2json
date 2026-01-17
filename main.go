// main.go
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"dat2json/internal/geoip"
	"dat2json/internal/geosite"
	"dat2json/pkg/format"
)

var (
	inputFile     = flag.String("i", "", "Input .dat file")
	outputFile    = flag.String("o", "", "Output file (.json/.yaml/.yml)")
	outputDir     = flag.String("output-dir", "", "Output directory for per-tag/country files")
	tagFilter     = flag.String("tag", "", "Comma-separated tags (geosite only)")
	countryFilter = flag.String("country", "", "Comma-separated country codes (geoip only)")
	listTags      = flag.Bool("list-tags", false, "List all tags in geosite.dat and exit")
	sortKeys      = flag.Bool("sort", false, "Sort keys")
	formatFlag    = flag.String("format", "", "Output format: json or yaml")
	ipMode        = flag.Bool("ip", false, "Treat input as geoip.dat")
	siteMode      = flag.Bool("site", false, "Treat input as geosite.dat")
	help          = flag.Bool("h", false, "Show help")
)

func isValidFormat(f string) bool {
	return f == "json" || f == "yaml"
}

func getOutputFormat() (string, error) {
	if *formatFlag != "" {
		if !isValidFormat(*formatFlag) {
			return "", fmt.Errorf("--format must be 'json' or 'yaml'")
		}
		return *formatFlag, nil
	}

	if *outputFile != "" {
		ext := strings.ToLower(filepath.Ext(*outputFile))
		switch ext {
		case ".json":
			return "json", nil
		case ".yaml", ".yml":
			return "yaml", nil
		default:
			return "", fmt.Errorf("cannot determine format from extension %q", ext)
		}
	}

	if *outputDir != "" {
		return "yaml", nil
	}

	return "", fmt.Errorf("unable to determine output format")
}

func parseList(listStr string, toUpper bool) []string {
	if listStr == "" {
		return nil
	}
	parts := strings.Split(listStr, ",")
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			if toUpper {
				items = append(items, strings.ToUpper(trimmed))
			} else {
				items = append(items, strings.ToLower(trimmed))
			}
		}
	}
	return items
}

type progressReporter struct {
	total      int
	lastReport time.Time
	mu         sync.Mutex
}

func newProgressReporter(total int) *progressReporter {
	return &progressReporter{
		total:      total,
		lastReport: time.Now(),
	}
}

func (p *progressReporter) maybeReport(current int) {
	if p.total <= 10000 {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	now := time.Now()
	if now.Sub(p.lastReport) < 500*time.Millisecond {
		return
	}
	fmt.Fprintf(os.Stderr, "\rProcessing... %d/%d (%.1f%%)", current, p.total, float64(current)/float64(p.total)*100)
	p.lastReport = now
}

func (p *progressReporter) finish() {
	if p.total > 10000 {
		p.mu.Lock()
		defer p.mu.Unlock()
		fmt.Fprintln(os.Stderr, "\rProcessing... done.            ")
	}
}

func writeFileSafe(path string, data []byte) error {
	dir := filepath.Dir(path)
	if dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return os.WriteFile(path, data, 0o644)
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s -i input.dat [options]\n", os.Args[0])
		fmt.Fprintln(os.Stderr, "\nOptions:")
		fmt.Fprintln(os.Stderr, "  -i FILE             Input .dat file (required)")
		fmt.Fprintln(os.Stderr, "  --ip                Treat input as geoip.dat")
		fmt.Fprintln(os.Stderr, "  --site              Treat input as geosite.dat")
		fmt.Fprintln(os.Stderr, "  -o FILE             Output file (.json/.yaml/.yml)")
		fmt.Fprintln(os.Stderr, "  --output-dir DIR    Output each tag/country to separate file")
		fmt.Fprintln(os.Stderr, "  --format FMT        Output format: json or yaml")
		fmt.Fprintln(os.Stderr, "  --tag LIST          Filter geosite by tags")
		fmt.Fprintln(os.Stderr, "  --country LIST      Filter geoip by country codes")
		fmt.Fprintln(os.Stderr, "  --list-tags         List all tags in geosite.dat and exit")
		fmt.Fprintln(os.Stderr, "  --sort              Sort keys")
		fmt.Fprintln(os.Stderr, "  -h                  Show this help")
	}
	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	if *inputFile == "" {
		log.Fatal("error: -i input file is required")
	}

	if *outputFile != "" && *outputDir != "" {
		log.Fatal("error: cannot use both -o and --output-dir")
	}

	if !*listTags && *outputFile == "" && *outputDir == "" {
		log.Fatal("error: either -o, --output-dir, or --list-tags must be specified")
	}

	if *ipMode && *siteMode {
		log.Fatal("error: cannot use both --ip and --site")
	}

	if !*ipMode && !*siteMode {
		log.Fatal("error: must specify --ip or --site")
	}

	outFormat, err := getOutputFormat()
	if err != nil && !*listTags {
		log.Fatal("error:", err)
	}

	data, err := os.ReadFile(*inputFile)
	if err != nil {
		log.Fatal("error reading input file:", err)
	}

	if len(data) == 0 {
		log.Fatal("error: input file is empty")
	}

	isGeoSite := *siteMode
	var fullResult map[string][]string
	var decodeErr error

	if *ipMode {
		fullResult, decodeErr = geoip.Decode(data)
		if decodeErr != nil {
			log.Fatalf("error decoding as geoip.dat: %v", decodeErr)
		}
	} else if *siteMode {
		fullResult, decodeErr = geosite.Decode(data)
		if decodeErr != nil {
			log.Fatalf("error decoding as geosite.dat: %v", decodeErr)
		}
	}

	// === ОБРАБОТКА --list-tags ===
	if *listTags {
		if !isGeoSite {
			log.Fatal("--list-tags is only supported for geosite.dat (--site)")
		}
		tags := make([]string, 0, len(fullResult))
		for tag := range fullResult {
			tags = append(tags, tag)
		}
		if *sortKeys {
			sort.Strings(tags)
		}
		for _, tag := range tags {
			fmt.Println(tag)
		}
		os.Exit(0)
	}

	// === ПРИМЕНЕНИЕ ФИЛЬТРОВ ===
	var filtered map[string][]string
	warnings := []string{}

	if isGeoSite {
		if *countryFilter != "" {
			warnings = append(warnings, "--country is ignored for geosite.dat")
		}
		if *tagFilter != "" {
			tags := parseList(*tagFilter, false)
			filtered = make(map[string][]string)
			lowerMap := make(map[string]string)
			for k := range fullResult {
				lowerMap[strings.ToLower(k)] = k
			}
			for _, t := range tags {
				if origTag, ok := lowerMap[t]; ok {
					domains := fullResult[origTag]
					if *sortKeys {
						sort.Strings(domains)
					}
					filtered[origTag] = domains
				} else {
					warnings = append(warnings, fmt.Sprintf("tag '%s' not found", t))
				}
			}
			if len(filtered) == 0 {
				log.Fatal("error: no valid tags found")
			}
		} else {
			filtered = fullResult
			if *sortKeys {
				keys := make([]string, 0, len(filtered))
				for k := range filtered {
					keys = append(keys, k)
				}
				sort.Strings(keys)
				sorted := make(map[string][]string)
				for _, k := range keys {
					domains := filtered[k]
					if *sortKeys {
						sort.Strings(domains)
					}
					sorted[k] = domains
				}
				filtered = sorted
			}
		}
	} else {
		if *tagFilter != "" {
			warnings = append(warnings, "--tag is ignored for geoip.dat")
		}
		if *countryFilter != "" {
			countries := parseList(*countryFilter, true)
			filtered = make(map[string][]string)
			for _, code := range countries {
				if cidrs, ok := fullResult[code]; ok {
					filtered[code] = cidrs
				} else {
					warnings = append(warnings, fmt.Sprintf("country code '%s' not found", code))
				}
			}
			if len(filtered) == 0 {
				log.Fatal("error: no valid country codes found")
			}
		} else {
			filtered = fullResult
		}
		if *sortKeys {
			keys := make([]string, 0, len(filtered))
			for k := range filtered {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			sorted := make(map[string][]string)
			for _, k := range keys {
				sorted[k] = filtered[k]
			}
			filtered = sorted
		}
	}

	// === ВЫВОД ПРЕДУПРЕЖДЕНИЙ ===
	for _, w := range warnings {
		fmt.Fprintf(os.Stderr, "⚠️ Warning: %s\n", w)
	}

	// === ЭКСПОРТ ===
	if *outputDir != "" {
		if err := os.MkdirAll(*outputDir, 0o755); err != nil {
			log.Fatal("error creating output directory:", err)
		}

		ext := "yaml"
		if outFormat == "json" {
			ext = "json"
		}

		var wg sync.WaitGroup
		sem := make(chan struct{}, 32)

		for key, entries := range filtered {
			wg.Add(1)
			go func(k string, v []string) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				single := map[string][]string{k: v}
				data, err := format.Serialize(single, outFormat)
				if err != nil {
					log.Fatalf("error serializing %s: %v", k, err)
				}
				filename := fmt.Sprintf("%s.%s", k, ext)
				path := filepath.Join(*outputDir, filename)
				if err := writeFileSafe(path, data); err != nil {
					log.Fatalf("error writing %s: %v", path, err)
				}
			}(key, entries)
		}
		wg.Wait()
		fmt.Printf("✅ Exported %d files to %s (%s)\n", len(filtered), *outputDir, outFormat)
	} else if *outputFile != "" {
		outBytes, err := format.Serialize(filtered, outFormat)
		if err != nil {
			log.Fatal("error serializing output:", err)
		}
		if err := writeFileSafe(*outputFile, outBytes); err != nil {
			log.Fatal("error writing output file:", err)
		}
		desc := outFormat
		if isGeoSite && *tagFilter != "" {
			desc += " (tags: " + *tagFilter + ")"
		} else if !isGeoSite && *countryFilter != "" {
			desc += " (countries: " + *countryFilter + ")"
		}
		if *sortKeys {
			desc += " + sorted"
		}
		fmt.Printf("✅ Successfully converted %s → %s (%s)\n", *inputFile, *outputFile, desc)
	}
}
