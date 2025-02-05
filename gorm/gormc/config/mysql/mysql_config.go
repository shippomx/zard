package mysql

import (
	"errors"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/shippomx/zard/core/logx"
	"github.com/shippomx/zard/dbresolver"
	"github.com/shippomx/zard/gorm/gormc"
	"github.com/shippomx/zard/gorm/gormc/config"
)

// nolint:all
type Conf struct {
	Sources  []dbresolver.DSN
	Replicas []dbresolver.DSN

	MaxIdleConns    int           `json:",default=10"` // 空闲中的最大连接数
	MaxOpenConns    int           `json:",default=10"` // 打开到数据库的最大连接数,
	ConnMaxIdleTime time.Duration `json:",default=0s"` // 连接最大空闲时间
	ConnMaxLifetime time.Duration `json:",default=0s"` // 连接最大生命周期
	// TODO: 后面都升级上来后，再将类型换成 time.Duration .
	// 并移除 `string` json tag.
	HostResolveInterval string `json:",string,default=2s"`   // DNS 解析轮询间隔
	HealthCheckInterval string `json:",string,default=2s"`   // 数据库实例健康检查间隔
	HealthCheckTimeout  string `json:",string,default=2s"`   // 数据库实例健康检查超时
	DBStatusLogInterval string `json:",string,default=15s"`  // 数据库实例状态打印间隔
	SlowThreshold       string `json:",string,default=10ms"` // 慢查询阈值

	MaxHealthCheckRetry uint32        `json:",default=10"`                                        // 数据库实例错误重试次
	Policy              string        `json:",default=roundRobin,options=roundRobin|random"`      // 负载均衡策略
	LogLevel            string        `json:",default=info,options=debug|info|warn|error|silent"` // 日志等级
	TraceResolverMode   bool          `json:",default=false"`                                     // 是否开启TraceResolver模式,
	Role                string        `json:",default=master"`                                    // 数据库标识，用于 metric
	ObserveInterval     time.Duration `json:",default=10s"`                                       // db.Stat 指标监控周期
}

func (c *Conf) GetGormLogMode() logger.LogLevel {
	return config.OverwriteGormLogMode(c.LogLevel)
}

func (c *Conf) GetSlowThreshold() time.Duration {
	return CompatNumberToDuration(c.SlowThreshold, time.Millisecond)
}

func (c *Conf) GetColorful() bool {
	return true
}

func Connect(c Conf) (*gorm.DB, error) {
	return ConnectWithConfig(c, nil)
}

func MustConnect(c Conf) *gorm.DB {
	db, err := Connect(c)
	logx.Must(err)
	return db
}

func ConnectWithConfig(c Conf, cfg *gorm.Config) (*gorm.DB, error) {
	var dsn string
	switch {
	case len(c.Sources) != 0:
		dsn = c.Sources[0].Str()
	case len(c.Replicas) != 0:
		dsn = c.Replicas[0].Str()
	default:
		return nil, errors.New("empty data sources")
	}
	mysqlCfg := mysql.Config{
		DSN: dsn,
	}
	if cfg == nil {
		cfg = &gorm.Config{}
	}
	if cfg.Logger == nil {
		cfg.Logger = config.NewLogxGormLogger(&c)
	}
	db, err := gorm.Open(mysql.New(mysqlCfg), cfg)
	if err != nil {
		return nil, err
	}

	if err = initPlugin(db, c); err != nil {
		return nil, err
	}
	return db, nil
}

func MustConnectWithConfig(c Conf, cfg *gorm.Config) *gorm.DB {
	db, err := ConnectWithConfig(c, cfg)
	logx.Must(err)
	return db
}

func initPlugin(db *gorm.DB, conf Conf) error {
	db.Use(dbresolver.Register(
		dbresolver.Config{
			Sources:             conf.Sources,
			Replicas:            conf.Replicas,
			MaxIdleConns:        conf.MaxIdleConns,
			MaxOpenConns:        conf.MaxOpenConns,
			ConnMaxIdleTime:     conf.ConnMaxIdleTime,
			ConnMaxLifetime:     conf.ConnMaxLifetime,
			HostResolveInterval: CompatNumberToDuration(conf.HostResolveInterval, time.Second),
			HealthCheckInterval: CompatNumberToDuration(conf.HealthCheckInterval, time.Second),
			HealthCheckTimeout:  CompatNumberToDuration(conf.HealthCheckTimeout, time.Second),
			DBStatusLogInterval: CompatNumberToDuration(conf.DBStatusLogInterval, time.Second),
			MaxHealthCheckRetry: conf.MaxHealthCheckRetry,
			Policy:              dbresolver.PolicyFrom(conf.Policy),
			TraceResolverMode:   conf.TraceResolverMode,
		},
	))

	if err := db.Use(gormc.OtelPlugin{}); err != nil {
		return err
	}

	metricsPlugin := gormc.MetricsPlugin{
		DbRole:          conf.Role,
		ObserveInterval: conf.ObserveInterval,
	}
	if err := db.Use(&metricsPlugin); err != nil {
		return err
	}

	return nil
}
