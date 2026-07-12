#!/usr/bin/env bash

# Strict bash mode
# https://gist.github.com/mohanpedala/1e2ff5661761d3abd0385e8223e16425
set -euo pipefail

if ! test -f cover.out; then
  echo "coverage file does not exist"
  exit 1
fi

mkdir -p _site

go tool cover -html cover.out -o _site/cover.html

COVERAGE=$(go tool cover -func=cover.out | grep total: | grep -Eo '[0-9]+\.[0-9]+')

BADGE_COLOR=yellow
if (( $(echo "$COVERAGE <= 50" | bc -l) )); then
  BADGE_COLOR=red
elif (( $(echo "$COVERAGE > 80" | bc -l) )); then
  BADGE_COLOR=green
fi

MAGNIFYING_GLASS="data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHZpZXdCb3g9IjAgMCAyNCAyNCIgZmlsbD0id2hpdGUiPjxwYXRoIGZpbGwtcnVsZT0iZXZlbm9kZCIgZD0iTTkuNSAyYTcuNSA3LjUgMCAxIDAgNC41NSAxMy40Nmw1LjI0IDUuMjRhMS40IDEuNCAwIDAgMCAxLjk4LTEuOThsLTUuMjQtNS4yNEE3LjUgNy41IDAgMCAwIDkuNSAyem0wIDNhNC41IDQuNSAwIDEgMSAwIDkgNC41IDQuNSAwIDAgMSAwLTl6Ii8+PC9zdmc+"

curl -s "https://img.shields.io/badge/acc%20coverage-$COVERAGE%25-$BADGE_COLOR?logo=$MAGNIFYING_GLASS" > _site/badge.svg
