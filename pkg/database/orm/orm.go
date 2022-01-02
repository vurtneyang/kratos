package orm

import (
	"context"
	"fmt"
	"strings"
	"time"

	"kratos/pkg/ecode"
	"kratos/pkg/log"
	xtime "kratos/pkg/time"

	// database driver
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"

	mysqlV2 "gorm.io/driver/mysql"
	gormV2 "gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

// Config mysql config.
type Config struct {
	DSN         string         // data source name.
	Active      int            // pool
	Idle        int            // pool
	IdleTimeout xtime.Duration // connect max life time.
}

type ormLog struct{}

func (l ormLog) Print(v ...interface{}) {
	log.Info(strings.Repeat("%v ", len(v)), v...)
}

func (l ormLog) Printf(format string, v ...interface{}) {
	log.Warnv(context.Background(),
		log.KVString("log", fmt.Sprintf(format, v...)),
		log.KVString("source", "mysql-log"),
	)
}

func init() {
	gorm.ErrRecordNotFound = ecode.NothingFound
}

// NewMySQL new db and retry connection when has error.
func NewMySQL(c *Config) (db *gorm.DB) {
	db, err := gorm.Open("mysql", c.DSN)
	if err != nil {
		log.Error("orm: open error(%v)", err)
		panic(err)
	}
	db.DB().SetMaxIdleConns(c.Idle)
	db.DB().SetMaxOpenConns(c.Active)
	db.DB().SetConnMaxLifetime(time.Duration(c.IdleTimeout))
	db.SetLogger(ormLog{})

	db.Callback().Create().Replace("gorm:update_time_stamp", updateTimeStampForCreateCallback)
	db.Callback().Update().Replace("gorm:update_time_stamp", updateTimeStampForUpdateCallback)
	return
}

func NewMySqlV2(c *Config) (db *gormV2.DB) {
	db, err := gormV2.Open(mysqlV2.Open(c.DSN), &gormV2.Config{
		Logger: gormLogger.New(ormLog{}, gormLogger.Config{
			SlowThreshold: 200 * time.Millisecond,
			LogLevel:      gormLogger.Warn,
		}),
	})

	if err != nil {
		log.Error("orm: open error(%v)", err)
		panic(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Error("orm: db pool error(%v)", err)
		panic(err)
	}
	sqlDB.SetMaxIdleConns(c.Idle)
	sqlDB.SetMaxOpenConns(c.Active)
	sqlDB.SetConnMaxLifetime(time.Duration(c.IdleTimeout))

	return
}
