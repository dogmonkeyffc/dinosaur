#!/usr/bin/env bash
set -euo pipefail

project_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$project_root"
target="${1:-all}"
if (($# > 1)) || [[ "$target" != mobile && "$target" != tv && "$target" != all ]]; then
  echo 'Usage: scripts/build_apps.sh [mobile|tv|all]' >&2
  exit 2
fi
tools=(java python3)
if [[ "$target" != mobile ]]; then tools+=(node npm); fi
for tool in "${tools[@]}"; do
  if ! command -v "$tool" >/dev/null; then
    echo "Android build requires $tool." >&2
    exit 1
  fi
done
if [[ -z "${ANDROID_HOME:-}" ]]; then
  echo 'Set ANDROID_HOME to the Android SDK directory.' >&2
  exit 1
fi

if [[ "$target" != mobile ]]; then
  npm ci --no-audit --no-fund
  npm run typecheck
fi
python3 scripts/check_file_lines.py
mkdir -p "$project_root/bin/apps"
for app in mobile tv; do
  if [[ "$target" != all && "$target" != "$app" ]]; then
    continue
  fi
  if [[ "$app" == mobile ]]; then
    : "${DINOSAUR_ANDROID_KEYSTORE:?Set DINOSAUR_ANDROID_KEYSTORE to the existing mobile signing key}"
    native="$project_root/apps/mobile-native"
    (
      cd "$native"
      ./gradlew :app:assembleRelease --no-daemon --max-workers="${DINOSAUR_ANDROID_BUILD_WORKERS:-2}"
    )
    version="$(awk '/^[[:space:]]*versionName / { gsub(/[^0-9.]/, "", $2); print $2; exit }' "$native/app/build.gradle")"
    output="$native/app/build/outputs/apk/release/app-release.apk"
  else
    (
    cd "$project_root/apps/$app"
    CI=1 node "$project_root/node_modules/expo/bin/cli" prebuild \
      --platform android --clean --no-install --template expo-template-bare-minimum@56.0.38
    cd android
    CI=1 NODE_ENV=production ./gradlew :app:assembleRelease --no-daemon \
      --max-workers="${DINOSAUR_ANDROID_BUILD_WORKERS:-2}" \
      -PreactNativeArchitectures="${DINOSAUR_ANDROID_ABIS:-arm64-v8a,x86_64}"
    )
    version="$(node -p "require('./apps/$app/app.json').expo.version")"
    output="$project_root/apps/$app/android/app/build/outputs/apk/release/app-release.apk"
  fi
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
