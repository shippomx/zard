package pg

import (
	"errors"
	"fmt"
	"time"

	"github.com/shippomx/zard/gorm/gormc"
	"github.com/shippomx/zard/gorm/gormc/config"

	"github.com/shippomx/zard/core/logx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Conf struct {
	Username      string
	Password      string
	Path          string
	Port          int    `json:",default=5432"`
	SslMode       string `json:",default=disable,options=disable|enable"`
	TimeZone      string `json:",default=Asia/Shanghai"`
	Dbname        string
	MaxIdleConns  int    `json:",default=10"`                               // 空闲中的最大连接数
	MaxOpenConns  int    `json:",default=10"`                               // 打开到数据库的最大连接数
	LogMode       string `json:",default=dev,options=dev|test|prod|silent"` // 是否开启Gorm全局日志
	SlowThreshold int64  `json:",default=1000"`
}

func (m *Conf) Dsn() string {
	return fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%d sslmode=%s TimeZone=%s", m.Username, m.Password, m.Dbname, m.Path, m.Port, m.SslMode, m.TimeZone)
}
func (m *Conf) GetGormLogMode() logger.LogLevel {
	return config.OverwriteGormLogMode(m.LogMode)
}

func (m *Conf) GetSlowThreshold() time.Duration {
	return time.Duration(m.SlowThreshold) * time.Millisecond
}
func (m *Conf) GetColorful() bool {
	return true
}

func Connect(m Conf) (*gorm.DB, error) {
	if m.Dbname == "" {
		return nil, errors.New("database name is empty")
	}
	newLogger := config.NewLogxGormLogger(&m)
	pgsqlCfg := postgres.Config{
		DSN:                  m.Dsn(),
		PreferSimpleProtocol: true, // disables implicit prepared statement usage
	}
	db, err := gorm.Open(postgres.New(pgsqlCfg), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		return nil, err
	}

	if err = initPlugin(db); err != nil {
		return nil, err
	}

	sqldb, _ := db.DB()
	sqldb.SetMaxIdleConns(m.MaxIdleConns)
	sqldb.SetMaxOpenConns(m.MaxOpenConns)
	return db, nil

}

func MustConnect(m Conf) *gorm.DB {
	if m.Dbname == "" {
		logx.Must(errors.New("database name is empty"))
	}
	newLogger := config.NewLogxGormLogger(&m)
	pgsqlCfg := postgres.Config{
		DSN:                  m.Dsn(),
		PreferSimpleProtocol: true, // disables implicit prepared statement usage
	}
	db, err := gorm.Open(postgres.New(pgsqlCfg), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		logx.Must(err)
	}

	if err = initPlugin(db); err != nil {
		logx.Must(err)
	}

	sqldb, _ := db.DB()
	sqldb.SetMaxIdleConns(m.MaxIdleConns)
	sqldb.SetMaxOpenConns(m.MaxOpenConns)
	return db

}

func ConnectWithConfig(m Conf, cfg *gorm.Config) (*gorm.DB, error) {
	if m.Dbname == "" {
		return nil, errors.New("database name is empty")
	}
	pgsqlCfg := postgres.Config{
		DSN:                  m.Dsn(),
		PreferSimpleProtocol: true, // disables implicit prepared statement usage
	}
	db, err := gorm.Open(postgres.New(pgsqlCfg), cfg)
	if err != nil {
		return nil, err
	}

	if err = initPlugin(db); err != nil {
		return nil, err
	}

	sqldb, _ := db.DB()
	sqldb.SetMaxIdleConns(m.MaxIdleConns)
	sqldb.SetMaxOpenConns(m.MaxOpenConns)
	return db, nil

}

func MustConnectWithConfig(m Conf, cfg *gorm.Config) *gorm.DB {
	if m.Dbname == "" {
		logx.Must(errors.New("database name is empty"))
	}
	pgsqlCfg := postgres.Config{
		DSN:                  m.Dsn(),
		PreferSimpleProtocol: true, // disables implicit prepared statement usage
	}
	db, err := gorm.Open(postgres.New(pgsqlCfg), cfg)
	if err != nil {
		logx.Must(err)
	}

	if err = initPlugin(db); err != nil {
		logx.Must(err)
	}

	sqldb, _ := db.DB()
	sqldb.SetMaxIdleConns(m.MaxIdleConns)
	sqldb.SetMaxOpenConns(m.MaxOpenConns)

	return db
}

func initPlugin(db *gorm.DB) error {
	if err := db.Use(gormc.OtelPlugin{}); err != nil {
		return err
	}

	// 由于 metricsPlugin 做过重构，且并未对 pg 做过测试，这块暂时先注释了。
	// TODO: support metrics in the future
	// if err := db.Use(&metricsPlugin); err != nil {
	// 	return err
	// }

	return nil
}
