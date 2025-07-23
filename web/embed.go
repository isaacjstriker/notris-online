package web

import "embed"

//go:embed all:static all:templates
var Files embed.FS
