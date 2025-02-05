// nolint: mnd // This is strongly tied to the business, skipping all magic number checks.
// In Go, floating-point arithmetic can lead to precision loss. Please convert everything to decimal for calculations.
// Arbitrary-precision fixed-point decimal numbers in go.
// Note: Decimal library can "only" represent numbers with a maximum of 2^31 digits after the decimal point.
// https://github.com/govalues/decimal
// https://gtglobal.jp.larksuite.com/sheets/CY47s29DmhmBtEtLTA5jKGNYpVb?from=from_copylink
package decimal

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

type Decimal struct {
	decimal.Decimal
}

// "数字处理规则
// a.数据小于1万：正常显示，无小数，例：9999
// b.数据大于等于1万，小于1亿：使用万（萬）为单位处理数字
// c.数据大于等于1亿：使用亿（億）为单位处理数字
// 处理后的纯数字展示规则
//
//	a.数据大于等于100：小数点后展示1位数字，之后四舍五入 例：381.3万
//	b.数据小于100：小数点后展示2位数字，小数点不足使用0填充，之后四舍五入 例：89.12万".
func (d *Decimal) StringStatsIntCN() string {
	tenThousand := decimal.NewFromInt(1e4)    // 1万
	hundredMillion := decimal.NewFromInt(1e8) // 1亿

	switch {
	case d.LessThan(tenThousand) && d.GreaterThan(tenThousand.Neg()): // 数据小于1万：正常显示，无小数
		return d.StringFixed(0)
	case d.LessThan(hundredMillion) && d.GreaterThan(hundredMillion.Neg()): // 数据大于等于1万，小于1亿：使用“万”为单位
		val := d.Div(tenThousand)
		return formatWithUnit(val, "万")
	default: // 数据大于等于1亿：使用“亿”为单位
		val := d.Div(hundredMillion)
		return formatWithUnit(val, "亿")
	}
}

// "数字处理规则
// a.数据小于1000：正常显示，无小数， 例：999，（若出现异常数据，整数后截断）
// b.数据大于等于1000，小于100万：使用K为单位处理数字
// c..数据大于等于100万，小于10亿：使用M为单位处理数字
// d.数据大于等于10亿,小于10000亿：使用B为单位处理数字
// e.数据大于等于10000亿：使用T为单位处理数字
// 处理后的纯数字展示规则
//
//	a.数据大于等于100：小数点后展示1位数字，小数点不足使用0填充，之后四舍五入 例：381.3M
//	b.数据小于100：小数点后展示2位数字，小数点不足使用0填充，之后四舍五入 例：89.12M".
func (d *Decimal) StringStatsInt() string {
	billion := decimal.NewFromInt(1e9)   // 10亿
	million := decimal.NewFromInt(1e6)   // 100万
	thousand := decimal.NewFromInt(1e3)  // 1000
	trillion := decimal.NewFromInt(1e12) // 1万亿

	switch {
	case d.LessThan(thousand) && d.GreaterThan(thousand.Neg()): // 数据小于1000：正常显示，无小数
		return d.StringIntPart()
	case d.LessThan(million) && d.GreaterThan(million.Neg()): // 数据大于等于1000，小于100万：使用K为单位
		val := d.Div(thousand)
		return formatWithUnit(val, "K")
	case d.LessThan(billion) && d.GreaterThan(billion.Neg()): // 数据大于等于100万，小于10亿：使用M为单位
		val := d.Div(million)
		return formatWithUnit(val, "M")
	case d.LessThan(trillion) && d.GreaterThan(trillion.Neg()): // 数据大于等于10亿,小于10000亿：使用B为单位
		val := d.Div(billion)
		return formatWithUnit(val, "B")
	default: // 数据大于等于10000亿：使用T为单位
		val := d.Div(trillion)
		return formatWithUnit(val, "T")
	}
}

func (d *Decimal) StringStatsIntTW() string {
	tenThousand := decimal.NewFromInt(1e4)    // 1万
	hundredMillion := decimal.NewFromInt(1e8) // 1亿
	switch {
	case d.LessThan(tenThousand) && d.GreaterThan(tenThousand.Neg()): // 数据小于1万：正常显示，无小数
		return d.StringFixed(0)
	case d.LessThan(hundredMillion) && d.GreaterThan(hundredMillion.Neg()): // 数据大于等于1万，小于1亿：使用“万”为单位
		val := d.Div(tenThousand)
		return formatWithUnit(val, "萬")
	default: // 数据大于等于1亿：使用“亿”为单位
		val := d.Div(hundredMillion)
		return formatWithUnit(val, "億")
	}
}

func formatWithUnit(val decimal.Decimal, unit string) string {
	if val.GreaterThanOrEqual(decimal.NewFromInt(100)) || val.LessThanOrEqual(decimal.NewFromInt(-100)) {
		return fmt.Sprintf("%s%s", val.StringFixed(1), unit)
	}
	return fmt.Sprintf("%s%s", val.StringFixed(2), unit)
}

// "数字处理规则
// a.数据小于1000：正常显示，小数点后显示2位数字，小数点不足使用0填充，之后四舍五入， 例：999.12
// b.数据大于等于1000，小于100万：使用K为单位处理数字
// c..数据大于等于100万，小于10亿：使用M为单位处理数字
// d.数据大于等于10亿,小于10000亿：使用B为单位处理数字
// e.数据大于等于10000亿：使用T为单位处理数字
// 处理后的纯数字展示规则
//
//	a.数据大于等于100：小数点后展示1位数字，1位之后四舍五入 例：381.3M
//	b.数据小于100：小数点后展示2位数字，小数点不足使用0填充，之后四舍五入  例：89.12M".
func (d *Decimal) StringStatsDec() string {
	thousand := decimal.NewFromInt(1e3)  // 1000
	million := decimal.NewFromInt(1e6)   // 100万
	billion := decimal.NewFromInt(1e9)   // 10亿
	trillion := decimal.NewFromInt(1e12) // 1万亿

	switch {

	case d.LessThan(thousand) && d.GreaterThan(thousand.Neg()): // 数据小于1000：正常显示，小数点后显示2位数字
		return d.StringFixed(2)

	case d.LessThan(million) && d.GreaterThan(million.Neg()): // 数据大于等于1000，小于100万：使用K为单位
		val := d.Div(thousand)
		return formatWithUnit(val, "K")

	case d.LessThan(billion) && d.GreaterThan(billion.Neg()): // 数据大于等于100万，小于10亿：使用M为单位
		val := d.Div(million)
		return formatWithUnit(val, "M")

	case d.LessThan(trillion) && d.GreaterThan(trillion.Neg()): // 数据大于等于10亿,小于10000亿：使用B为单位
		val := d.Div(billion)
		return formatWithUnit(val, "B")

	default: // 数据大于等于10000亿：使用T为单位
		val := d.Div(trillion)
		return formatWithUnit(val, "T")
	}
}

// "数字处理规则
// a.数据小于1万：正常显示，小数点后显示2位数字，小数点不足使用0填充，之后四舍五入 例：9999.12
// b.数据大于等于1万，小于1亿：使用万（萬）为单位处理数字
// c.数据大于等于1亿：使用亿（億）为单位处理数字
// 处理后的纯数字展示规则
//
//	a.数据大于等于100：小数点后展示1位数字，之后四舍五入 例：381.3万
//	b.数据小于100：小数点后展示2位数字，小数点不足使用0填充，之后四舍五入 例：89.12万".
func (d *Decimal) StringStatsDecCN() string {
	tenThousand := decimal.NewFromInt(1e4)    // 1万
	hundredMillion := decimal.NewFromInt(1e8) // 1亿

	switch {
	case d.LessThan(tenThousand) && d.GreaterThan(tenThousand.Neg()): // 数据小于1万：正常显示，小数点后显示2位数字
		return d.StringFixed(2)

	case d.LessThan(hundredMillion) && d.GreaterThan(hundredMillion.Neg()): // 数据大于等于1万，小于1亿：使用“万”为单位
		val := d.Div(tenThousand)
		return formatWithUnit(val, "万")

	default: // 数据大于等于1亿：使用“亿”为单位
		val := d.Div(hundredMillion)
		return formatWithUnit(val, "亿")
	}
}

func (d *Decimal) StringStatsDecTW() string {
	tenThousand := decimal.NewFromInt(1e4)    // 1万
	hundredMillion := decimal.NewFromInt(1e8) // 1亿
	switch {
	case d.LessThan(tenThousand) && d.GreaterThan(tenThousand.Neg()): // 数据小于1万：正常显示，小数点后显示2位数字
		return d.StringFixed(2)
	case d.LessThan(hundredMillion) && d.GreaterThan(hundredMillion.Neg()): // 数据大于等于1万，小于1亿：使用“万”为单位
		val := d.Div(tenThousand)
		return formatWithUnit(val, "萬")
	default: // 数据大于等于1亿：使用“亿”为单位
		val := d.Div(hundredMillion)
		return formatWithUnit(val, "億")
	}
}

// 请注意需要保证传入的数为整数d.IsInteger(),否者截断.
func (d *Decimal) StringInt() string {
	return d.StringIntPart()
}

func (d *Decimal) StringDecLow() string {
	return truncateToDecimalPlaces(d, 2, false)
}

// 截断到指定的小数位数,addZero 设置为 true 时补零.
func truncateToDecimalPlaces(d *Decimal, places int, addZero bool) string {
	// 转换为字符串并保留最多 `places` 位小数
	str := d.String()
	decimalIndex := len(str)
	for i, ch := range str {
		if ch == '.' {
			decimalIndex = i
			break
		}
	}
	if decimalIndex == len(str) {
		// 没有小数点
		if addZero {
			return str + "." + strings.Repeat("0", places)
		}
		return str
	}

	// 如果小数部分超过所需的小数位数，进行截断
	decimalPart := str[decimalIndex+1:]
	if len(decimalPart) > places {
		decimalPart = decimalPart[:places]
	}
	if addZero {
		decimalPart = decimalPart + strings.Repeat("0", places-len(decimalPart))
		res := str[:decimalIndex+1] + decimalPart
		return res
	}
	res := str[:decimalIndex+1] + decimalPart
	ress, _ := decimal.NewFromString(res)
	return ress.String()
}

func (d *Decimal) StringDecMid() string {
	return truncateToDecimalPlaces(d, 6, false)
}

func (d *Decimal) StringDecHigh() string {
	return truncateToDecimalPlaces(d, 8, false)
}

// "8位小数以内有有效数字，最大显示8位，不足不补0（数据尾部的0都不展示），超出截断，例：0.12345678、0.122
// 8位小数以内无有效数字全是0，最大显示12位，不足不补0（数据尾部的0都不展示），超出截断，例：0.000000001234、0.000000001".
func (d *Decimal) StringUltraHigh() string {
	// 获取小数位数
	decimalPart := d.String()
	decimalIndex := len(decimalPart)
	for i, ch := range decimalPart {
		if ch == '.' {
			decimalIndex = i
			break
		}
	}
	if decimalIndex == len(decimalPart) {
		// 没有小数点
		return d.String()
	}

	// 小数部分
	decimalPart = decimalPart[decimalIndex+1:]
	// 前12位为0，取整数
	if strings.HasPrefix(decimalPart, "000000000000") {
		return d.StringFixed(0)
	}
	// 前八位为0，取12位
	if strings.HasPrefix(decimalPart, "00000000") {
		return truncateToDecimalPlaces(d, 12, false)
	}
	return truncateToDecimalPlaces(d, 8, false)
}

func (d *Decimal) StringParcentInt() string {
	return strings.Split(d.Mul(decimal.NewFromInt(100)).String(), ".")[0] + "%"
}

// Percent_Low	百分比，最多小数点后2位 ，超过截断 ，不足不补0（数据尾部的0都不展示）.
func (d *Decimal) StringParcentLow() string {
	return truncateToDecimalPlaces(&Decimal{d.Mul(decimal.NewFromInt(100))}, 2, false) + "%"
}

func (d *Decimal) StringParcentMid() string {
	return truncateToDecimalPlaces(&Decimal{d.Mul(decimal.NewFromInt(100))}, 4, true) + "%"
}

func (d *Decimal) StringParcentHigh() string {
	return truncateToDecimalPlaces(&Decimal{d.Mul(decimal.NewFromInt(100))}, 6, true) + "%"
}

func NewFromFloat(f float64) Decimal {
	return Decimal{Decimal: decimal.NewFromFloat(f)}
}

func NewFromInt(i int64) Decimal {
	return Decimal{Decimal: decimal.NewFromInt(i)}
}

func NewFromUint64(i uint64) Decimal {
	return Decimal{Decimal: decimal.NewFromUint64(i)}
}

func NewFromFloat64(f float64) Decimal {
	return Decimal{Decimal: decimal.NewFromFloat(f)}
}

func NewFromString(s string) (Decimal, error) {
	d, err := decimal.NewFromString(s)
	if err != nil {
		return Decimal{}, err
	}
	return Decimal{Decimal: d}, nil
}

func (d *Decimal) StringIntPart() string {
	return strings.Split(d.String(), ".")[0]
}
