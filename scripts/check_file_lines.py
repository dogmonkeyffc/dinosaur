"""Enforce the source file line limit without counting generated/runtime files."""
from pathlib import Path
import sys

extensions = {'.go', '.js', '.cjs', '.mjs', '.jsx', '.ts', '.tsx', '.css', '.html', '.py', '.sql', '.sh'}
roots = ['cmd', 'internal', 'web', 'ui/src', 'ui/public', 'scripts', 'deploy']
generated = [Path('web/dist'), Path('ui/public/assets/vendor')]
violations = []
paths = [Path('build.sh'), Path('ui/vite.config.ts'), Path('ui/index.html')]
for root in roots:
    paths.extend(Path(root).rglob('*'))
for path in paths:
    if any(path.is_relative_to(folder) for folder in generated):
        continue
    if path.is_file() and path.suffix in extensions:
        count = len(path.read_text().splitlines())
        if count > 250:
            violations.append(f'{path}: {count} lines (maximum 250)')
if violations:
    print('\n'.join(violations))
    sys.exit(1)
print('All source files are within the 250-line limit.')
