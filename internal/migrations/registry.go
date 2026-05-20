// Package migrations is the single source of truth for migration order.
// Each domain embeds its own SQL files; this package collects them in
// dependency order for the migrate command to execute sequentially.
package migrations

import (
	"database/sql"
	"fmt"
	"io/fs"
	"strings"

	"github.com/pressly/goose/v3"
)

// Domain pairs a name with its embedded migration filesystem.
type Domain struct {
	Name string
	FS   fs.FS
}

// Ordered defines the execution order — dependencies before dependents.
// To add a new domain: embed its migrations in its own package and append here.
var Ordered []Domain

// RunUp applies all pending migrations for every registered domain.
func RunUp(db *sql.DB) error {
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	for _, d := range Ordered {
		goose.SetTableName("goose_db_version_" + d.Name)
		goose.SetBaseFS(d.FS)
		fmt.Printf("── migrate %s\n", d.Name)
		if err := goose.Up(db, "migrations"); err != nil && !strings.Contains(err.Error(), "no next version found") {
			return fmt.Errorf("migrate %s: %w", d.Name, err)
		}
	}
	return nil
}
