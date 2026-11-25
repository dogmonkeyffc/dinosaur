#!/usr/bin/env bash
set -euo pipefail

project_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$project_root"
target="${1:-all}"
if (($# > 1)) || [[ "$target" != mobile && "$target" != tv && "$target" != all ]]; then
  echo 'Usage: scripts/build_apps.sh [mobile|tv|all]' >&2
  exit 2
fi
for tool in java python3; do
  if ! command -v "$tool" >/dev/null; then
    echo "Android build requires $tool." >&2
    exit 1
  fi
done
if [[ -z "${ANDROID_HOME:-}" ]]; then
  echo 'Set ANDROID_HOME to the Android SDK directory.' >&2
  exit 1
fi

python3 scripts/check_file_lines.py
mkdir -p "$project_root/bin/apps"
native="$project_root/apps/mobile-native"
for app in mobile tv; do
  if [[ "$target" != all && "$target" != "$app" ]]; then
    continue
  fi
  if [[ "$app" == mobile ]]; then
    : "${DINOSAUR_ANDROID_KEYSTORE:?Set DINOSAUR_ANDROID_KEYSTORE to the existing mobile signing key}"
    module=app
    source="$native/app"
  else
    : "${DINOSAUR_TV_KEYSTORE:?Set DINOSAUR_TV_KEYSTORE to the existing TV signing key}"
    module=tv
    source="$project_root/apps/tv-native"
  fi
  (
    cd "$native"
    ./gradlew ":${module}:assembleRelease" --no-daemon --max-workers="${DINOSAUR_ANDROID_BUILD_WORKERS:-2}"
  )
  version="$(awk '/^[[:space:]]*versionName / { gsub(/[^0-9.]/, "", $2); print $2; exit }' "$source/build.gradle")"
  output="$source/build/outputs/apk/release/$module-release.apk"
  artifact="$project_root/bin/apps/dinosaur-$app-$version-preview.apk"
  cp "$output" "$artifact"
  python3 - "$artifact" <<'PY'
import hashlib
from pathlib import Path
import sys
path = Path(sys.argv[1])
digest = hashlib.sha256(path.read_bytes()).hexdigest()
Path(str(path) + '.sha256').write_text(f'{digest}  {path.name}\n')
PY
  echo "Built $artifact"
done
