// Extract the upstream stylesheet without modifying or minifying the release.
const fs = require('node:fs');
const path = require('node:path');
const folder = path.resolve(__dirname, '../ui/public/assets/vendor');
const Artplayer = require(path.join(folder, 'artplayer.js'));
fs.writeFileSync(path.join(folder, 'artplayer.css'), Artplayer.STYLE);
