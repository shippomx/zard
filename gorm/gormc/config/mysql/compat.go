// 向后兼容的补偿代码
package mysql

import (
	"strconv"
	"time"

	"github.com/shippomx/zard/core/logx"
)

// `dStr` can be an number string or duration string: "0", "10", "3s", "1m", "1h1m", etc.
// if `dStr` is an number string, the default unit is **second**.
func CompatNumberToDuration(dStr string, unit time.Duration) (d time.Duration) {
	if num, err := strconv.Atoi(dStr); err == nil {
		logx.Warn("(Deprecated) database: number as duration, use duration string like '1s'")
		d = time.Duration(num) * unit
	} else {
		// var err error
		d, err = time.ParseDuration(dStr)
		logx.Must(err)
	}
	return d
}
