package savings

import (
	"embed"

	"github.com/spendalt/backend/internal/migrations"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

func init() {
	migrations.Ordered = append(migrations.Ordered, migrations.Domain{
		Name: "savings",
		FS:   migrationFiles,
	})
}
