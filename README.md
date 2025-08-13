# goprobe

A high-performance TCP port scanner and probe tool written in Go.

## Overview

`goprobe` is a CLI utility designed to efficiently check TCP port availability for a list of hosts. It demonstrates practical knowledge of TCP/IP networking, concurrency, error handling, and Go best practices. The project is structured for extensibility, reliability, and testability, making it suitable for both production use and technical interviews.

## Features

-   **TCP/IP Fundamentals:** Uses Go's `net` package to perform low-level TCP port checks.
-   **Concurrency:** Scans multiple hosts and ports in parallel for speed and efficiency.
-   **Customizable:** Accepts host lists and port lists via CLI flags and files.
-   **Output Options:** Results can be printed to stdout, or saved as CSV/JSON reports.
-   **Timeout Control:** Configurable connection timeouts for robust scanning.
-   **Comprehensive Testing:** Includes unit, property-based, fuzz, and benchmark tests for reliability.
-   **Automation:** Makefile for easy testing, coverage, and benchmarking.

## Usage

### Basic Scan

```sh
goprobe --hosts hosts.txt --ports=22,80,443
```

### Save Results

```sh
goprobe --hosts hosts.txt --csv results.csv --json results.json
```

### Custom Timeout

```sh
goprobe --hosts hosts.txt --ports=8080 --timeout 500ms
```

## Example Output

```
HOSTNAME       PORT         STATUS
test.com       22           open
test.com       80           open
example.org    22           closed
example.org    80           open
```

## Installation

```sh
git clone https://github.com/n0sh4d3/goprobe.git
cd goprobe
go build -o goprobe
```

## Testing & Reliability

-   **Run all tests:**
    ```sh
    make all
    ```
-   **Run fuzz tests:**
    ```sh
    make fuzz-all
    ```
-   **View coverage:**
    ```sh
    make coverage
    ```

## Project Structure

-   `main.go` — CLI logic and entry point
-   `tcpCon/` — TCP connection scanner implementation
-   `main_test.go`, `tcpCon/tcpCon_test.go` — Comprehensive test suites
-   `Makefile` — Automation for testing, coverage, and benchmarking
-   `testdata/` — Sample data for testing

## Technical Highlights

-   **TCP/IP:** Uses `net.DialTimeout` for port probing, demonstrating understanding of TCP handshakes and timeouts.
-   **Concurrency:** Utilizes goroutines and sync primitives for parallel scanning.
-   **Error Handling:** Gracefully handles unreachable hosts, closed ports, and invalid input.
-   **Extensibility:** Easily add new output formats or scanning strategies.
-   **Testing:** Fuzz, benchmark, and property-based tests ensure reliability and robustness.

## Why This Project?

This project showcases:

-   Practical TCP/IP and networking skills
-   Idiomatic Go code and concurrency
-   Automated testing and CI readiness
-   Real-world error handling and reporting

## License

MIT

---

_Created for technical interviews and as a demonstration of Go, TCP/IP, and software engineering best practices._
