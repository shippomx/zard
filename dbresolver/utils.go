package dbresolver

import (
	"fmt"
	"regexp"
	"sync/atomic"
	"time"
)

var fromTableRegexp = regexp.MustCompile("(?i)(?:FROM|UPDATE|MERGE INTO|INSERT [a-z ]*INTO) ['`\"]?([a-zA-Z0-9_]+)([ '`\",)]|$)")

func getTableFromRawSQL(sql string) string {
	if matches := fromTableRegexp.FindAllStringSubmatch(sql, -1); len(matches) > 0 {
		return matches[0][1]
	}

	return ""
}

func StartCronJob(stop chan bool, atOnce bool, interval time.Duration, jobName string, job func()) {
	fmt.Println("starting cron job:", jobName)
	if atOnce {
		job()
	}
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-stop:
			fmt.Println("stopping cron job:", jobName)
			return
		case <-t.C:
			job()
		}
	}
}

func NewAtomicBool(initialValue bool) *atomic.Bool {
	b := &atomic.Bool{}
	b.Store(initialValue)
	return b
}
