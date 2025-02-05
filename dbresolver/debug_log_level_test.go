package dbresolver

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm/logger"
)

type mockWriter struct {
	msg string
}

func (m *mockWriter) Printf(msg string, data ...interface{}) {
	data = data[1:]
	m.msg = fmt.Sprintf(msg, data...)
}

func (m *mockWriter) Reset() {
	m.msg = ""
}

func TestTransformLogger(t *testing.T) {
	w := &mockWriter{}
	log := logger.New(w, logger.Config{LogLevel: logger.Info})
	loggerWithDebug := TransformLogger(log)
	loggerWithDebug.Debug(context.Background(), "test1")
	assert.Equal(t, "", w.msg)
	EnableDebug()
	loggerWithDebug.Debug(context.Background(), "test2")
	assert.Contains(t, w.msg, "test2")
	w.Reset()

	newLogger := &GormLogWithDebug{log}
	assert.Equal(t, newLogger, TransformLogger(newLogger))
}
