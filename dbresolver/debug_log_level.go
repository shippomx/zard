package dbresolver

import (
	"context"

	"gorm.io/gorm/logger"
)

var (
	DEBUG = NewAtomicBool(false)

	_ LoggerWithDebug  = (*GormLogWithDebug)(nil)
	_ logger.Interface = (*GormLogWithDebug)(nil)
)

type LoggerWithDebug interface {
	logger.Interface
	Debug(context.Context, string, ...interface{})
}

type GormLogWithDebug struct {
	logger.Interface
}

// EnableDebug 开启 debug 日志.
// Note(roby): 仅当单独引入 dbresolver 的时候，该函数才可生效。因为如果使用了 gorm-zero,
// gorm-zero 会默认使用其自带的 logger 替换 gorm.DB 里的 logger.
func EnableDebug() {
	DEBUG.Store(true)
}

func (g *GormLogWithDebug) Debug(ctx context.Context, format string, data ...any) {
	if !DEBUG.Load() {
		return
	}
	g.Interface.Info(ctx, format, data...)
}

// TransformLogger 转换 gorm.DB 的 Logger 为带有 Debug 函数的 logger.
func TransformLogger(logger logger.Interface) LoggerWithDebug {
	if loggerWithDebug, ok := logger.(LoggerWithDebug); ok {
		return loggerWithDebug
	}
	return &GormLogWithDebug{logger}
}
