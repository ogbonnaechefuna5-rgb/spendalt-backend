// Package migrations is the single source of truth for migration order.
// Each domain embeds its own SQL files; this package collects them in
// dependency order for the migrate command to execute sequentially.
package migrations

import "io/fs"

// Domain pairs a name with its embedded migration filesystem.
type Domain struct {
	Name string
	FS   fs.FS
}

// Ordered defines the execution order — dependencies before dependents.
// To add a new domain: embed its migrations in its own package and append here.
var Ordered []Domain
