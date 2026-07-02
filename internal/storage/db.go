package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/Velarvo/velarvo-desktop/internal/logger"
	"github.com/Velarvo/velarvo-desktop/internal/storage/migrations"
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
	_ "modernc.org/sqlite"
)

type DB struct {
	read  *sql.DB
	write *sql.DB
	path  string
}

var migrationMu sync.Mutex

func storageLog() *zap.SugaredLogger {
	return logger.Named("storage")
}

func Open(dataDir string) (*DB, error) {
	log := storageLog()
	dbPath := filepath.Join(dataDir, "vault.db")

	dsn := fmt.Sprintf(
		"file:%s?_pragma=journal_mode(WAL)"+
			"&_pragma=synchronous(NORMAL)"+
			"&_pragma=foreign_keys(ON)"+
			"&_pragma=temp_store(MEMORY)"+
			"&_pragma=mmap_size(134217728)"+ // 128 MiB
			"&_pragma=cache_size(-65536)"+ // -64 MiB
			"&_pragma=busy_timeout(5000)",
		dbPath,
	)

	readDB, err := sql.Open("sqlite", dsn)
	if err != nil {
		log.Errorw("failed to open read database connection", "path", dbPath, "error", err)
		return nil, fmt.Errorf("open read db: %w", err)
	}
	readDB.SetMaxOpenConns(4)
	readDB.SetMaxIdleConns(4)
	readDB.SetConnMaxLifetime(time.Hour)

	writeDB, err := sql.Open("sqlite", dsn)
	if err != nil {
		_ = readDB.Close()
		log.Errorw("failed to open write database connection", "path", dbPath, "error", err)
		return nil, fmt.Errorf("open write db: %w", err)
	}
	writeDB.SetMaxOpenConns(1)
	writeDB.SetMaxIdleConns(1)
	writeDB.SetConnMaxLifetime(time.Hour)

	db := &DB{read: readDB, write: writeDB, path: dbPath}
	log.Infow("database connections opened", "path", dbPath)

	if err := db.verify(); err != nil {
		_ = db.Close()
		log.Errorw("database verification failed", "path", dbPath, "error", err)
		return nil, fmt.Errorf("verify: %w", err)
	}
	if err := db.migrate(); err != nil {
		_ = db.Close()
		log.Errorw("database migration failed", "path", dbPath, "error", err)
		return nil, fmt.Errorf("migrate: %w", err)
	}

	log.Infow("database ready", "path", dbPath)
	return db, nil
}

func (db *DB) verify() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var journalMode string
	if err := db.write.QueryRowContext(ctx, "PRAGMA journal_mode").Scan(&journalMode); err != nil {
		return fmt.Errorf("read journal_mode: %w", err)
	}
	if journalMode != "wal" {
		return fmt.Errorf("expected WAL journal mode, got %q", journalMode)
	}

	var foreignKeys int
	if err := db.write.QueryRowContext(ctx, "PRAGMA foreign_keys").Scan(&foreignKeys); err != nil {
		return fmt.Errorf("read foreign_keys: %w", err)
	}
	if foreignKeys != 1 {
		return errors.New("foreign_keys not enabled")
	}

	storageLog().Debugw("database pragmas verified", "path", db.path, "journalMode", journalMode, "foreignKeys", foreignKeys)
	return nil
}

func (db *DB) migrate() error {
	migrationMu.Lock()
	defer migrationMu.Unlock()

	log := storageLog()

	goose.SetBaseFS(migrations.FS)
	goose.SetLogger(goose.NopLogger())
	if err := goose.SetDialect("sqlite3"); err != nil {
		log.Errorw("failed to configure migration dialect", "path", db.path, "error", err)
		return err
	}

	log.Infow("running database migrations", "path", db.path)

	if err := goose.Up(db.write, "."); err != nil {
		log.Errorw("database migrations failed", "path", db.path, "error", err)
		return err
	}

	log.Infow("database migrations completed", "path", db.path)
	return nil
}

func (db *DB) Read() *sql.DB { return db.read }

func (db *DB) Write() *sql.DB { return db.write }

func (db *DB) Tx(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := db.write.BeginTx(ctx, nil)
	if err != nil {
		storageLog().Errorw("failed to begin transaction", "path", db.path, "error", err)
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (db *DB) Checkpoint(ctx context.Context) error {
	_, err := db.write.ExecContext(ctx, "PRAGMA wal_checkpoint(TRUNCATE)")
	if err != nil {
		storageLog().Errorw("database checkpoint failed", "path", db.path, "error", err)
		return err
	}

	storageLog().Debugw("database checkpoint completed", "path", db.path)
	return err
}

func (db *DB) Backup(ctx context.Context, destPath string) error {
	log := storageLog()
	log.Infow("starting database backup", "sourcePath", db.path, "destinationPath", destPath)

	if err := db.Checkpoint(ctx); err != nil {
		return fmt.Errorf("checkpoint: %w", err)
	}
	if _, err := db.write.ExecContext(ctx, "VACUUM INTO ?", destPath); err != nil {
		log.Errorw("database backup failed", "sourcePath", db.path, "destinationPath", destPath, "error", err)
		return fmt.Errorf("vacuum into: %w", err)
	}

	log.Infow("database backup completed", "sourcePath", db.path, "destinationPath", destPath)
	return nil
}

func (db *DB) Path() string { return db.path }

func (db *DB) Close() error {
	log := storageLog()
	var firstErr error
	if err := db.read.Close(); err != nil {
		log.Errorw("failed to close read database connection", "path", db.path, "error", err)
		firstErr = err
	}
	if err := db.write.Close(); err != nil && firstErr == nil {
		log.Errorw("failed to close write database connection", "path", db.path, "error", err)
		firstErr = err
	}
	if firstErr == nil {
		log.Infow("database connections closed", "path", db.path)
	}
	return firstErr
}
