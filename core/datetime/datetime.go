// nolint: mnd //This is strongly tied to the business, skipping all magic number checks.
// https://gtglobal.jp.larksuite.com/sheets/CY47s29DmhmBtEtLTA5jKGNYpVb?from=from_copylink
package datetime

import (
	"fmt"
	"time"
)

func FormatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

func FormatRelativeTime(published time.Time) string {
	return FormatRelativeTimeWithCurrentTime(time.Now(), published)
}

func FormatRelativeTimeWithCurrentTime(now, published time.Time) string {
	diff := now.Sub(published)
	if diff < time.Minute {
		return "Just now"
	} else if diff < time.Hour {
		minutes := int(diff.Minutes())
		return fmt.Sprintf("%dm ago", minutes)
	} else if diff < 24*time.Hour {
		// 1 hour ≤ published time < 1 day
		hours := int(diff.Hours())
		return fmt.Sprintf("%dh ago", hours)
	} else if diff < 3*24*time.Hour {
		// 1 day ≤ published time < 3 days
		days := int(diff.Hours() / 24)
		return fmt.Sprintf("%dd ago", days)
	} else if published.Year() == now.Year() {
		// Within this year
		return published.Format("01-02")
	}
	// Not this year
	return published.Format("2006-01-02")
}

func FormatRelativeTimeCN(published time.Time) string {
	return FormatRelativeTimeWithCurrentTimeCN(time.Now(), published)
}

func FormatRelativeTimeWithCurrentTimeCN(now time.Time, published time.Time) string {
	diff := now.Sub(published)
	if diff < time.Minute {
		return "刚刚"
	} else if diff < time.Hour {
		// 小于 1 小时
		minutes := int(diff.Minutes())
		return fmt.Sprintf("%d分钟前", minutes)
	} else if diff < 24*time.Hour {
		// 1 小时 ≤ 发布时间 < 1 天
		hours := int(diff.Hours())
		return fmt.Sprintf("%d小时前", hours)
	} else if diff < 3*24*time.Hour {
		// 1 天 ≤ 发布时间 < 3 天
		days := int(diff.Hours() / 24)
		return fmt.Sprintf("%d天前", days)
	} else if published.Year() == now.Year() {
		// 今年内
		return published.Format("01-02")
	}
	// 非今年
	return published.Format("2006-01-02")
}

func FormatRelativeTimeTW(published time.Time) string {
	return FormatRelativeTimeWithCurrentTimeTW(time.Now(), published)
}

func FormatRelativeTimeWithCurrentTimeTW(now time.Time, published time.Time) string {
	diff := now.Sub(published)
	if diff < time.Minute {
		return "剛剛"
	} else if diff < time.Hour {
		// 小于 1 小时
		minutes := int(diff.Minutes())
		return fmt.Sprintf("%d分鐘前", minutes)
	} else if diff < 24*time.Hour {
		// 1 小时 ≤ 发布时间 < 1 天
		hours := int(diff.Hours())
		return fmt.Sprintf("%d小時前", hours)
	} else if diff < 3*24*time.Hour {
		// 1 天 ≤ 发布时间 < 3 天
		days := int(diff.Hours() / 24)
		return fmt.Sprintf("%d天前", days)
	} else if published.Year() == now.Year() {
		// 今年内
		return published.Format("01-02")
	}
	// 非今年
	return published.Format("2006-01-02")
}
