"""Build a Linux/amd64 native release with a verified FFmpeg archive."""
import argparse
from datetime import datetime, timezone
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess
import tarfile
import tempfile


def sha256(path):
    digest = hashlib.sha256()
    with path.open('rb') as stream:
        for chunk in iter(lambda: stream.read(1024 * 1024), b''):
            digest.update(chunk)
    return digest.hexdigest()


def copy_ffmpeg(archive, destination):
    copied = set()
    with tarfile.open(archive, 'r:*') as source:
        for member in source:
            if not member.isfile():
                continue
            path = Path(member.name)
            if path.is_absolute() or '..' in path.parts:
                raise ValueError('Unsafe archive member')
            if path.name in ('ffmpeg', 'ffprobe') and path.parent.name == 'bin':
                target = destination / path.name
                copied.add(path.name)
            elif path.name.lower().startswith(('license', 'copying', 'readme')):
                target = destination / 'licenses' / 'ffmpeg' / Path(*path.parts[1:])
            else:
                continue
            target.parent.mkdir(parents=True, exist_ok=True)
            with source.extractfile(member) as input_file, target.open('wb') as output:
                shutil.copyfileobj(input_file, output)
            target.chmod(0o755 if path.name in ('ffmpeg', 'ffprobe') else 0o644)
    if copied != {'ffmpeg', 'ffprobe'}:
        raise ValueError('Archive must include bin/ffmpeg and bin/ffprobe')


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--ffmpeg-archive', type=Path, required=True)
    parser.add_argument('--ffmpeg-sha256', required=True)
    parser.add_argument('--ffmpeg-source', default='')
    args = parser.parse_args()
    if sha256(args.ffmpeg_archive) != args.ffmpeg_sha256.lower():
        raise SystemExit('FFmpeg archive SHA-256 mismatch')
    root = Path(__file__).resolve().parent.parent
    revision = subprocess.check_output(['git', 'rev-parse', '--short', 'HEAD'], cwd=root, text=True).strip()
    dirty = subprocess.check_output(['git', 'diff', '--name-only', 'HEAD'], cwd=root, text=True).strip()
    stamp = datetime.now(timezone.utc).strftime('%Y%m%dT%H%M%SZ')
    version = revision + ('-dirty' if dirty else '') + '-' + stamp
    output = root / 'bin' / 'native'
    output.mkdir(parents=True, exist_ok=True)
    archive = output / ('dinosaur-' + version + '-linux-amd64.tar.gz')
    with tempfile.TemporaryDirectory(prefix='dinosaur-package-') as temporary:
        stage = Path(temporary) / ('dinosaur-' + version)
        stage.mkdir()
        environment = dict(os.environ, CGO_ENABLED='0', GOOS='linux', GOARCH='amd64')
        subprocess.run(['go', 'build', '-trimpath', '-o', str(stage / 'dinosaur'), './cmd/dinosaur'],
                       cwd=root, env=environment, check=True)
        copy_ffmpeg(args.ffmpeg_archive, stage)
        for name in ('dinosaur.service', 'dinosaur.env.example', 'install.py'):
            shutil.copy2(root / 'deploy' / 'native' / name, stage / name)
        (stage / 'VERSION').write_text(version + '\n')
        manifest = {
            'version': version, 'platform': 'linux-amd64',
            'ffmpeg_archive_sha256': args.ffmpeg_sha256.lower(),
            'ffmpeg_source': args.ffmpeg_source,
            'files': {str(path.relative_to(stage)): sha256(path) for path in stage.rglob('*') if path.is_file()},
        }
        (stage / 'manifest.json').write_text(json.dumps(manifest, indent=2) + '\n')
        with tarfile.open(archive, 'w:gz') as target:
            target.add(stage, arcname=stage.name)
    archive.with_suffix(archive.suffix + '.sha256').write_text(sha256(archive) + '  ' + archive.name + '\n')
    print(archive)


if __name__ == '__main__':
    main()
