package migrations

import "embed"

// This package is here only to allow other packages to use this
// embeded FS without worrying about directing to parent directories
// and trying to find workarounds since relative paths are not allowed.

//go:embed *.sql
var EmbedPath embed.FS
