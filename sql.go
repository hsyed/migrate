package migrate

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

type sqlBackend struct {
	db *sql.DB
}

func Sql(db *sql.DB) Backend {
	return &sqlBackend{db: db}
}

func (s *sqlBackend) DestroyAndApply(ctx context.Context, schema *Schema) error {
	return s.doApply(ctx, true, schema)
}

func (s *sqlBackend) Apply(ctx context.Context, schema *Schema) error {
	return s.doApply(ctx, false, schema)
}

func (s *sqlBackend) doApply(ctx context.Context, destroy bool, schema *Schema) error {
	if err := validateSchema(schema); err != nil {
		return err
	}

	if conn, err := s.db.Conn(ctx); err != nil {
		return err
	} else {
		defer func() { _ = conn.Close() }()
		var version string
		if err := s.db.QueryRow(`SELECT version();`).Scan(&version); err != nil {
			return fmt.Errorf("could not determine if postgres: %w", err)
		} else if strings.HasPrefix(version, "PostgresSQL") {
			return errors.New("unknown database")
		} else if err := s.applyPostgres(ctx, conn, destroy, schema); err != nil {
			return err
		} else {
			return nil
		}
	}
}

func (s *sqlBackend) applyPostgres(ctx context.Context, conn *sql.Conn, destroy bool, schema *Schema) error {
	const pgLockKey int64 = 1616476926335464400

	var destroySql string
	if destroy {
		destroySql = fmt.Sprintf(`DROP SCHEMA IF EXISTS "%s" CASCADE;`, schema.Name)
	}

	var acquireSql = fmt.Sprintf(`
SELECT pg_advisory_lock(%d);
%s
CREATE SCHEMA IF NOT EXISTS "%s";
SET search_path TO "%s", public;
CREATE TABLE IF NOT EXISTS schema_migrations (version INT NOT NULL);
INSERT INTO schema_migrations(version) SELECT 0 WHERE 0=(SELECT count(*) FROM schema_migrations);
`, pgLockKey, destroySql, schema.Name, schema.Name)

	var releaseSql = fmt.Sprintf(`
SET search_path TO DEFAULT;
SELECT pg_advisory_unlock(%d);
	`, pgLockKey)

	if _, err := conn.ExecContext(ctx, acquireSql); err != nil {
		return fmt.Errorf("migration setup failed: %w", err)
	}

	defer func() {
		// Failing to reset the connection should result in a panic
		if _, err := conn.ExecContext(ctx, releaseSql); err != nil {
			panic(err)
		}
	}()

	var currentVersion int
	if err := conn.QueryRowContext(ctx, `SELECT version FROM schema_migrations;`).Scan(&currentVersion); err != nil {
		return fmt.Errorf("could not get migration state: %w", err)
	}

	for _, v := range filterSortChanges(currentVersion, schema.Changes) {
		if tx, err := conn.BeginTx(ctx, &sql.TxOptions{}); err != nil {
			return err
		} else if _, err = tx.ExecContext(ctx, v.ddl); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("could not apply change: %d", v.ordinal)
		} else if _, err = tx.ExecContext(ctx, `UPDATE schema_migrations SET version = $1;`, v.ordinal); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("could not update migration version for change %d: %w", v.ordinal, err)
		} else if err := tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}
