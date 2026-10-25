"""Vendor an explicit official Artplayer release into the React public assets."""
import base64
import hashlib
import io
import json
from pathlib import Path
import subprocess
import sys
import tarfile
import urllib.request

version = sys.argv[1] if len(sys.argv) == 2 else ''
if not version or any(c not in '0123456789.' for c in version):
    raise SystemExit('Usage: python3 scripts/update_artplayer.py <exact-version>')
root = Path(__file__).resolve().parent.parent
folder = root / 'ui' / 'public' / 'assets' / 'vendor'
folder.mkdir(parents=True, exist_ok=True)
with urllib.request.urlopen(f'https://registry.npmjs.org/artplayer/{version}', timeout=30) as response:
    metadata = json.load(response)
with urllib.request.urlopen(metadata['dist']['tarball'], timeout=60) as response:
    archive = response.read()
integrity = metadata['dist']['integrity']
algorithm, expected = integrity.split('-', 1)
actual = base64.b64encode(hashlib.new(algorithm, archive).digest()).decode()
if actual != expected:
    raise SystemExit('npm archive integrity check failed')
with tarfile.open(fileobj=io.BytesIO(archive), mode='r:gz') as tar:
    member = tar.getmember('package/dist/artplayer.js')
    (folder / 'artplayer.js').write_bytes(tar.extractfile(member).read())
commit = metadata['gitHead']
license_url = f'https://raw.githubusercontent.com/zhw2590582/ArtPlayer/{commit}/LICENSE'
with urllib.request.urlopen(license_url, timeout=30) as response:
    (folder / 'ARTPLAYER-LICENSE.txt').write_bytes(response.read())
subprocess.run(['node', str(root / 'scripts' / 'extract_artplayer_css.cjs')], check=True)
manifest = {
    'name': 'artplayer', 'version': version, 'license': 'MIT',
    'repository': 'https://github.com/zhw2590582/ArtPlayer', 'commit': commit,
    'tarball': metadata['dist']['tarball'], 'integrity': integrity,
    'files': {
        name: hashlib.sha256((folder / name).read_bytes()).hexdigest()
        for name in ['artplayer.js', 'artplayer.css']
    },
}
(folder / 'artplayer-manifest.json').write_text(json.dumps(manifest, indent=2) + '\n')
print(f'Vendored Artplayer {version}')
