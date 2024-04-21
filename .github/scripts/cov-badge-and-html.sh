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

curl -s "https://img.shields.io/badge/acc%20coverage-$COVERAGE%25-$BADGE_COLOR" > _site/badge.svg
