#!/usr/bin/env bash
set -euo pipefail

project_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
cd "$project_root"
output="$project_root/bin/dinosaur"
frontend_only=false
while (($#)); do
  case "$1" in
    --frontend-only) frontend_only=true; shift ;;
    --output)
      if (($# < 2)); then
        echo 'Usage: ./build.sh [--frontend-only] [--output FILE]' >&2
        exit 2
      fi
      output="$2"
      shift 2
      ;;
    *) echo "Unknown build option: $1" >&2; exit 2 ;;
  esac
done

for tool in node npm python3; do
  if ! command -v "$tool" >/dev/null; then
    echo "Build requires $tool; see ui/package.json for the Node.js version." >&2
    exit 1
  fi
done
if ! "$frontend_only" && ! command -v go >/dev/null; then
  echo 'Build requires Go on PATH.' >&2
  exit 1
fi

npm --prefix ui ci --include=dev --no-audit --no-fund
npm --prefix ui run build
touch web/dist/.gitkeep
python3 scripts/check_file_lines.py
if "$frontend_only"; then
  exit 0
fi
mkdir -p -- "$(dirname -- "$output")"
go build -trimpath -o "$output" ./cmd/dinosaur
echo "Built $output"
