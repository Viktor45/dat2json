# `dat2json` — Decode/restore Xray/Mihomo `geoip.dat/geosite.dat` Files to JSON/YAML source

[![Go Version](https://img.shields.io/badge/Go-1.23%2B-blue)](https://golang.org)  [![License](https://img.shields.io/badge/License-MIT-green)](LICENSE)

- [`dat2json` — Decode/restore Xray/Mihomo `geoip.dat/geosite.dat` Files to JSON/YAML source](#dat2json--decoderestore-xraymihomo-geoipdatgeositedat-files-to-jsonyaml-source)
  - [✨ Features](#-features)
  - [🚀 Installation](#-installation)
    - [Prerequisites](#prerequisites)
    - [From Source](#from-source)
  - [📖 Usage](#-usage)
    - [Basic Examples](#basic-examples)
    - [Full Flag Reference](#full-flag-reference)
  - [🧪 Examples](#-examples)
    - [1. Inspect a Custom `geosite.dat`](#1-inspect-a-custom-geositedat)
    - [2. Build a Minimal GeoIP for EU](#2-build-a-minimal-geoip-for-eu)
    - [3. Bulk Export All Countries (Parallel)](#3-bulk-export-all-countries-parallel)
    - [4. Work with Protobuf Files (Mihomo Runtime)](#4-work-with-protobuf-files-mihomo-runtime)
  - [🛠 Technical Details](#-technical-details)
    - [File Format Support](#file-format-support)
    - [Output Format](#output-format)
    - [Performance](#performance)
  - [🧪 Testing](#-testing)
  - [🤝 Contributing](#-contributing)
  - [📄 License](#-license)
  - [🙏 Acknowledgements](#-acknowledgements)


`dat2json` is a high-performance CLI tool for decoding **`geoip.dat`** and **`geosite.dat`** files used by [Xray](https://github.com/XTLS/Xray-core), [Mihomo (Clash Meta)](https://github.com/MetaCubeX/mihomo), and other V2Ray-based projects. It supports full or filtered export to **JSON** or **YAML**, with parallel processing, progress tracking, and multi-file output.

> 🔍 **Why use this?**  
> Official `.dat` files are binary and opaque. This tool lets you recover binary geolocation data files — **without relying on closed-source generators**.

---

## ✨ Features

- ✅ **Decode both formats**:  
  - `geoip.dat` → country → CIDR lists (`1.2.3.0/24`)  
  - `geosite.dat` → tag → domain rules (`domain:`, `full:`, `regexp:`, `keyword:`)
- 📁 **Flexible output**:  
  - Single file (`-o output.json`)  
  - One file per tag/country (`--output-dir ./export`)
- 📦 **Multiple formats**: JSON, YAML (`.yaml` or `.yml`)
- 🔍 **Filtering**:  
  - `--tag=google,netflix` (for `geosite.dat`)  
  - `--country=US,DE,CN` (for `geoip.dat`)
- 🔤 **Sorting**: `--sort` for deterministic, readable output
- 📊 **Progress bar**: Automatic for large files (>10k entries)
- ⚡ **Parallel export**: Up to 32 concurrent writers for `--output-dir`
- 🧩 **Universal compatibility**: Works with:
  - Official `.dat` from [`v2fly/geoip`](https://github.com/v2fly/geoip)
  - Official `.dat` from [`v2fly/domain-list-community`](https://github.com/v2fly/domain-list-community)
  - Protobuf files from Mihomo runtime
- 🛡️ **Explicit mode selection**: No ambiguity — you **must specify** `--ip` or `--site`

---

## 🚀 Installation

### Prerequisites

- **Go 1.23 or higher**
- Git

### From Source

```bash
git clone https://github.com/viktor45/dat2json.git
cd dat2json
go build -o dat2json .
```

> 💡 Add to `PATH`:  
> ```bash
> sudo mv dat2json /usr/local/bin/
> ```

---

## 📖 Usage

### Basic Examples

```bash
# Convert geoip.dat to JSON (explicit --ip required)
./dat2json -i geoip.dat --ip -o countries.json

# Export only US and CN to YAML
./dat2json -i geoip.dat --ip -o eu.yaml --country=NL,RU

# List all tags in geosite.dat
./dat2json -i geosite.dat --site --list-tags

# Export Netflix and Google to separate YAML files
./dat2json -i geosite.dat --site --output-dir ./rules --tag=netflix,google
```

### Full Flag Reference

| Flag               | Description                                               | Required                                      |
| ------------------ | --------------------------------------------------------- | --------------------------------------------- |
| `-i FILE`          | Input `.dat` file                                         | ✅ Yes                                         |
| `--ip`             | Treat input as `geoip.dat` (IP → CIDR)                    | ✅ **One of `--ip` or `--site`**               |
| `--site`           | Treat input as `geosite.dat` (domains → rules)            | ✅ **One of `--ip` or `--site`**               |
| `-o FILE`          | Output file (`.json`, `.yaml`, or `.yml`)                 | ❌<br>(unless `--output-dir` or `--list-tags`) |
| `--output-dir DIR` | Export each tag/country to `DIR/{name}.{ext}`             | ❌                                             |
| `--format FMT`     | Force output format: `json` or `yaml`                     | ❌                                             |
| `--tag LIST`       | Comma-separated tags (e.g., `google,netflix`)             | ❌<br>(`--site` only)                          |
| `--country LIST`   | Comma-separated ISO 3166-1 alpha-2 codes (e.g., `US,DE`)  | ❌<br>(`--ip` only)                            |
| `--list-tags`      | Print all tags in `geosite.dat`/`geoip.dat` and exit      | ❌<br>(`--site` only)                          |
| `--sort`           | Sort keys alphabetically (countries/tags + domains/CIDRs) | ❌                                             |
| `-h`               | Show help                                                 | ❌                                             |

> ⚠️ **Notes**:
> - Use **either** `-o` **or** `--output-dir` — not both.
> - Country codes are **case-insensitive** (`us` = `US`).
> - Tags are **case-insensitive** (`GOOGLE` = `google`).

---

## 🧪 Examples

### 1. Inspect a Custom `geosite.dat`

```bash
# See what’s inside
./dat2json -i custom-geosite.dat --site --list-tags | grep -i "ad"

# Export ad-related tags
./dat2json -i custom-geosite.dat \
  --site \
  --output-dir ./ads \
  --tag=category-ads,ads-all \
  --sort
```

### 2. Build a Minimal GeoIP for EU

```bash
./dat2json -i geoip.dat \
  --ip \
  -o eu-countries.yaml \
  --country=DE,FR,IT,ES,NL,BE,AT,CH,SE,NO,DK,FI \
  --sort
```

### 3. Bulk Export All Countries (Parallel)

```bash
./dat2json -i geoip.dat --ip --output-dir ./countries --format json
# → Creates ./countries/US.json, ./countries/CN.json, etc.
```

### 4. Work with Protobuf Files (Mihomo Runtime)

```bash
# Mihomo's internal protobuf files have no header — use explicit mode
./dat2json -i mihomo-geoip.pb --ip -o countries.json
./dat2json -i mihomo-geosite.pb --site --output-dir ./rules
```

---

## 🛠 Technical Details

### File Format Support

| File          | Signature         | Content                                          |
| ------------- | ----------------- | ------------------------------------------------ |
| `geoip.dat`   | `GEOI` (optional) | `{ "US": ["1.2.3.0/24", ...], ... }`             |
| `geosite.dat` | `GEOS` (optional) | `{ "google": ["domain:.google.com", ...], ... }` |

> 💡 The tool **does not rely on signatures** — it uses the **explicit `--ip`/`--site` flag** to determine the parser.

### Output Format

- **JSON**: Standard indented JSON.
- **YAML**: Clean, human-readable YAML (uses `.yaml` extension by default; `.yml` accepted on input).

### Performance

- **Parsing**: Optimized for speed and memory efficiency.
- **Export**: Parallelized (up to 32 goroutines) when using `--output-dir`.
- **Progress**: Shown automatically for files with >10,000 entries.

---

## 🧪 Testing

Run unit tests:

```bash
go test ./...
```

> 🔒 Tests cover:
> - Binary and protobuf parsing
> - Format serialization
> - Filtering and sorting
> - Error handling

---

## 🤝 Contributing

Contributions are welcome! Please:

1. Fork the repo
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

---

## 📄 License

Distributed under the MIT License. See [LICENSE](LICENSE) for details.

---

## 🙏 Acknowledgements

- [V2Fly](https://github.com/v2fly) — for the original `.dat` format and community datasets
- [Mihomo (Clash Meta)](https://github.com/MetaCubeX/mihomo) — for the protobuf schema and reference implementation
- [YAML Spec](https://yaml.org) — for standardizing `.yaml` over `.yml`

---

> 💡 **Pro Tip**: Combine with [`v2fly/domain-list-community`](https://github.com/v2fly/domain-list-community) and [`v2fly/geoip`](https://github.com/v2fly/geoip) to rebuild `.dat` files after modification!
