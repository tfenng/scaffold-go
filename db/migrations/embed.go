// Package migrations embeds the schema migration files so they can be applied
// without relying on an external CLI or shipping SQL files alongside the binary.
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
