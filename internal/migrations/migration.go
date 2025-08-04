package migration

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func RunMigrations(db *sql.DB) {
	// Only apply files that start with "001", "002", ... (ordered execution)
	files, err := filepath.Glob("db/migrations/*.sql")
	if err != nil {
		log.Fatalf("❌ Failed to list migration files: %v", err)
	}

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			log.Fatalf("❌ Failed to read migration file %s: %v", file, err)
		}

		if _, err := db.Exec(string(content)); err != nil {
			log.Fatalf("❌ Failed to execute migration %s: %v", file, err)
		}

		fmt.Printf("✅ Migration applied: %s\n", file)
	}
}
