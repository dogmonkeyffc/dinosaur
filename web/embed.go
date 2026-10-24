// Package web contains the embedded browser interface.
package web

import "embed"

//go:embed *.html *.css *.js vendor/*
var Files embed.FS
