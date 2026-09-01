#!/bin/bash
# Usage: ./check.sh [goa version]   e.g. ./check.sh v3.30.0
set -u
ver="${1:-v3.31.0-preview.3}"
go get "goa.design/goa/v3@${ver}" >/dev/null 2>&1 && go mod tidy >/dev/null 2>&1
echo "goa ${ver}"
backup=$(cat design/design.go)
run() {
  cp "variants/$1.go" design/design.go
  rm -rf gen
  out=$(go run goa.design/goa/v3/cmd/goa gen repro/design -o . 2>&1)
  st=$?
  rm -rf gen
  if [ $st -eq 0 ]; then
    echo "  [OK]     $2"
  elif echo "$out" | grep -q "^panic"; then
    echo "  [PANIC]  $2"
  else
    echo "  [FAILED] $2: $(echo "$out" | grep -vE '^exit status' | head -1)"
  fi
}
run untagged       "two success responses, no Tag"
run named_body     "two success responses, Tag + Body(attribute)"
run tagged_no_body "two success responses, Tag, no Body"
run single_response "single success response"
printf '%s' "$backup" > design/design.go
