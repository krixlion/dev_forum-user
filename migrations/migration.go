// This package is here only to allow other packages to use its
// embeded FS without worrying about directing to parent directories
// and trying to find workarounds since relative paths are not allowed.
package migrations

import "embed"

//go:embed *.sql
var EmbedPath embed.FS
