"""Install a native release for the current user and start its systemd service."""
import hashlib
import json
import os
from pathlib import Path
import re
import shutil
import subprocess
import time
import urllib.error
import urllib.request
import uuid


def run(*args):
    subprocess.run(args, check=True)


def verify_package(source):
    manifest = json.loads((source / 'manifest.json').read_text())
    for name, expected in manifest['files'].items():
        path = source / name
        if Path(name).is_absolute() or '..' in Path(name).parts:
            raise SystemExit('Unsafe manifest entry')
        if hashlib.sha256(path.read_bytes()).hexdigest() != expected:
            raise SystemExit('Package verification failed: ' + name)
    if manifest['platform'] != 'linux-amd64' or os.uname().machine != 'x86_64':
        raise SystemExit('This package requires Linux x86_64')
    if not re.fullmatch(r'[A-Za-z0-9_.-]+', manifest['version']):
        raise SystemExit('Invalid release version')
    for name in ('ffmpeg', 'ffprobe'):
        subprocess.run([str(source / name), '-version'], stdout=subprocess.DEVNULL, check=True)
    help_text = subprocess.check_output([str(source / 'ffmpeg'), '-hide_banner', '-h', 'full'],
                                        stderr=subprocess.STDOUT, text=True)
    encoders = subprocess.check_output([str(source / 'ffmpeg'), '-hide_banner', '-encoders'], text=True)
    if '-readrate_initial_burst' not in help_text or 'libx264' not in encoders or ' aac ' not in encoders:
        raise SystemExit('Bundled FFmpeg lacks required options or encoders')
    return manifest


def activate(target, link):
    temporary = link.with_name('.current-' + uuid.uuid4().hex)
    temporary.symlink_to(target)
    temporary.replace(link)


def main():
    if os.geteuid() == 0:
        raise SystemExit('Run as the normal NAS user, without sudo')
    source = Path(__file__).resolve().parent
    manifest = verify_package(source)
    home = Path.home()
    application = home / '.local' / 'share' / 'dinosaur'
    data = home / '.cache' / 'dinosaur'
    units = home / '.config' / 'systemd' / 'user'
    release = application / 'releases' / manifest['version']
    current = application / 'current'
    previous = current.resolve() if current.is_symlink() else None
    application.mkdir(parents=True, exist_ok=True, mode=0o700)
    data.mkdir(parents=True, exist_ok=True, mode=0o700)
    units.mkdir(parents=True, exist_ok=True)
    if release.exists():
        verify_package(release)
    else:
        shutil.copytree(source, release)
    environment = application / 'dinosaur.env'
    if not environment.exists():
        text = (source / 'dinosaur.env.example').read_text().replace('@HOME@', str(home))
        descriptor = os.open(environment, os.O_WRONLY | os.O_CREAT | os.O_EXCL, 0o600)
        with os.fdopen(descriptor, 'w') as stream:
            stream.write(text)
    unit = units / 'dinosaur.service'
    shutil.copy2(source / 'dinosaur.service', unit)
    activate(release, current)
    run('systemctl', '--user', 'daemon-reload')
    run('systemctl', '--user', 'enable', 'dinosaur.service')
    run('systemctl', '--user', 'restart', 'dinosaur.service')
    values = {}
    for line in environment.read_text().splitlines():
        if '=' in line and not line.lstrip().startswith('#'):
            key, value = line.split('=', 1)
            values[key.strip()] = value.strip().strip('"')
    port = values.get('DINOSAUR_ADDR', '0.0.0.0:18096').rsplit(':', 1)[1]
    url = 'http://127.0.0.1:' + port + '/health/ready'
    for _ in range(30):
        try:
            with urllib.request.urlopen(url, timeout=1) as response:
                if response.status == 200:
                    print('Installed release: ' + manifest['version'])
                    print('Healthy on port ' + port)
                    print('Data directory: ' + str(data))
                    return
        except (OSError, urllib.error.URLError):
            pass
        time.sleep(0.5)
    if previous is not None:
        activate(previous, current)
        run('systemctl', '--user', 'restart', 'dinosaur.service')
    else:
        run('systemctl', '--user', 'disable', '--now', 'dinosaur.service')
    raise SystemExit('Health check failed; inspect journalctl --user -u dinosaur.service')


if __name__ == '__main__':
    main()
