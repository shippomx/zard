package mysql

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/shippomx/zard/core/conf"
)

func TestCompatNumberToDuration(t *testing.T) {
	var c struct {
		Duration string `json:",string,default=1s"` // nolint:all
	}

	for _, tt := range []struct {
		configYAML string
		expected   time.Duration
	}{
		{`Duration: 10`, 10 * time.Second},
		{`Duration: 10.0`, 10 * time.Second},
		{`Duration: 10m`, 10 * time.Minute},
		{`Test: 10`, 1 * time.Second},
	} {
		assert.Nil(t, conf.LoadFromYamlBytes([]byte(tt.configYAML), &c))
		assert.Equal(t, tt.expected, CompatNumberToDuration(c.Duration, time.Second))
	}
}
