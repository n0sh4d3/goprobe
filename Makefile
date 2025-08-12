FUZZ_TESTS := $(shell ./scripts/list_fuzz_tests.sh)

fuzz-all:
	@for fuzz in $(FUZZ_TESTS); do \
		echo "Running $$fuzz for 1 minute..."; \
		go test -fuzz=$$fuzz -fuzztime=1m || exit 1; \
	done

test:
	go test -v ./...

race:
	go test -race ./...

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out

bench:
	go test -bench=. ./...

fmt:
	go fmt *.go

vet:
	go vet ./...

all: vet test race coverage bench fuzz-all