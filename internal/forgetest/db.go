// Package forgetest provides test infrastructure utilities for forge-generated applications.
// It includes helpers for isolated PostgreSQL test databases, HTTP test servers,
// and Datastar SSE form submission testing.
package forgetest

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os/exec"
	"path/filepath"
	"runtime"

	pgtestdb "github.com/peterldowns/pgtestdb"
	"github.com/peterldowns/pgtestdb/migrators/common"

	_ "github.com/jackc/pgx/v5/stdlib" // pgx sql driver
	"github.com/jackc/pgx/v5/pgxpool"
	"testing"
)

// TestDBConfig holds connection and migration configuration for test databases.
type TestDBConfig struct {
	Host        string // default: "localhost"
	Port        string // default: "5432"
	User        string // default: "postgres"
	Password    string // default: "postgres"
	Database    string // default: "postgres"
	Options     string // default: "sslmode=disable"
	MigrationDir string // path to migrations/ directory
	AtlasBin    string // path to atlas binary (e.g. .forge/bin/atlas)
}

// DefaultTestDBConfig returns a TestDBConfig using localhost PostgreSQL defaults.
// The MigrationDir and AtlasBin are resolved relative to the repository root
// using runtime.Caller so they work regardless of which package calls NewTestDB.
func DefaultTestDBConfig() TestDBConfig {
	_, filename, _, _ := runtime.Caller(0)
	// filename is internal/forgetest/db.go; repo root is two levels up
	repoRoot := filepath.Join(filepath.Dir(filename), "../..")
	repoRoot = filepath.Clean(repoRoot)

	return TestDBConfig{
		Host:         "localhost",
		Port:         "5432",
		User:         "postgres",
		Password:     "postgres",
		Database:     "postgres",
		Options:      "sslmode=disable",
		MigrationDir: filepath.Join(repoRoot, "migrations"),
		AtlasBin:     filepath.Join(repoRoot, ".forge", "bin", "atlas"),
	}
}

// WithMigrationDir returns an option function that overrides the migrations directory.
func WithMigrationDir(path string) func(*TestDBConfig) {
	return func(c *TestDBConfig) {
		c.MigrationDir = path
	}
}

// WithAtlasBin returns an option function that overrides the atlas binary path.
func WithAtlasBin(path string) func(*TestDBConfig) {
	return func(c *TestDBConfig) {
		c.AtlasBin = path
	}
}

// WithDatabaseURL returns an option function that parses a DATABASE_URL and
// sets the Host, Port, User, Password, Database, and Options fields.
func WithDatabaseURL(databaseURL string) func(*TestDBConfig) {
	return func(c *TestDBConfig) {
		u, err := url.Parse(databaseURL)
		if err != nil {
			return
		}
		if h := u.Hostname(); h != "" {
			c.Host = h
		}
		if p := u.Port(); p != "" {
			c.Port = p
		}
		if u.User != nil {
			if user := u.User.Username(); user != "" {
				c.User = user
			}
			if pass, ok := u.User.Password(); ok && pass != "" {
				c.Password = pass
			}
		}
		if db := u.Path; db != "" {
			// strip leading slash
			c.Database = db[1:]
		}
		if q := u.RawQuery; q != "" {
			c.Options = q
		}
	}
}

// atlasMigrator is a pgtestdb.Migrator that uses the Atlas CLI to apply
// migrations from a migrations/ directory.
type atlasMigrator struct {
	migrationDir string
	atlasBin     string
}

// Hash returns a deterministic hash based on the contents of all .sql files
// in the migrations directory. If no migrations exist, returns "empty".
func (m atlasMigrator) Hash() (string, error) {
	h, err := common.HashDir(m.migrationDir)
	if err != nil {
		// If migrations dir doesn't exist yet, return stable empty hash
		return "empty", nil
	}
	return h, nil
}

// Migrate applies all pending migrations to the template database using atlas migrate apply.
func (m atlasMigrator) Migrate(_ context.Context, _ *sql.DB, conf pgtestdb.Config) error {
	dbURL := conf.URL()

	var stdout, stderr bytes.Buffer
	cmd := exec.Command(
		m.atlasBin,
		"migrate", "apply",
		"--dir", "file://"+m.migrationDir,
		"--url", dbURL,
	)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("atlas migrate apply failed: %w\nOutput: %s%s", err, stdout.String(), stderr.String())
	}

	return nil
}

// NewTestDB creates an isolated PostgreSQL test schema per test using pgtestdb.
// The schema is created from a template using Atlas migrations. Each test gets
// its own database that is automatically dropped when the test completes.
//
// Option functions can override any TestDBConfig field:
//
//	db := forgetest.NewTestDB(t, forgetest.WithDatabaseURL(os.Getenv("DATABASE_URL")))
func NewTestDB(t *testing.T, opts ...func(*TestDBConfig)) *sql.DB {
	t.Helper()

	cfg := DefaultTestDBConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	conf := pgtestdb.Config{
		DriverName: "pgx",
		Host:       cfg.Host,
		Port:       cfg.Port,
		User:       cfg.User,
		Password:   cfg.Password,
		Database:   cfg.Database,
		Options:    cfg.Options,
	}

	migrator := atlasMigrator{
		migrationDir: cfg.MigrationDir,
		atlasBin:     cfg.AtlasBin,
	}

	return pgtestdb.New(t, conf, migrator)
}

// NewTestPool creates an isolated PostgreSQL test schema per test and returns
// a *pgxpool.Pool for use with the project's pgxpool-based code.
//
// Most forge integration tests should use NewTestPool rather than NewTestDB.
//
// Option functions can override any TestDBConfig field:
//
//	pool := forgetest.NewTestPool(t, forgetest.WithDatabaseURL(os.Getenv("DATABASE_URL")))
func NewTestPool(t *testing.T, opts ...func(*TestDBConfig)) *pgxpool.Pool {
	t.Helper()

	cfg := DefaultTestDBConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	// Use pgtestdb.Custom to get the config of the isolated test database
	// without keeping an open *sql.DB connection that would interfere with pgxpool.
	conf := pgtestdb.Config{
		DriverName: "pgx",
		Host:       cfg.Host,
		Port:       cfg.Port,
		User:       cfg.User,
		Password:   cfg.Password,
		Database:   cfg.Database,
		Options:    cfg.Options,
	}

	migrator := atlasMigrator{
		migrationDir: cfg.MigrationDir,
		atlasBin:     cfg.AtlasBin,
	}

	testConf := pgtestdb.Custom(t, conf, migrator)
	if testConf == nil {
		t.Fatal("pgtestdb.Custom returned nil config")
		return nil
	}

	pool, err := pgxpool.New(context.Background(), testConf.URL())
	if err != nil {
		t.Fatalf("forgetest.NewTestPool: pgxpool.New failed: %v", err)
		return nil
	}

	t.Cleanup(func() {
		pool.Close()
	})

	return pool
}
