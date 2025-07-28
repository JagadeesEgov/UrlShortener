package migration

import (
	"database/sql"
	"log"
)

func RunMigrations(db *sql.DB) {
	query := `
	CREATE TABLE IF NOT EXISTS eg_url_shortener (
		short_key VARCHAR(4) PRIMARY KEY,
		url VARCHAR(2048) NOT NULL UNIQUE,
		validfrom BIGINT,
		validto BIGINT
	);
	`
	_, err := db.Exec(query)
	if err != nil {
		log.Fatalf("Failed to run migration: %v", err)
	}
	log.Println("✅ Migration completed successfully.")
}
