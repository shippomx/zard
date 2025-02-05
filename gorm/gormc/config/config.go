package config

import (
	"time"

	"gorm.io/gorm/logger"

	logx "github.com/shippomx/zard/gorm/gormc"
)

type GormLogConfigI interface {
	GetGormLogMode() logger.LogLevel
	GetSlowThreshold() time.Duration
	GetColorful() bool
}

func NewLogxGormLogger(cfg GormLogConfigI) logger.Interface {
	return &logx.GormLog{
		Level:                     cfg.GetGormLogMode(),
		IgnoreRecordNotFoundError: true,
		SlowThreshold:             cfg.GetSlowThreshold(),
	}
}

func OverwriteGormLogMode(mode string) logger.LogLevel {
	switch mode {
	case "debug":
		return logx.Debug
	case "info":
		return logx.Info
	case "warn":
		return logx.Warn
	case "error":
		return logx.Error
	case "silent":
		return logx.Silent
	default:
		return logx.Info
	}
}
