"""Enforce the source file line limit without counting generated/runtime files."""
from pathlib import Path
import sys

extensions = {'.go', '.js', '.cjs', '.mjs', '.jsx', '.ts', '.tsx', '.css', '.html', '.py', '.sql'}
roots = ['cmd', 'internal', 'web', 'scripts', 'deploy']
violations = []
for root in roots:
    for path in Path(root).rglob('*'):
        if path.is_file() and path.suffix in extensions:
            count = len(path.read_text().splitlines())
            if count > 250:
                violations.append(f'{path}: {count} lines (maximum 250)')
if violations:
    print('\n'.join(violations))
    sys.exit(1)
print('All source files are within the 250-line limit.')
