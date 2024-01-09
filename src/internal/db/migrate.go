package custdb

import (
	"context"
	custerror "github.com/CE-Thesis-2023/ltd/internal/error"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func buildGorm(ctx context.Context, options *Options) (*gorm.DB, error) {
	connString := options.globalConfigs.Connection
	db, err := gorm.Open(
		sqlite.Open(connString),
		&gorm.Config{})
	if err != nil {
		return nil, custerror.FormatInternalError("buildGorm: err = %s", err)
	}

	return db, nil
}

func Migrate(schemas ...interface{}) error {
	return gormDb.AutoMigrate(schemas...)
}
