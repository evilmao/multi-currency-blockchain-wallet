package db

import (
	"fmt"
	"time"

	"upex-wallet/wallet-base/newbitx/misclib/log"

	"github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	sqltrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/database/sql"
)

var defaultdb *gorm.DB

// New initialze gorm db instance.
func New(dsn string, serviceName string) (*gorm.DB, error) {
	var (
		conn *gorm.DB
		err  error
	)

	if len(serviceName) == 0 {
		conn, err = gorm.Open("mysql", dsn)
		if err != nil {
			log.Errorf("failed to connect database %+v", err)
			return nil, err
		}
	} else {
		option := sqltrace.WithServiceName(fmt.Sprintf("%s-mysql", serviceName))
		sqltrace.Register("mysql", &mysql.MySQLDriver{}, option)

		rawDB, err := sqltrace.Open("mysql", dsn)
		if err != nil {
			log.Errorf("failed to connect database %+v", err)
			return nil, err
		}

		conn, err = gorm.Open("mysql", rawDB)
		if err != nil {
			return nil, err
		}
	}

	conn.DB().Ping()
	conn.DB().SetConnMaxLifetime(time.Minute * 5)
	conn.DB().SetMaxIdleConns(10)
	conn.DB().SetMaxOpenConns(10)
	// conn.LogMode(true)
	if defaultdb == nil {
		defaultdb = conn
	}

	return conn, nil
}

// Default returns default db instance.
func Default() *gorm.DB {
	return defaultdb
}
