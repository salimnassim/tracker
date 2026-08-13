package tracker

import "embed"

//go:embed templates/*.html templates/components/*.html
var Templates embed.FS

//go:embed static/*.css static/*.js
var Static embed.FS
