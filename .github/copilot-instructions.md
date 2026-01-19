# Copilot Instructions for dat2json

## Project Overview

**dat2json** is a high-performance CLI tool for decoding binary `geoip.dat` and `geosite.dat` files (used by Xray/Mihomo) into JSON/YAML formats. The project supports two distinct file formats: binary (with "GEOI"/"GEOS" magic headers) and Protobuf-based formats.

### Architecture

The codebase is organized into three main layers:

1. **`main.go`** — CLI orchestration
   - Flag parsing (mandatory `--ip` or `--site` mode selector)
   - Dispatches to appropriate decoder (`geoip` or `geosite`)
   - Handles output routing (single file or per-tag directory with parallel workers)
   - Implements filter logic (e.g., `--country=US,DE` or `--tag=google,netflix`)

2. **Decoders** (`internal/geoip`, `internal/geosite`)
   - Format detection: binary (magic header check) vs. Protobuf
   - Binary: Uses varint encoding + fixed-width IP/mask bytes
   - Protobuf: Unmarshals `router.GeoIP` or `router.GeoSite` messages
   - Returns normalized `map[string][]string` (country/tag → values)

3. **Serializers** (`pkg/format/serialize.go`)
   - Converts maps to JSON/YAML with optional sorting
   - Respects format inferred from file extension or explicit `--format` flag

### Data Flow

```
.dat file → [Binary/Protobuf detection] → [Decode to map] → [Filter by country/tag] → [Serialize] → JSON/YAML
```

## Critical Patterns & Conventions

### Format Detection

Both decoders use **magic header detection**:

- `geoip.dat`: Starts with `"GEOI"` (binary) → falls back to Protobuf
- `geosite.dat`: Starts with `"GEOS"` (binary) → falls back to Protobuf

This dual-support is intentional and tested in `main_test.go`. Never assume one format only.

### Varint Encoding

The `internal/geodata/varint.go` utility decodes varint-prefixed strings used in both binary formats:

```go
ReadVarintString(r *bytes.Reader) // reads: [varint length][N bytes of data]
```

This pattern appears in country code, tag names, and domain values. Always validate `length > r.Len()` to prevent overruns.

### CLI Mode Selection

The `--ip` and `--site` flags are **mutually exclusive and required**:

- Validation in `main.go`: Both cannot be true; at least one must be
- This explicitness prevents ambiguous behavior when processing unfamiliar files
- Reflected in integration tests via `flag` resets

### Output Directory with Parallel Writers

When using `--output-dir`, the code uses a `WaitGroup` + goroutine pool (up to 32 concurrent writers) for performance:

- Each country/tag writes to `{tag_or_country}.yaml` or `.json`
- Progress bar shown for >10k entries
- Sorting optional via `--sort` flag (deterministic output for diffs/versioning)

### Error Handling

- **Immediate validation**: Input file existence, format correctness checked before processing
- **Boundary checks**: Varint readers check length against `r.Len()` before allocation
- **No silent failures**: All decode errors wrapped with context (e.g., "read country code: %w")

## Testing

- **`main_test.go`** — Integration tests with temporary binary data
- **`internal/geoip/decode_test.go`** and **`internal/geosite/decode_test.go`** — Format-specific decoding
- **Test pattern**: `TestIntegration*` for binary format, `TestDecodeProtobuf*` for Protobuf
- **Flag reset**: `resetFlags()` clears global flag state between tests (important for multi-test suites)

Build & run tests:

```bash
go test ./... -v  # Full test suite
go test ./internal/geoip -v  # Single package
```

## Go Code Standards

Refer to [`.github/instructions/go.instructions.md`](./instructions/go.instructions.md) for detailed Go idioms. Key points:

- **Naming**: Package names lowercase, exported functions PascalCase, unexported camelCase
- **Interfaces**: Use `-er` suffix (e.g., `Reader`, `Writer`)
- **Comments**: Document exported types and complex logic; write "why" not "what"
- **Error handling**: Check errors immediately after calls; avoid `_` unless justified
- **Package declarations**: Each `.go` file has exactly ONE `package` declaration matching its directory

## Building & Running

```bash
# Build binary
go build -o dat2json .

# Run with examples
./dat2json -i geoip.dat --ip -o countries.json
./dat2json -i geosite.dat --site --output-dir ./rules --tag=netflix,google
./dat2json -i geoip.dat --ip --list-tags

# Help
./dat2json -h
```

## External Dependencies

- **`google.golang.org/protobuf`** — Unmarshalling `router.GeoIP` / `router.GeoSite` messages (auto-generated from `internal/geodata/router/config.proto`)
- **`gopkg.in/yaml.v3`** — YAML serialization with custom indent settings

## Key Files & Their Roles

| File                                                        | Purpose                                                |
| ----------------------------------------------------------- | ------------------------------------------------------ |
| [main.go](../main.go)                                       | CLI entry point, flag parsing, output dispatch         |
| [internal/geoip/decode.go](../internal/geoip/decode.go)     | GeoIP binary/Protobuf decoding, CIDR normalization     |
| [internal/geosite/decode.go](../internal/geosite/decode.go) | GeoSite binary/Protobuf decoding, domain rule prefixes |
| [internal/geodata/varint.go](../internal/geodata/varint.go) | Shared varint string reader utility                    |
| [pkg/format/serialize.go](../pkg/format/serialize.go)       | JSON/YAML formatting with optional sorting             |

---

**Last Updated:** January 2026  
**Go Version:** 1.23+
