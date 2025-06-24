package ui

import "embed"

//go:embed html/* html/partials/* html/pages/* static/css/* static/js/* static/img/*
var Files embed.FS
