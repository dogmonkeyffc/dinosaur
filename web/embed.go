// Package web embeds the React production build created by build.sh.
package web

import "embed"

//go:embed all:dist
var Files embed.FS
