package custdb

import (
	"context"
	"labs/local-transcoder/internal/configs"
	custerror "labs/local-transcoder/internal/error"
	"labs/local-transcoder/internal/logger"
	"os"
	"path"
	"sync"

	_ "github.com/glebarez/go-sqlite"
	"github.com/jmoiron/sqlx"
	"gorm.io/gorm"
)

var once sync.Once

var db *sqlx.DB

var gormDb *gorm.DB

func Init(ctx context.Context, options ...Optioner) {
	once.Do(func() {
		opts := &Options{}
		for _, opt := range options {
			opt(opts)
		}

		globalConfigs := opts.globalConfigs
		log := logger.Sugar()

		log.Infof("db.sqlite.Init: creating database dsn = %s", globalConfigs.Connection)

		if err := createIfNotExists(globalConfigs.Connection); err != nil {
			log.Fatal(err)
			return
		}

		client, err := sqlx.Connect("sqlite", globalConfigs.Connection)
		if err != nil {
			log.Fatalf("db.sqlite.Init: open err = %s", err)
			return
		}

		gormClient, err := buildGorm(ctx, opts)
		if err != nil {
			log.Fatalf("db.sqlite.Init: gormClient err = %s", err)
			return
		}

		db = client
		gormDb = gormClient

		LayeredInit()
	})
}

func createIfNotExists(p string) error {
	fs, err := os.Stat(p)
	if err != nil {
		if !os.IsNotExist(err) {
			return custerror.FormatInternalError("db.sqlite.createIfNotExists: os.Stat err = %s", err)
		}
	}

	if fs != nil {
		return nil
	}
	dir := path.Dir(p)

	os.MkdirAll(dir, 0755)
	if _, err := os.Create(p); err != nil {
		return custerror.FormatInternalError("db.sqlite.createIfNotExists: os.Create err = %s", err)
	}

	return nil
}

func Db() *sqlx.DB {
	return db
}

func Gorm() *gorm.DB {
	return gormDb
}

func Stop(ctx context.Context) error {
	if db != nil {
		if err := db.Close(); err != nil {
			return custerror.FormatInternalError("db.sqlite.Stop: sqlx.Close() err = %s", err)
		}
	}
	return nil
}

type Options struct {
	globalConfigs *configs.DatabaseConfigs
}

type Optioner func(opts *Options)

func WithGlobalConfigs(configs *configs.DatabaseConfigs) Optioner {
	return func(opts *Options) {
		opts.globalConfigs = configs
	}
}
