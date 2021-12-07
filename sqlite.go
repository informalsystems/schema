package schema

import (
	"fmt"
	"strings"
	"time"
	"unicode"
)

// SQLite is the dialect for sqlite3 databases
var SQLite = &sqliteDialect{}

type sqliteDialect struct{}

// CreateSQL takes the name of the migration tracking table and
// returns the SQL statement needed to create it
func (s sqliteDialect) CreateSQL(tableName string) string {
	return fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id TEXT NOT NULL,
			checksum TEXT NOT NULL DEFAULT '',
			execution_time_in_millis INTEGER NOT NULL DEFAULT 0,
			applied_at DATETIME
		);`, tableName)
}

// InsertSQL takes the name of the migration tracking table and
// returns the SQL statement needed to insert a migration into it
func (s *sqliteDialect) InsertSQL(tableName string) string {
	return fmt.Sprintf(`
		INSERT INTO %s
		( id, checksum, execution_time_in_millis, applied_at )
		VALUES
		( ?, ?, ?, ? )
		`, tableName)
}

// GetAppliedMigrations retrieves all data from the migrations tracking table
//
func (s sqliteDialect) GetAppliedMigrations(tx Queryer, tableName string) (migrations []*AppliedMigration, err error) {
	migrations = make([]*AppliedMigration, 0)

	query := fmt.Sprintf(`
		SELECT id, checksum, execution_time_in_millis, applied_at
		FROM %s
		ORDER BY id ASC
	`, tableName)
	rows, err := tx.Query(query)
	if err != nil {
		return migrations, err
	}
	defer rows.Close()

	for rows.Next() {
		migration := AppliedMigration{}
		err = rows.Scan(&migration.ID, &migration.Checksum, &migration.ExecutionTimeInMillis, &migration.AppliedAt)
		if err != nil {
			err = fmt.Errorf("Failed to GetAppliedMigrations. Did somebody change the structure of the %s table?: %w", tableName, err)
			return migrations, err
		}
		migration.AppliedAt = migration.AppliedAt.In(time.Local)
		migrations = append(migrations, &migration)
	}

	return migrations, err
}

// QuotedTableName returns the string value of the name of the migration
// tracking table after it has been quoted for SQLite
func (s sqliteDialect) QuotedTableName(schemaName, tableName string) string {
	ident := schemaName + tableName
	if ident == "" {
		return ""
	}

	var sb strings.Builder
	sb.WriteRune('"')
	for _, r := range ident {
		switch {
		case unicode.IsSpace(r):
			// Skip spaces
			continue
		case r == '"':
			// Escape double-quotes with repeated double-quotes
			sb.WriteString(`""`)
		case r == ';':
			// Ignore the command termination character
			continue
		default:
			sb.WriteRune(r)
		}
	}
	sb.WriteRune('"')
	return sb.String()

}
