package testutil

import (
	"database/sql"
	"embed"
	"testing"

	"repub/internal/db/sqlite"
)

//go:embed schema_sqlite.sql
var schemaFS embed.FS

func SetupSQLiteDB(t *testing.T) (*sql.DB, *sqlite.Queries) {
	sqliteDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open SQLite database: %v", err)
	}

	// Read schema from embedded file
	schema, err := schemaFS.ReadFile("schema_sqlite.sql")
	if err != nil {
		t.Fatalf("Failed to read schema: %v", err)
	}

	if _, err := sqliteDB.Exec(string(schema)); err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}

	queries := sqlite.New(sqliteDB)
	return sqliteDB, queries
}