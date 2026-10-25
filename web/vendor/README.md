# Artplayer distribution

Pinned version: **5.4.0**, MIT licensed, from the official npm release.

- `artplayer.js`: the original upstream browser distribution, copied without modification or additional minification.
- `artplayer.css`: the identical stylesheet exported as `Artplayer.STYLE`, extracted for local external loading under our CSP.
- `ARTPLAYER-LICENSE.txt`: the upstream license at the release commit.
- `artplayer-manifest.json`: release provenance, npm integrity and SHA-256 of the shipped assets.

These are third-party release artifacts. Project-owned source is maintained separately in the small adapter files; we do not compress our own code to meet the 250-line limit.

Assets are embedded in the Go binary and served locally. No CDN requests are needed. The external stylesheet link uses `id="artplayer-style"` so Artplayer reuses it instead of injecting an inline stylesheet. The adapter supplies inline-style-free SVGs for the check, playback-rate and custom-setting icons. The CSP permits only the SHA-256 hash of the upstream version link's `width:100%;` style attribute, even when the context menu is disabled. Scripts and style elements remain restricted to local resources; arbitrary inline styles and scripts stay blocked. The vendor bundle is unchanged.

To refresh a pinned release, install Node.js for the maintenance step and run `python3 scripts/update_artplayer.py 5.4.0`. Normal development and builds use the checked-in assets and require only Go.

## TMDB brand asset

`tmdb-logo.svg` is the unmodified TMDB blue square logo, obtained from the public copy at [Wikimedia Commons](https://commons.wikimedia.org/wiki/File:Tmdb.new.logo.svg). That page identifies the original TMDB asset source. It is used for provider attribution under [TMDB brand guidelines](https://www.themoviedb.org/about/logos-attribution); it is not covered by Artplayer's MIT license.
