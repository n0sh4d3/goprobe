# goprobe

A fast, flexible TCP port scanner written in Go.  
Supports CSV, JSON, and colored table output, with robust CLI flags for easy use.

## Features

-   Scan multiple hosts and ports from files or CLI flags
-   Output results as a colored table, CSV, or JSON
-   Flexible output: print to stdout or write to files
-   User-friendly CLI with clear help and examples
-   Comprehensive error handling and notifications
-   Extensive test coverage (unit, fuzz, benchmark)

## Installation

Clone and build:

```sh
git clone https://github.com/n0sh4d3/goprobe.git
cd goprobe
go build
```

## Usage

### Quick Start

Scan hosts and ports from files, print results as a table:

```sh
go run . --hosts test.txt --ports 22,80,443 --stdout
```

### Output Options

-   `--stdout`  
    Print results to the terminal as a colored table (default if no output flags are given).

-   `--csv [filename]`  
    Write results as CSV to the specified file. If no filename is given, defaults to `goprobe.csv`.

-   `--json [filename]`  
    Write results as JSON to the specified file. If no filename is given, defaults to `goprobe.json`.

You can combine output flags to print and save results at the same time:

```sh
go run . --hosts test.txt --ports=22,80,443 --stdout --csv results.csv --json results.json
```

### Example

```sh
go run . --hosts test.txt --ports 22,80,443 --stdout
go run . --hosts test.txt --ports 22,80,443 --csv
go run . --hosts test.txt --ports 22,80,443 --json
go run . --hosts test.txt --ports 22,80,443 --stdout --csv --json
```

### Flags

-   `--hosts <file>`: File with hostnames/IPs (one per line)
-   `--ports=<list>`: Comma-separated list of ports (e.g. `22,80,443`)
-   `--timeout <ms>`: Timeout per connection (default: 1000ms)
-   `--csv [file]`: Write CSV report (default: `goprobe.csv`)
-   `--json [file]`: Write JSON report (default: `goprobe.json`)
-   `--stdout`: Print results to terminal as a colored table

## Output Formats

**Table (stdout):**

```
hostname        port    status
--------------- ------- -------
example.com     22      open
example.com     80      closed
```

**CSV:**

```
hostname,port,status
example.com,22,open
example.com,80,closed
```

**JSON:**

```json
[
	{ "hostname": "example.com", "port": 22, "status": "open" },
	{ "hostname": "example.com", "port": 80, "status": "closed" }
]
```

## Notifications

When writing to files, you'll see info messages like:

```
[INFO] CSV file created: results.csv
[INFO] JSON file created: results.json
```

## Testing

Run all tests (unit, fuzz, benchmark):

```sh
make test
```

## License

MIT

---

**Interview-ready:**  
This project demonstrates advanced Go skills: concurrency, TCP/IP, CLI design, error handling, output formatting, and comprehensive testing.
