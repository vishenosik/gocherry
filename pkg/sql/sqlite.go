package sql

import (
	"context"
	"fmt"
	"io/fs"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"github.com/pressly/goose/v3"
	"github.com/vishenosik/gocherry/pkg/config"
)

type SqliteConfig struct {
	StorePath string `validate:"required"`
}

type SqliteConfigEnv struct {
	StorePath string `env:"SQLITE_STORE_PATH" default:"./storage/store.db" desc:"A path to sqlite store file"`
}

type SqliteStore struct {
	storePath      string
	migrationsFS   fs.FS
	migrationsPath string
	db             *sqlx.DB
}

type SqliteStoreOption func(*SqliteStore)

func validateSqliteConfig(conf SqliteConfig) error {
	const op = "validateConfig"
	valid := validator.New()
	if err := valid.Struct(conf); err != nil {
		return errors.Wrap(err, op)
	}
	return nil
}

func NewSqliteStore(opts ...SqliteStoreOption) (*SqliteStore, error) {
	var envConf SqliteConfigEnv
	if err := config.ReadConfig(&envConf); err != nil {
		return nil, errors.Wrap(err, "setup logger: failed to read config")
	}

	return NewSqliteStoreConfig(SqliteConfig{
		StorePath: envConf.StorePath,
	}, opts...)
}

func NewSqliteStoreConfig(conf SqliteConfig, opts ...SqliteStoreOption) (*SqliteStore, error) {

	if err := validateSqliteConfig(conf); err != nil {
		return nil, err
	}

	ss := &SqliteStore{
		storePath: conf.StorePath,
	}

	for _, opt := range opts {
		opt(ss)
	}

	return ss, nil
}

func (ss *SqliteStore) Close(_ context.Context) error {
	return ss.db.Close()
}

func (ss *SqliteStore) Open(_ context.Context) (*sqlx.DB, error) {
	db, err := sqlx.Open("sqlite3", ss.storePath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to sqlite")
	}

	ss.db = db

	if ss.migrationsFS == nil {
		return db, nil
	}

	goose.SetLogger(goose.NopLogger())
	goose.SetBaseFS(ss.migrationsFS)

	if err := goose.SetDialect("sqlite"); err != nil {
		return nil, fmt.Errorf("failed to set sqlite: %w", err)
	}

	if err := goose.Up(db.DB, ss.migrationsPath); err != nil {
		return nil, errors.Wrap(err, "failed to run migrations up")
	}
	return db, nil
}

func WithMigration(
	fs fs.FS,
	path string,
) SqliteStoreOption {
	return func(ss *SqliteStore) {
		if fs == nil || path == "" {
			return
		}
		ss.migrationsFS = fs
		ss.migrationsPath = path
	}
}

func SqliteUniqueError(err error, mapper map[string]string) error {

	if !strings.Contains(err.Error(), "UNIQUE constraint failed") {
		return nil
	}

	for fld := range mapper {
		if strings.Contains(err.Error(), fld) {
			return errors.Wrap(ErrAlreadyExists, mapper[fld])
		}
	}

	return nil
}
