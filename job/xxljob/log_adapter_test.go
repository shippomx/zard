package xxljob

import (
	"testing"

	"github.com/shippomx/zard/core/logx"
	"github.com/stretchr/testify/assert"
)

func TestLogxAdapter_Info(t *testing.T) {
	writer := &TestLogWriter{}
	logx.SetWriter(writer)

	adapter := &logxAdapter{}
	adapter.Info("test message %s", "info")

	assert.Contains(t, writer.String(), "DEBUG: test message info")
}

func TestLogxAdapter_Error(t *testing.T) {
	writer := &TestLogWriter{}
	logx.SetWriter(writer)

	adapter := &logxAdapter{}
	adapter.Error("test message %s", "error")

	assert.Contains(t, writer.String(), "ERROR: test message error")
}
