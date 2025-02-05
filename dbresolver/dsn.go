package dbresolver

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
)

var DSNDefaultTimeout = 3 * time.Second

//nolint:all
type DSN struct {
	Path     string         // 服务器地址
	Port     int            `json:",default=3306"`
	Config   string         `json:",optional"`
	Dbname   string         // 数据库名
	Username string         // 数据库用户名
	Password string         // 数据库密码
	Type     DataSourceType `json:",optional"`
	IPv4     string         `json:",optional"`
}

func (dsn DSN) Str() string {
	host := dsn.IPv4
	if host == "" {
		host = dsn.Path
	}
	dsnStr := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s",
		dsn.Username, dsn.Password,
		host, dsn.Port,
		dsn.Dbname,
		dsn.GetConnConfig())

	cfg, _ := mysql.ParseDSN(dsnStr)
	if cfg == nil {
		return dsnStr
	}

	if cfg.Timeout == 0 {
		cfg.Timeout = DSNDefaultTimeout
	}

	return cfg.FormatDSN()
}

func (dsn DSN) Addr() Addr {
	return (Addr)(fmt.Sprintf("%s:%d", dsn.IPv4, dsn.Port))
}

func (dsn DSN) GetConnConfig() string {
	if dsn.Config == "" {
		return "charset=utf8mb4&parseTime=True&loc=Local"
	}
	return dsn.Config
}

func (dsn DSN) EqualAddrTo(other DSN) bool {
	if dsn.Port != other.Port {
		return false
	}
	getHost := func(dsn DSN) string {
		hostmap := map[string]string{
			"localhost": "127.0.0.1",
			"127.1":     "127.0.0.1",
		}
		host, ok := hostmap[dsn.Path]
		if !ok {
			return dsn.Path
		}
		return host
	}
	return getHost(dsn) == getHost(other)
}

func (dsn DSN) DEBUG() string {
	return strings.Join(append([]string{},
		fmt.Sprintf("DataSourceType=%s", dsn.Type),
		fmt.Sprintf("Host=%s", dsn.Path),
		fmt.Sprintf("IPv4=%s", dsn.IPv4),
		fmt.Sprintf("Port=%d", dsn.Port),
		fmt.Sprintf("Database=%s", dsn.Dbname),
		fmt.Sprintf("User=%s", dsn.Username),
	), "\t")
}
