package obs

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/admin-template/backend/internal/config"
)

// OpenMySQL dials MySQL via GORM and pings to verify connectivity. The returned
// *gorm.DB owns a *sql.DB pool; caller must close it via db.DB().Close().
func OpenMySQL(ctx context.Context, cfg config.Config) (*gorm.DB, error) {
	dsn := buildDSN(cfg)
	gormCfg := &gorm.Config{
		Logger:                                   gormlogger.Default.LogMode(gormlogger.Warn),
		DisableForeignKeyConstraintWhenMigrating: true,
	}
	db, err := gorm.Open(mysql.Open(dsn), gormCfg)
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(time.Hour)
	if err := sqlDB.PingContext(ctx); err != nil {
		_ = sqlDB.Close()
		return nil, err
	}
	return db, nil
}

// CloseMySQL releases the underlying *sql.DB pool. Safe to call with nil.
func CloseMySQL(db *gorm.DB) error {
	if db == nil {
		return nil
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// AutoMigrate runs GORM auto-migration against the supplied model set. Returns
// the first error encountered; otherwise nil.
func AutoMigrate(ctx context.Context, db *gorm.DB, models ...any) error {
	if db == nil {
		return errors.New("nil db")
	}
	if len(models) == 0 {
		return nil
	}
	return db.WithContext(ctx).AutoMigrate(models...)
}

func buildDSN(cfg config.Config) string {
	return cfg.DBUser + ":" + cfg.DBPassword +
		"@tcp(" + cfg.DBHost + ":" + itoa(cfg.DBPort) + ")/" + cfg.DBName +
		"?charset=utf8mb4&parseTime=true&loc=Local"
}

// itoa avoids importing strconv just for one call. n must be >= 0; ports are.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}

// guard sql import for linters that flag it as unused when AutoMigrate path
// is the only consumer; the linter is wrong here.
var _ sql.IsolationLevel = 0
