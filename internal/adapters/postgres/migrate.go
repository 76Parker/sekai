package postgres

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const migrationLockID int64 = 839470861337251

var migrationFileRE = regexp.MustCompile(`^(\d+)_.+\.up\.sql$`)

type migration struct {
	version  int64
	name     string
	checksum string
	sql      string
}

func Migrate(ctx context.Context, db *pgxpool.Pool, dir string) error {
	migrations, err := loadMigrations(dir)
	if err != nil {
		return err
	}
	if len(migrations) == 0 {
		return nil
	}

	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock($1);`, migrationLockID); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `
	CREATE TABLE IF NOT EXISTS public.schema_migrations (
		version BIGINT PRIMARY KEY,
		name TEXT NOT NULL,
		checksum TEXT NOT NULL,
		applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
	);
	`); err != nil {
		return err
	}

	applied, err := appliedMigrations(ctx, tx)
	if err != nil {
		return err
	}

	for _, migration := range migrations {
		checksum, ok := applied[migration.version]
		if ok {
			if checksum != migration.checksum {
				return fmt.Errorf("migration %s checksum mismatch", migration.name)
			}
			continue
		}

		if _, err := tx.Exec(ctx, migration.sql); err != nil {
			return fmt.Errorf("apply migration %s: %w", migration.name, err)
		}

		if _, err := tx.Exec(ctx, `
		INSERT INTO public.schema_migrations (version, name, checksum)
		VALUES ($1, $2, $3);
		`, migration.version, migration.name, migration.checksum); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func loadMigrations(dir string) ([]migration, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	migrations := make([]migration, 0)
	versions := make(map[int64]string)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		matches := migrationFileRE.FindStringSubmatch(name)
		if matches == nil {
			continue
		}

		version, err := strconv.ParseInt(matches[1], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parse migration version %s: %w", name, err)
		}
		if existing, ok := versions[version]; ok {
			return nil, fmt.Errorf("duplicate migration version %d: %s and %s", version, existing, name)
		}
		versions[version] = name

		path := filepath.Join(dir, name)
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}

		hash := sha256.Sum256(content)
		migrations = append(migrations, migration{
			version:  version,
			name:     name,
			checksum: hex.EncodeToString(hash[:]),
			sql:      string(content),
		})
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].version < migrations[j].version
	})

	return migrations, nil
}

func appliedMigrations(ctx context.Context, tx pgx.Tx) (map[int64]string, error) {
	rows, err := tx.Query(ctx, `
	SELECT version, checksum
	FROM public.schema_migrations;
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[int64]string)
	for rows.Next() {
		var (
			version  int64
			checksum string
		)
		if err := rows.Scan(&version, &checksum); err != nil {
			return nil, err
		}
		applied[version] = checksum
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return applied, nil
}
