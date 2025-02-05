package gormc

import (
	"context"
	"errors"
	"time"

	"github.com/shippomx/zard/core/logx"
	"github.com/shippomx/zard/dbresolver"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	// Silent silent log level.
	Silent logger.LogLevel = iota + github.com/shippomx/zard1
	// Error error log level.
	Error
	// Warn warn log level.
	Warn
	// Info info log level.
	Info
	// Debug debug log level.
	Debug
)

var _ dbresolver.LoggerWithDebug = (*GormLog)(nil)

type GormLog struct {
	Level                     logger.LogLevel
	IgnoreRecordNotFoundError bool
	SlowThreshold             time.Duration
}

func (g *GormLog) LogMode(level logger.LogLevel) logger.Interface {
	newLog := *g
	newLog.Level = level
	return &newLog
}

func (g *GormLog) Debug(ctx context.Context, format string, data ...any) {
	if g.Level < Debug {
		return
	}
	logx.WithContext(ctx).Debugf(format, data...)
}

func (g *GormLog) Info(ctx context.Context, format string, data ...any) {
	if g.Level < Info {
		return
	}
	logx.WithContext(ctx).Infof(format, data...)
}

func (g *GormLog) Warn(ctx context.Context, format string, data ...any) {
	if g.Level < Warn {
		return
	}
	logx.WithContext(ctx).Warnf(format, data...)
}

func (g *GormLog) Error(ctx context.Context, format string, data ...any) {
	if g.Level < Error {
		return
	}
	logx.WithContext(ctx).Errorf(format, data...)
}

func (g *GormLog) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if g.Level <= Silent {
		return
	}

	elapsed := time.Since(begin)
	switch {
	case err != nil && g.Level >= logger.Error && (!errors.Is(err, gorm.ErrRecordNotFound) || !g.IgnoreRecordNotFoundError):
		sql, rows := fc()
		if rows == -1 {
			logx.WithContext(ctx).WithDuration(elapsed).Errorw(
				err.Error(),
				logx.Field("sql", sql),
			)
		} else {
			logx.WithContext(ctx).WithDuration(elapsed).Errorw(
				err.Error(),
				logx.Field("sql", sql),
				logx.Field("rows", rows),
			)
		}
	case elapsed > g.SlowThreshold && g.SlowThreshold != 0 && g.Level >= logger.Warn:
		sql, rows := fc()
		if rows == -1 {
			logx.WithContext(ctx).WithDuration(elapsed).Sloww(
				"slow sql",
				logx.Field("sql", sql),
			)
		} else {
			logx.WithContext(ctx).WithDuration(elapsed).Sloww(
				"slow sql",
				logx.Field("sql", sql),
				logx.Field("rows", rows),
			)
		}
	case g.Level == logger.Info:
		sql, rows := fc()
		if rows == -1 {
			logx.WithContext(ctx).WithDuration(elapsed).Infow(
				"",
				logx.Field("sql", sql),
			)
		} else {
			logx.WithContext(ctx).WithDuration(elapsed).Infow(
				"",
				logx.Field("sql", sql),
				logx.Field("rows", rows),
			)
		}
	}
}
