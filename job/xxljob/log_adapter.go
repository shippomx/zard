package xxljob

import (
	"github.com/shippomx/zard/core/logx"
)

type logxAdapter struct{}

func (l *logxAdapter) Info(format string, a ...interface{}) {
	logx.Debugf(format, a...)
}

func (l *logxAdapter) Error(format string, a ...interface{}) {
	logx.Errorf(format, a...)
}
