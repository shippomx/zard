package gormc

import (
	"database/sql"
	"time"

	"github.com/shippomx/zard/core/metric"
	"github.com/shippomx/zard/dbresolver"
	"gorm.io/gorm"
)

const (
	metricsTimeKey   = "metricsTime"
	metricsNamespace = "gorm_zero"
)

var (
	sqlCount = metric.NewCounterVec(&metric.CounterVecOpts{
		Namespace: metricsNamespace,
		Name:      "sql_count_total",
		Help:      "sql request counter",
		Labels:    []string{"role", "table", "action"},
	})

	queryTimeHistogram = metric.NewHistogramVec(&metric.HistogramVecOpts{ // nolint: promlinter
		Namespace: metricsNamespace,
		Name:      "sql_duration_ms",
		Help:      "sql request duration ms",
		Labels:    []string{"role", "table", "action"},
		Buckets:   []float64{5, 10, 20, 50, 100, 200, 500, 1000, 2000},
	})

	errCount = metric.NewCounterVec(&metric.CounterVecOpts{
		Namespace: metricsNamespace,
		Name:      "sql_err_count_total",
		Help:      "sql err counter",
		Labels:    []string{"role", "table", "action"},
	})

	maxSQLOpenConnections = metric.NewGaugeVec(&metric.GaugeVecOpts{
		Namespace: metricsNamespace,
		Name:      "sql_max_open_connections",
		Help:      "maximum number of open connections to the database",
		Labels:    []string{"role", "address"},
	})

	sqlOpenConnections = metric.NewGaugeVec(&metric.GaugeVecOpts{
		Namespace: metricsNamespace,
		Name:      "sql_open_connections",
		Help:      "the number of established connections both in use and idle",
		Labels:    []string{"role", "address"},
	})

	sqlConnectionsInUse = metric.NewGaugeVec(&metric.GaugeVecOpts{
		Namespace: metricsNamespace,
		Name:      "sql_connections_in_use",
		Help:      "the number of connections currently in use",
		Labels:    []string{"role", "address"},
	})

	sqlConnectionsIdle = metric.NewGaugeVec(&metric.GaugeVecOpts{
		Namespace: metricsNamespace,
		Name:      "sql_connections_idle",
		Help:      "the number of idle connections",
		Labels:    []string{"role", "address"},
	})

	sqlWaitCount = metric.NewGaugeVec(&metric.GaugeVecOpts{ // nolint: promlinter
		Namespace: metricsNamespace,
		Name:      "sql_wait_count",
		Help:      "the total number of connections waited for",
		Labels:    []string{"role", "address"},
	})

	sqlWaitDurtion = metric.NewGaugeVec(&metric.GaugeVecOpts{
		Namespace: metricsNamespace,
		Name:      "sql_wait_duration",
		Help:      "the total time blocked waiting for a new connection",
		Labels:    []string{"role", "address"},
	})

	sqlMaxIdleClosed = metric.NewGaugeVec(&metric.GaugeVecOpts{
		Namespace: metricsNamespace,
		Name:      "sql_max_idle_closed",
		Help:      "the total number of connections closed due to SetMaxIdleConns",
		Labels:    []string{"role", "address"},
	})

	sqlMaxIdleTimeClosed = metric.NewGaugeVec(&metric.GaugeVecOpts{
		Namespace: metricsNamespace,
		Name:      "sql_max_idle_time_closed",
		Help:      "the total number of connections closed due to SetConnMaxIdleTime",
		Labels:    []string{"role", "address"},
	})

	sqlMaxLifetimeClosed = metric.NewGaugeVec(&metric.GaugeVecOpts{
		Namespace: metricsNamespace,
		Name:      "sql_max_life_time_closed",
		Help:      "the total number of connections closed due to SetConnMaxLifeTime",
		Labels:    []string{"role", "address"},
	})
)

var setDBStatsMetric = func(stat sql.DBStats, role, addr string) {
	maxSQLOpenConnections.Set(float64(stat.MaxOpenConnections), role, addr)
	sqlOpenConnections.Set(float64(stat.OpenConnections), role, addr)
	sqlConnectionsInUse.Set(float64(stat.InUse), role, addr)
	sqlConnectionsIdle.Set(float64(stat.Idle), role, addr)
	sqlWaitCount.Set(float64(stat.WaitCount), role, addr)
	sqlWaitDurtion.Set(float64(stat.WaitDuration.Milliseconds()), role, addr)
	sqlMaxIdleClosed.Set(float64(stat.MaxIdleClosed), role, addr)
	sqlMaxIdleTimeClosed.Set(float64(stat.MaxIdleTimeClosed), role, addr)
	sqlMaxLifetimeClosed.Set(float64(stat.MaxLifetimeClosed), role, addr)
}

type MetricsPlugin struct {
	ObserveInterval time.Duration
	DbRole          string
}

func (o *MetricsPlugin) Name() string {
	return "gorm_metrics"
}

func (o *MetricsPlugin) Initialize(db *gorm.DB) (err error) {
	if err = db.Callback().Query().Before("*").Register("gorm-zero:query_metrics_before", o.Before); err != nil {
		return err
	}

	if err = db.Callback().Query().After("*").Register("gorm-zero:query_metrics_after", o.QueryAfter); err != nil {
		return err
	}

	if err = db.Callback().Create().Before("*").Register("gorm-zero:create_metrics_before", o.Before); err != nil {
		return err
	}

	if err = db.Callback().Create().After("*").Register("gorm-zero:create_metrics_after", o.CreateAfter); err != nil {
		return err
	}

	if err = db.Callback().Update().Before("*").Register("gorm-zero:update_metrics_before", o.Before); err != nil {
		return err
	}

	if err = db.Callback().Update().After("*").Register("gorm-zero:update_metrics_after", o.UpdateAfter); err != nil {
		return err
	}

	if err = db.Callback().Delete().Before("*").Register("gorm-zero:delete_metrics_before", o.Before); err != nil {
		return err
	}

	if err = db.Callback().Delete().After("*").Register("gorm-zero:delete_metrics_after", o.DeleteAfter); err != nil {
		return err
	}

	InitStdDatabaseSQLMetrics(db, o.DbRole, o.ObserveInterval)
	return
}

func InitStdDatabaseSQLMetrics(db *gorm.DB, dbRole string, interval time.Duration) {
	if interval < 1*time.Second {
		// TODO: add a warn log
		interval = 1 * time.Second
	}
	go func() {
		for {
			for _, inst := range dbresolver.GlobalInstances.List() {
				sqlDB, err := inst.SQLDB()
				if err != nil || sqlDB == nil {
					continue
				}
				stat := sqlDB.Stats()
				addr := inst.GetAddress()
				setDBStatsMetric(stat, dbRole, addr)
			}
			sqlDB, err := db.DB()
			if err == nil && sqlDB != nil {
				setDBStatsMetric(sqlDB.Stats(), dbRole, "global")
			}
			time.Sleep(interval)
		}
	}()
}

const (
	ActionCreate = "create"
	ActionUpdate = "update"
	ActionDelete = "delete"
	ActionQuery  = "query"
)

func (o *MetricsPlugin) Before(db *gorm.DB) {
	now := time.Now()
	db.InstanceSet(metricsTimeKey, now)
}

func (o *MetricsPlugin) QueryAfter(db *gorm.DB) {
	o.After(db, ActionQuery)
}

func (o *MetricsPlugin) CreateAfter(db *gorm.DB) {
	o.After(db, ActionCreate)
}

func (o *MetricsPlugin) UpdateAfter(db *gorm.DB) {
	o.After(db, ActionUpdate)
}

func (o *MetricsPlugin) DeleteAfter(db *gorm.DB) {
	o.After(db, ActionDelete)
}

func (o *MetricsPlugin) After(db *gorm.DB, action string) {
	value, ok := db.InstanceGet(metricsTimeKey)
	if !ok {
		return
	}

	startTime := value.(time.Time)
	sqlTime := time.Since(startTime).Milliseconds()
	queryTimeHistogram.Observe(sqlTime, o.DbRole, db.Statement.Table, action)

	sqlCount.Inc(o.DbRole, db.Statement.Table, action)

	if db.Error != nil {
		errCount.Inc(o.DbRole, db.Statement.Table, action)
	}
}
