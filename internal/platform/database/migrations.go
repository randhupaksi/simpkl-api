package database

import (
	"context"
	"database/sql"
	"fmt"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	migrationassets "simpkl-api"
)

const migrationLockName = "simpkl:database:migrations"

var migrationFilename = regexp.MustCompile(`^(\d+)_.*\.up\.sql$`)

type migrationFile struct {
	Version int
	Name    string
}

// RunMigrations applies every embedded up migration newer than the database
// version. MySQL advisory locking prevents two API replicas from migrating at
// the same time. A dirty migration fails closed so a partially applied schema
// cannot be mistaken for a healthy application.
func RunMigrations(ctx context.Context, db *sql.DB) error {
	lockCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var acquired int
	if err := db.QueryRowContext(lockCtx, "SELECT GET_LOCK(?, 30)", migrationLockName).Scan(&acquired); err != nil {
		return fmt.Errorf("acquire migration lock: %w", err)
	}
	if acquired != 1 {
		return fmt.Errorf("migration lock was not acquired")
	}
	defer func() {
		_, _ = db.ExecContext(context.Background(), "SELECT RELEASE_LOCK(?)", migrationLockName)
	}()

	if err := ensureMigrationTable(ctx, db); err != nil {
		return err
	}
	currentVersion, dirty, err := currentMigration(ctx, db)
	if err != nil {
		return err
	}
	if dirty {
		return fmt.Errorf("database migration is marked dirty at version %d; repair the database before starting the API", currentVersion)
	}
	if currentVersion == 0 {
		baseline, baselineErr := detectExistingSchemaBaseline(ctx, db)
		if baselineErr != nil {
			return baselineErr
		}
		if baseline > 0 {
			if _, err := db.ExecContext(ctx, "UPDATE schema_migrations SET version = ?, dirty = FALSE", baseline); err != nil {
				return fmt.Errorf("record existing schema baseline %d: %w", baseline, err)
			}
			currentVersion = baseline
		}
	}

	migrations, err := embeddedMigrations()
	if err != nil {
		return err
	}
	for _, migration := range migrations {
		if migration.Version <= currentVersion {
			continue
		}
		if err := applyMigration(ctx, db, migration); err != nil {
			return err
		}
		currentVersion = migration.Version
	}
	return nil
}

func detectExistingSchemaBaseline(ctx context.Context, db *sql.DB) (int, error) {
	containsTable := func(tableName string) (bool, error) {
		var count int
		err := db.QueryRowContext(ctx, `SELECT COUNT(*)
			FROM information_schema.tables
			WHERE table_schema = DATABASE() AND table_name = ?`, tableName).Scan(&count)
		return count > 0, err
	}

	baselines := []struct {
		version int
		tables  []string
	}{
		{version: 3, tables: []string{"school_profiles", "generated_documents"}},
		{version: 2, tables: []string{"permissions", "document_types"}},
		{version: 1, tables: []string{"users", "periods", "placements"}},
	}
	for _, baseline := range baselines {
		complete := true
		for _, table := range baseline.tables {
			present, err := containsTable(table)
			if err != nil {
				return 0, fmt.Errorf("inspect existing schema baseline: %w", err)
			}
			if !present {
				complete = false
				break
			}
		}
		if complete {
			return baseline.version, nil
		}
	}
	return 0, nil
}

func ensureMigrationTable(ctx context.Context, db *sql.DB) error {
	const statement = `CREATE TABLE IF NOT EXISTS schema_migrations (
		version BIGINT NOT NULL,
		dirty BOOLEAN NOT NULL DEFAULT FALSE
	)`
	if _, err := db.ExecContext(ctx, statement); err != nil {
		return fmt.Errorf("create schema_migrations table: %w", err)
	}
	var count int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM schema_migrations").Scan(&count); err != nil {
		return fmt.Errorf("inspect schema_migrations table: %w", err)
	}
	if count == 0 {
		if _, err := db.ExecContext(ctx, "INSERT INTO schema_migrations (version, dirty) VALUES (0, FALSE)"); err != nil {
			return fmt.Errorf("initialize schema_migrations table: %w", err)
		}
	}
	return nil
}

func currentMigration(ctx context.Context, db *sql.DB) (int, bool, error) {
	var version int
	var dirty bool
	if err := db.QueryRowContext(ctx, "SELECT version, dirty FROM schema_migrations ORDER BY version DESC LIMIT 1").Scan(&version, &dirty); err != nil {
		return 0, false, fmt.Errorf("read schema_migrations version: %w", err)
	}
	return version, dirty, nil
}

func embeddedMigrations() ([]migrationFile, error) {
	entries, err := migrationassets.FS.ReadDir("migrations")
	if err != nil {
		return nil, fmt.Errorf("read embedded migrations: %w", err)
	}
	result := make([]migrationFile, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".up.sql") {
			continue
		}
		match := migrationFilename.FindStringSubmatch(entry.Name())
		if len(match) != 2 {
			return nil, fmt.Errorf("invalid migration filename %q", entry.Name())
		}
		version, err := strconv.Atoi(match[1])
		if err != nil {
			return nil, fmt.Errorf("parse migration version %q: %w", entry.Name(), err)
		}
		result = append(result, migrationFile{Version: version, Name: entry.Name()})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Version < result[j].Version })
	for index := 1; index < len(result); index++ {
		if result[index-1].Version == result[index].Version {
			return nil, fmt.Errorf("duplicate migration version %d", result[index].Version)
		}
	}
	return result, nil
}

func applyMigration(ctx context.Context, db *sql.DB, migration migrationFile) error {
	content, err := migrationassets.FS.ReadFile(path.Join("migrations", migration.Name))
	if err != nil {
		return fmt.Errorf("read migration %s: %w", migration.Name, err)
	}
	if _, err := db.ExecContext(ctx, "UPDATE schema_migrations SET version = ?, dirty = TRUE", migration.Version); err != nil {
		return fmt.Errorf("mark migration %s dirty: %w", migration.Name, err)
	}
	for index, statement := range splitSQLStatements(string(content)) {
		if _, err := db.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("apply migration %s statement %d: %w", migration.Name, index+1, err)
		}
	}
	if _, err := db.ExecContext(ctx, "UPDATE schema_migrations SET dirty = FALSE WHERE version = ?", migration.Version); err != nil {
		return fmt.Errorf("mark migration %s clean: %w", migration.Name, err)
	}
	return nil
}

// splitSQLStatements handles the migration subset used by SIMPKL: regular
// MySQL statements with semicolons ignored while inside quoted strings.
func splitSQLStatements(source string) []string {
	var statements []string
	var current strings.Builder
	var quote rune
	escaped := false
	for _, character := range source {
		if quote != 0 {
			current.WriteRune(character)
			if escaped {
				escaped = false
				continue
			}
			if character == '\\' {
				escaped = true
			} else if character == quote {
				quote = 0
			}
			continue
		}
		if character == '\'' || character == '"' || character == '`' {
			quote = character
			current.WriteRune(character)
			continue
		}
		if character == ';' {
			if statement := strings.TrimSpace(current.String()); statement != "" {
				statements = append(statements, statement)
			}
			current.Reset()
			continue
		}
		current.WriteRune(character)
	}
	if statement := strings.TrimSpace(current.String()); statement != "" {
		statements = append(statements, statement)
	}
	return statements
}
