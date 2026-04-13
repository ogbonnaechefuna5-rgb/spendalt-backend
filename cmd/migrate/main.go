package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/spendalt/backend/config"
	"github.com/spendalt/backend/internal/migrations"

	// Blank imports trigger each domain's init(), registering its migrations.
	// Add new domains here in dependency order.
	_ "github.com/spendalt/backend/internal/user"
	_ "github.com/spendalt/backend/internal/category"
	_ "github.com/spendalt/backend/internal/transaction"
	_ "github.com/spendalt/backend/internal/budget"
	_ "github.com/spendalt/backend/internal/savings"
	_ "github.com/spendalt/backend/internal/auth"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: migrate [up|down|status|reset]")
	}
	command := os.Args[1]

	cfg := config.Load()
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatal("failed to open db:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("failed to connect to db:", err)
	}

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatal(err)
	}

	switch command {
	case "up":
		runEach(db, "up")
	case "down":
		// Roll back one step in each domain, reverse order
		for i := len(migrations.Ordered) - 1; i >= 0; i-- {
			runDomain(db, migrations.Ordered[i], "down")
		}
	case "reset":
		for i := len(migrations.Ordered) - 1; i >= 0; i-- {
			resetDomain(db, migrations.Ordered[i])
		}
		runEach(db, "up")
	case "status":
		runEach(db, "status")
	default:
		log.Fatalf("unknown command %q — use up, down, status, or reset", command)
	}
}

func runEach(db *sql.DB, command string) {
	for _, d := range migrations.Ordered {
		runDomain(db, d, command)
	}
}

func runDomain(db *sql.DB, d migrations.Domain, command string) {
	// Each domain gets its own version tracking table: goose_db_version_<name>
	goose.SetTableName("goose_db_version_" + d.Name)
	goose.SetBaseFS(d.FS)
	fmt.Printf("── %s (%s)\n", d.Name, command)
	if err := goose.Run(command, db, "migrations"); err != nil {
		log.Fatalf("%s %s: %v", command, d.Name, err)
	}
}

// resetDomain rolls back all migrations for a single domain to version 0.
func resetDomain(db *sql.DB, d migrations.Domain) {
	goose.SetTableName("goose_db_version_" + d.Name)
	goose.SetBaseFS(d.FS)
	fmt.Printf("── %s (reset)\n", d.Name)
	if err := goose.DownTo(db, "migrations", 0); err != nil {
		log.Fatalf("reset %s: %v", d.Name, err)
	}
}
