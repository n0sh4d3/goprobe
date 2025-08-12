#!/bin/sh
grep -E '^func[[:space:]]+Fuzz_' main_test.go | awk '{print $2}' | sed 's/(.*//'

