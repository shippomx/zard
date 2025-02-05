package decimal

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestDecimal(t *testing.T) {
	d := decimal.NewFromInt(10)
	assert.Equal(t, d.String(), "10")

	d1 := NewFromFloat(1.000001)
	assert.Equal(t, d1.String(), "1.000001")

	d2 := NewFromFloat(.123123123123123)
	assert.Equal(t, d2.String(), "0.123123123123123")

	d3, err := NewFromString("1.000001")
	assert.Equal(t, err, nil)
	assert.Equal(t, d3.String(), "1.000001")

	_, err = NewFromString("null")
	assert.Error(t, err)
}

func TestStringStatsInt(t *testing.T) {
	// 规则4 大数统计-整数
	res1, err := NewFromString("10000000000001000000000000")
	if err != nil {
		t.Errorf("NewFromString() error = %v", err)
	}
	bigint, err := decimal.NewFromString("10000000000001000000000000")
	assert.Nil(t, err)
	tests := []struct {
		name string
		d    Decimal
		want string
	}{
		{"less than 1000", NewFromInt(999), "999"},
		{"equal to -1000", NewFromInt(-1000), "-1.00K"},
		{"equal to 1000", Decimal{decimal.NewFromInt(1000)}, "1.00K"},
		{"between -1000 and -1 million", Decimal{decimal.NewFromInt(-123456)}, "-123.5K"},
		{"between 1000 and 1 million", Decimal{decimal.NewFromInt(123456)}, "123.5K"},
		{"equal to -1 million", Decimal{decimal.NewFromInt(-1000000)}, "-1.00M"},
		{"equal to 1 million", Decimal{decimal.NewFromInt(1000000)}, "1.00M"},
		{"between -1 million and -1 billion", Decimal{decimal.NewFromInt(-123456789)}, "-123.5M"},
		{"between 1 million and 1 billion", Decimal{decimal.NewFromInt(123456789)}, "123.5M"},
		{"equal to 1 billion", Decimal{decimal.NewFromInt(1000000000)}, "1.00B"},
		{"between -1 billion and -1 trillion", Decimal{decimal.NewFromInt(-123456789012)}, "-123.5B"},
		{"between 1 billion and 1 trillion", Decimal{decimal.NewFromInt(123456789012)}, "123.5B"},
		{"equal to 1 trillion", Decimal{decimal.NewFromInt(1000000000000)}, "1.00T"},
		{"greater than 1 trillion", Decimal{decimal.NewFromInt(-1234567890123456)}, "-1234.6T"},
		{"greater than 1 trillion", Decimal{decimal.NewFromInt(1234567890123456)}, "1234.6T"},
		{"999999999999", NewFromInt(999999999999), "1000.0B"},
		{"-999999999999", NewFromInt(-999999999999), "-1000.0B"},
		{"bigint", Decimal{bigint}, "10000000000001.0T"},
		{"zero", NewFromInt(0), "0"},
		{"rounded less than 1000", NewFromFloat(999.99), "999"},
		{"equal to 1000", NewFromFloat(1000), "1.00K"},
		{"just below 1 million", NewFromFloat(999000), "999.0K"},
		{"equal to 1 million", NewFromFloat(1000000), "1.00M"},
		{"close to 900 million", NewFromFloat(899999999), "900.0M"},
		{"just below 1 billion", NewFromFloat(999940000), "999.9M"},
		{"equal to 1 billion", NewFromFloat(1000000000), "1.00B"},
		{"close to 900 billion", NewFromFloat(899999999999), "900.0B"},
		{"just below 1 trillion", NewFromFloat(999999999999), "1000.0B"},
		{"equal to 1 trillion", NewFromFloat(1000000000000), "1.00T"},
		{"specific 381 million", NewFromFloat(381000000), "381.0M"},
		{"close to 90 billion", NewFromFloat(89999000000), "90.00B"},
		{"large number over 10 trillion", res1, "10000000000001.0T"},
		{"slightly above 1000", NewFromFloat(1000.99), "1.00K"},
		{"negative 1234", NewFromFloat(-1234), "-1.23K"},
		{"negative 999 million", NewFromFloat(-999000000), "-999.0M"},
		{"negative just below 1 billion", NewFromFloat(-999999999), "-1000.0M"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.d.StringStatsInt(); got != tt.want {
				t.Errorf("input:%v StringStatsInt() = %v, want %v", tt.d, got, tt.want)
			}
		})
	}
}

func TestDecimal_StringStatsIntCN(t *testing.T) {
	s, err := NewFromString("10000000000001000000000000")
	assert.Nil(t, err)
	// 规则4 大数统计-整数 中文
	tests := []struct {
		name string
		d    Decimal
		want string
	}{
		{"less than 10,000", Decimal{decimal.NewFromInt(9999)}, "9999"},
		{"equal to -10,000", Decimal{decimal.NewFromInt(-10000)}, "-1.00万"},
		{"equal to 10,000", Decimal{decimal.NewFromInt(10000)}, "1.00万"},
		{"between -10,000 and -100,000,000", Decimal{decimal.NewFromInt(-50000)}, "-5.00万"},
		{"between 10,000 and 100,000,000", Decimal{decimal.NewFromInt(50000)}, "5.00万"},
		{"equal to -100,000,000", Decimal{decimal.NewFromInt(-100000000)}, "-1.00亿"},
		{"equal to 100,000,000", Decimal{decimal.NewFromInt(100000000)}, "1.00亿"},
		{"between -100,000,000 and -1,000,000,000,000", Decimal{decimal.NewFromInt(-500000000)}, "-5.00亿"},
		{"between 100,000,000 and 1,000,000,000,000", Decimal{decimal.NewFromInt(500000000)}, "5.00亿"},
		{"greater than 1,000,000,000,000", Decimal{decimal.NewFromInt(1234567890123456)}, "12345678.9亿"},
		{"bigger", s, "100000000000010000.0亿"},
		{"zero", Decimal{decimal.NewFromInt(0)}, "0"},
		{"one", Decimal{decimal.NewFromInt(1)}, "1"},
		{"one thousand", Decimal{decimal.NewFromInt(1000)}, "1000"},
		{"nine thousand nine hundred ninety-nine", Decimal{decimal.NewFromInt(9999)}, "9999"},
		{"greater than ten thousand", Decimal{decimal.NewFromInt(10001)}, "1.00万"},
		{"ten million", Decimal{decimal.NewFromInt(10000000)}, "1000.0万"},
		{"ninety-nine million nine hundred ninety-nine thousand nine hundred ninety-nine", Decimal{decimal.NewFromInt(99999999)}, "10000.0万"},
		{"one hundred million", Decimal{decimal.NewFromInt(100000000)}, "1.00亿"},
		{"ten billion", Decimal{decimal.NewFromInt(10000000000)}, "100.0亿"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.d.StringStatsIntCN(); got != tt.want {
				t.Errorf("input:%v StringStatsInt() = %v, want %v", tt.d, got, tt.want)
			}
		})
	}
}

func TestStringStatsDec(t *testing.T) {
	// 规则5 大数统计-带小数
	tests := []struct {
		name string
		d    Decimal
		want string
	}{
		{"less than 1000", Decimal{decimal.NewFromInt(999)}, "999.00"},
		{"equal to -1000", Decimal{decimal.NewFromInt(-1000)}, "-1.00K"},
		{"equal to 1000", Decimal{decimal.NewFromInt(1000)}, "1.00K"},
		{"between -1000 and -1 million", Decimal{decimal.NewFromInt(-123456)}, "-123.5K"},
		{"between 1000 and 1 million", Decimal{decimal.NewFromInt(123456)}, "123.5K"},
		{"equal to -1 million", Decimal{decimal.NewFromInt(-1000000)}, "-1.00M"},
		{"equal to 1 million", Decimal{decimal.NewFromInt(1000000)}, "1.00M"},
		{"between -1 million and -1 billion", Decimal{decimal.NewFromInt(-123456789)}, "-123.5M"},
		{"between 1 million and 1 billion", Decimal{decimal.NewFromInt(123456789)}, "123.5M"},
		{"equal to -1 billion", Decimal{decimal.NewFromInt(-1000000000)}, "-1.00B"},
		{"equal to 1 billion", Decimal{decimal.NewFromInt(1000000000)}, "1.00B"},
		{"between -1 billion and -1 trillion", Decimal{decimal.NewFromInt(-123456789012)}, "-123.5B"},
		{"between 1 billion and 1 trillion", Decimal{decimal.NewFromInt(123456789012)}, "123.5B"},
		{"equal to -1 trillion", Decimal{decimal.NewFromInt(-1000000000000)}, "-1.00T"},
		{"equal to 1 trillion", Decimal{decimal.NewFromInt(1000000000000)}, "1.00T"},
		{"less than -1 trillion", Decimal{decimal.NewFromInt(-1234567890123456)}, "-1234.6T"},
		{"greater than 1 trillion", Decimal{decimal.NewFromInt(1234567890123456)}, "1234.6T"},
		{"less than 1000", Decimal{decimal.NewFromFloat(999)}, "999.00"},
		{"exactly 999.12", Decimal{decimal.NewFromFloat(999.12)}, "999.12"},
		{"rounded 999.1267", Decimal{decimal.NewFromFloat(999.1267)}, "999.13"},
		{"slightly above 1000", Decimal{decimal.NewFromFloat(1000.4567)}, "1.00K"},
		{"just below 1 million", Decimal{decimal.NewFromFloat(999000.789)}, "999.0K"},
		{"just above 1 million", Decimal{decimal.NewFromFloat(1000000.891)}, "1.00M"},
		{"close to 900 million", Decimal{decimal.NewFromFloat(899999999.1234)}, "900.0M"},
		{"close to 1 billion", Decimal{decimal.NewFromFloat(999999999.1234)}, "1000.0M"},
		{"just above 1 billion", Decimal{decimal.NewFromFloat(1000000000.5678)}, "1.00B"},
		{"close to 900 billion", Decimal{decimal.NewFromFloat(899999999999.4321)}, "900.0B"},
		{"close to 1 trillion", Decimal{decimal.NewFromFloat(999999999999.4321)}, "1000.0B"},
		{"just above 1 trillion", Decimal{decimal.NewFromFloat(1000000000000.8765)}, "1.00T"},
		{"negative value", Decimal{decimal.NewFromFloat(-1234.5678)}, "-1.23K"},
		{"five million with precision", Decimal{decimal.NewFromFloat(5000000.123456789)}, "5.00M"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.d.StringStatsDec(); got != tt.want {
				t.Errorf("input:%v StringStatsInt() = %v, want %v", tt.d, got, tt.want)
			}
		})
	}
}

func TestDecimal_StringStatsDecCN(t *testing.T) {
	// 规则5 大数统计-带小数 中文
	tests := []struct {
		name string
		d    Decimal
		want string
	}{
		{"less than 10,000", Decimal{decimal.NewFromInt(9999)}, "9999.00"},
		{"equal to -10,000", Decimal{decimal.NewFromInt(-10000)}, "-1.00万"},
		{"equal to 10,000", Decimal{decimal.NewFromInt(10000)}, "1.00万"},
		{"between 10,000 and 100,000,000", Decimal{decimal.NewFromInt(50000)}, "5.00万"},
		{"equal to 100,000,000", Decimal{decimal.NewFromInt(100000000)}, "1.00亿"},
		{"between 100,000,000 and 1,000,000,000,000", Decimal{decimal.NewFromInt(123456789)}, "1.23亿"},
		{"greater than 100,000,000", Decimal{decimal.NewFromInt(123456789)}, "1.23亿"},
		{"less than -1,000,000,000,000", Decimal{decimal.NewFromInt(-123456789)}, "-1.23亿"},
		{"greater than 1,000,000,000,000", Decimal{decimal.NewFromInt(1234567890123456)}, "12345678.9亿"},
		{"zero", Decimal{decimal.NewFromInt(0)}, "0.00"},
		{"one", Decimal{decimal.NewFromInt(1)}, "1.00"},
		{"one thousand", Decimal{decimal.NewFromInt(1000)}, "1000.00"},
		{"nine thousand nine hundred ninety-nine", Decimal{decimal.NewFromInt(9999)}, "9999.00"},
		{"greater than ten thousand", Decimal{decimal.NewFromInt(10001)}, "1.00万"},
		{"ten million", Decimal{decimal.NewFromInt(10000000)}, "1000.0万"},
		{"ninety-nine million nine hundred ninety-nine thousand nine hundred ninety-nine", Decimal{decimal.NewFromInt(99999999)}, "10000.0万"},
		{"one hundred million", Decimal{decimal.NewFromInt(100000000)}, "1.00亿"},
		{"ten billion", Decimal{decimal.NewFromInt(10000000000)}, "100.0亿"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.d.StringStatsDecCN(); got != tt.want {
				t.Errorf("input:%v StringStatsInt() = %v, want %v", tt.d, got, tt.want)
			}
		})
	}
}

func TestStringInt(t *testing.T) {
	// 规则6 整数
	tests := []struct {
		name     string
		decimal  Decimal
		expected string
	}{
		{"equal to 1", NewFromFloat(1), "1"},
		{"equal to 0", NewFromFloat(0), "0"},
		{"equal to -1", NewFromFloat(-1.0), "-1"},
		{"more than 1", NewFromFloat(2.345), "2"},
		{"2 decimal places", NewFromFloat(3.45), "3"},
		{"equal to 50", NewFromFloat(50), "50"},
		{"equal to 0", NewFromFloat(0), "0"},
		{"equal to 100000000.123", NewFromFloat(100000000.123), "100000000"},
		{"equal to 123.956", NewFromFloat(123.956), "123"},
		{"equal to -789.1", NewFromFloat(-789.1), "-789"},
		{"equal to 123456789.123", NewFromFloat(123456789.123), "123456789"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.decimal.StringInt()
			assert.Equal(t, tt.expected, actual, tt.decimal)
		})
	}
}

func TestStringDecLow(t *testing.T) {
	// 规则7 小数-低精度
	tests := []struct {
		name     string
		decimal  Decimal
		expected string
	}{
		{"less than 1", NewFromFloat(0.123), "0.12"},
		{"equal to 0", NewFromFloat(0.0), "0"},
		{"equal to 1", NewFromFloat(1.0), "1"},
		{"equal to -1", NewFromFloat(-1.0), "-1"},
		{"equal to 1.1", NewFromFloat(1.10), "1.1"},
		{"less than 1.1", NewFromFloat(1.09), "1.09"},
		{"greater than 1", NewFromFloat(2.345), "2.34"},
		{"less than -1.1", NewFromFloat(-2.346), "-2.34"},
		{"less than -1", NewFromFloat(-2.342), "-2.34"},
		{"2 decimal places", NewFromFloat(3.45), "3.45"},
		{"more than 2 decimal places", NewFromFloat(4.56789), "4.56"},
		{"more than 2 decimal places with 0", NewFromFloat(4.5000), "4.5"},

		{"less than 1", NewFromFloat(0.12), "0.12"},
		{"equal to 1", NewFromFloat(1.0), "1"},
		{"equal to 1.1", NewFromFloat(1.10), "1.1"},
		{"greater than 1", NewFromFloat(2.345), "2.34"},
		{"2 decimal places", NewFromFloat(3.45), "3.45"},
		{"more than 2 decimal places", NewFromFloat(4.56789), "4.56"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.decimal.StringDecLow()
			if actual != tt.expected {
				t.Errorf("StringDecLow() = %q, want %q", actual, tt.expected)
			}
		})
	}
}

func TestDecimal_StringDecMid(t *testing.T) {
	// 规则8 小数-中精度
	tests := []struct {
		name string
		d    Decimal
		want string
	}{
		{"more than 6 decimal places", Decimal{decimal.NewFromFloat(1.123456789)}, "1.123456"},
		{"exactly 6 decimal places", Decimal{decimal.NewFromFloat(1.123456)}, "1.123456"},
		{"less than 6 decimal places", Decimal{decimal.NewFromFloat(1.12)}, "1.12"},
		{"integer", Decimal{decimal.NewFromInt(1)}, "1"},
		{"zero", Decimal{decimal.NewFromInt(0)}, "0"},
		{"more than 1 with zero", Decimal{decimal.NewFromFloat(0.100000)}, "0.1"},
		{"exactly 6 decimal places", NewFromFloat(0.123456), "0.123456"},
		{"negative with more than 6 decimal places", NewFromFloat(-123.456789), "-123.456789"},
		{"large number with more than 6 decimal places", NewFromFloat(1000.123456789012), "1000.123456"},
		{"zero with multiple decimal places", NewFromFloat(0.0000), "0"},
		{"small number close to 1", NewFromFloat(1.0000009), "1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.d.StringDecMid(); got != tt.want {
				t.Errorf("StringDecMid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecimal_StringDecHigh(t *testing.T) {
	// 规则9 小数-高精度
	tests := []struct {
		name     string
		decimal  Decimal
		expected string
	}{
		{"integer", NewFromInt(123), "123"},
		{"less than 8 places", NewFromFloat(123.456), "123.456"},
		{"exactly 8 places", NewFromFloat(123.456789), "123.456789"},
		{"more than 8 places", NewFromFloat(123.456789012), "123.45678901"},
		{"more than 8 places with 0", NewFromFloat(123.456789000), "123.456789"},
		{"more than 8 places with 0", NewFromFloat(123.456780000), "123.45678"},
		{"zero", NewFromFloat(0), "0"},
		{"exactly 8 decimal places", NewFromFloat(0.12345678), "0.12345678"},
		{"negative with more than 8 decimal places", NewFromFloat(-123.45678901), "-123.45678901"},
		{"large number with more than 8 decimal places", NewFromFloat(1000.123456789012), "1000.12345678"},
		{"zero with multiple decimal places", NewFromFloat(0.0000), "0"},
		{"small number with 7 decimal places", NewFromFloat(1.0000009), "1.0000009"},
		{"small number with 8 decimal places", NewFromFloat(1.00000009), "1.00000009"},
		{"small number with more than 8 places truncated", NewFromFloat(1.000000009), "1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.decimal.StringDecHigh()
			if actual != tt.expected {
				t.Errorf("StringDecHigh() = %q, want %q", actual, tt.expected)
			}
		})
	}
}

func TestDecimal_StringUltraHigh(t *testing.T) {
	// 规则10 超高精度
	tests := []struct {
		name     string
		decimal  Decimal
		expected string
	}{
		{"no decimal point", Decimal{decimal.NewFromInt(123)}, "123"},
		{"zero", Decimal{decimal.NewFromFloat(0)}, "0"},
		{"decimal part with less than 8 digits", Decimal{decimal.NewFromFloat(123.4567)}, "123.4567"},
		{"decimal part with exactly 8 digits", Decimal{decimal.NewFromFloat(123.456789)}, "123.456789"},
		{"decimal part with exactly 2 digits with 0", Decimal{decimal.NewFromFloat(123.4500)}, "123.45"},
		{"decimal part with more than 8 digits", Decimal{decimal.NewFromFloat(123.456789012)}, "123.45678901"},
		{"edge case: decimal part with 7 zeros", Decimal{decimal.NewFromFloat(0.00000001)}, "0.00000001"},
		{"edge case: decimal part with 8 zeros", Decimal{decimal.NewFromFloat(0.000000001)}, "0.000000001"},
		{"edge case: decimal part with 11 zeros", Decimal{decimal.NewFromFloat(0.000000000001)}, "0.000000000001"},
		{"edge case: decimal part with 12 zeros", Decimal{decimal.NewFromFloat(0.0000000000001)}, "0"},
		{"edge case: decimal part with 13 zeros", Decimal{decimal.NewFromFloat(1.00000000000001)}, "1"},
		{"edge case: decimal part with 14 zeros", Decimal{decimal.NewFromFloat(0.000000000000001)}, "0"},
		{"exactly 8 decimal places", NewFromFloat(0.12345678), "0.12345678"},
		{"more than 8 decimal places rounded", NewFromFloat(0.123456789), "0.12345678"},
		{"small number with 9 decimal places rounded to 8", NewFromFloat(0.000000123), "0.00000012"},
		{"zero", NewFromFloat(0), "0"},
		{"small number with 2 decimal places", NewFromFloat(0.01), "0.01"},
		{"small number with trailing zeros", NewFromFloat(0.010), "0.01"},
		{"negative number with small decimal part", NewFromFloat(-0.000000012345), "-0.00000001"},
		{"small number with 12 decimal places", NewFromFloat(0.000000000012345678), "0.000000000012"},
		{"small number with 14 decimal places truncated", NewFromFloat(0.00000000000012345678), "0"},
		{"small number with trailing zeros and 9 decimal places", NewFromFloat(0.00000000100), "0.000000001"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.decimal.StringUltraHigh()
			if actual != tt.expected {
				t.Errorf("StringUltraHigh() = %q, want %q", actual, tt.expected)
			}
		})
	}
}

func TestStringParcentInt(t *testing.T) {
	// 规则11 百分比-整数
	tests := []struct {
		name     string
		decimal  Decimal
		expected string
	}{
		{"positive", NewFromFloat(12.34), "1234%"},
		{"negative", NewFromFloat(-12.34), "-1234%"},
		{"zero", NewFromFloat(0), "0%"},
		{"more than 2 decimal places", NewFromFloat(12.3456), "1234%"},
		{"more than 2 decimal places", NewFromFloat(12.3456), "1234%"},
		{"half", NewFromFloat(0.50), "50%"},
		{"zero", NewFromFloat(0), "0%"},
		{"small value rounded down", NewFromFloat(0.0099), "0%"},
		{"ten multiplied", NewFromFloat(10), "1000%"},
		{"small negative value", NewFromFloat(-0.1), "-10%"},
		{"close to 1", NewFromFloat(0.999), "99%"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.decimal.StringParcentInt()
			if actual != tt.expected {
				t.Errorf("input %q StringParcentInt() = %q, want %q", tt.decimal, actual, tt.expected)
			}
		})
	}
}

func TestStringParcentLow(t *testing.T) {
	// 规则12 百分比-低精度
	tests := []struct {
		name     string
		decimal  Decimal
		expected string
	}{
		{"positive", NewFromFloat(12.34), "1234%"},
		{"negative", NewFromFloat(-12.34), "-1234%"},
		{"zero", NewFromFloat(0), "0%"},
		{"more than 2 decimal places", NewFromFloat(12.3456), "1234.56%"},
		{"exactly 2 decimal places", NewFromFloat(12.34), "1234%"},
		{"less than 2 decimal places", NewFromFloat(12.3), "1230%"},
		{"more than 2 decimal places with 0", NewFromFloat(12.3000), "1230%"},
		{"small positive value", NewFromFloat(0.0012), "0.12%"},
		{"small negative value", NewFromFloat(-1.23456), "-123.45%"},
		{"large positive value", NewFromFloat(1.00123456789), "100.12%"},
		{"slightly over 1", NewFromFloat(1.00109), "100.1%"},
		{"slightly more precision", NewFromFloat(1.00119), "100.11%"},
		{"almost exactly 1", NewFromFloat(1.000001), "100%"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.decimal.StringParcentLow()
			if actual != tt.expected {
				t.Errorf("StringParcentLow() = %q, want %q", actual, tt.expected)
			}
		})
	}
}

func TestStringParcentMid(t *testing.T) {
	// 规则13 百分比-中精度
	tests := []struct {
		name     string
		decimal  Decimal
		expected string
	}{
		{"positive more than 4 decimal places", NewFromFloat(12.3456), "1234.5600%"},
		{"positive exactly 4 decimal places", NewFromFloat(12.3456), "1234.5600%"},
		{"positive less than 4 decimal places", NewFromFloat(12.34), "1234.0000%"},
		{"negative more than 4 decimal places", NewFromFloat(-12.3456), "-1234.5600%"},
		{"negative exactly 4 decimal places", NewFromFloat(-12.3456), "-1234.5600%"},
		{"negative less than 4 decimal places", NewFromFloat(-12.3400), "-1234.0000%"},
		{"zero exactly 4 decimal places", NewFromFloat(0.00), "0.0000%"},
		{"zero", NewFromFloat(0), "0.0000%"},
		{"more than 4 decimal places with 0", NewFromFloat(0.0000), "0.0000%"},
		{"less than 4 decimal places with 0", NewFromFloat(0.1000), "10.0000%"},

		{"small positive value with high precision", NewFromFloat(0.0012345678), "0.1234%"},
		{"small negative value with high precision", NewFromFloat(-1.2345678), "-123.4567%"},
		{"exactly zero", NewFromFloat(0.00), "0.0000%"},
		{"small value with more than 4 decimal places rounded down", NewFromFloat(0.0000009), "0.0000%"},
		{"small value with 4 decimal places", NewFromFloat(0.0000099), "0.0009%"},
		{"small value with 3 decimal places", NewFromFloat(0.00009), "0.0090%"},
		{"large positive value with high precision", NewFromFloat(1.00123456789), "100.1234%"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.decimal.StringParcentMid()
			if actual != tt.expected {
				t.Errorf("StringParcentMid() = %q, want %q", actual, tt.expected)
			}
		})
	}
}

func TestStringParcentHigh(t *testing.T) {
	// 规则14 百分比-高精度
	tests := []struct {
		name     string
		decimal  Decimal
		expected string
	}{
		{"positive more than 6 decimal places", NewFromFloat(12.3456789), "1234.567890%"},
		{"positive exactly 6 decimal places", NewFromFloat(12.345678), "1234.567800%"},
		{"positive less than 6 decimal places", NewFromFloat(12.34), "1234.000000%"},
		{"negative more than 6 decimal places", NewFromFloat(-12.3456789), "-1234.567890%"},
		{"negative exactly 6 decimal places", NewFromFloat(-12.345678), "-1234.567800%"},
		{"negative less than 6 decimal places", NewFromFloat(-12.3400), "-1234.000000%"},
		{"zero exactly 6 decimal places", NewFromFloat(0.00), "0.000000%"},
		{"zero", NewFromFloat(0), "0.000000%"},
		{"small positive value with more than 6 decimal places", NewFromFloat(0.00123456), "0.123456%"},
		{"small negative value with more than 6 decimal places", NewFromFloat(-1.23456789), "-123.456789%"},
		{"large positive value with high precision", NewFromFloat(1.00123456789012), "100.123456%"},
		{"exactly zero", NewFromFloat(0.00), "0.000000%"},
		{"small value with more than 6 decimal places", NewFromFloat(0.000000900), "0.000090%"},
		{"small value with more than 8 decimal places", NewFromFloat(0.000000009), "0.000000%"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.decimal.StringParcentHigh()
			if actual != tt.expected {
				t.Errorf("StringParcentHigh() = %q, want %q", actual, tt.expected)
			}
		})
	}
}

func TestTruncateToDecimalPlaces(t *testing.T) {
	// 单元测试
	tests := []struct {
		name     string
		d        *Decimal
		places   int
		expected string
	}{
		{"no decimal point", &Decimal{decimal.NewFromInt(123)}, 0, "123"},
		{"decimal point with no decimal places", &Decimal{decimal.NewFromFloat(123.456)}, 0, "123"},
		{"decimal point with decimal places less than or equal to the specified places", &Decimal{decimal.NewFromFloat(123.456)}, 2, "123.45"},
		{"decimal point with decimal places greater than the specified places", &Decimal{decimal.NewFromFloat(123.456)}, 2, "123.45"},
		{"edge case: places is 0", &Decimal{decimal.NewFromFloat(123.456)}, 0, "123"},
		{"edge case: paces is 3", &Decimal{decimal.NewFromFloat(123.456)}, 3, "123.456"},
		{"edge case: zero ", &Decimal{decimal.NewFromFloat(123.0001)}, 2, "123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := truncateToDecimalPlaces(tt.d, tt.places, false)
			if actual != tt.expected {
				t.Errorf("truncateToDecimalPlaces() = %v, want %v", actual, tt.expected)
			}
		})
	}
}

func TestDecimal_StringStatsDecTW(t *testing.T) {
	tests := []struct {
		name string
		d    Decimal
		want string
	}{
		{"less than 10,000", Decimal{decimal.NewFromInt(9999)}, "9999.00"},
		{"equal to -10,000", Decimal{decimal.NewFromInt(-10000)}, "-1.00萬"},
		{"equal to 10,000", Decimal{decimal.NewFromInt(10000)}, "1.00萬"},
		{"between 10,000 and 100,000,000", Decimal{decimal.NewFromInt(50000)}, "5.00萬"},
		{"equal to 100,000,000", Decimal{decimal.NewFromInt(100000000)}, "1.00億"},
		{"between 100,000,000 and 1,000,000,000", Decimal{decimal.NewFromInt(500000000)}, "5.00億"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.d.StringStatsDecTW(); got != tt.want {
				t.Errorf("input:%v StringStatsInt() = %v, want %v", tt.d, got, tt.want)
			}
		})
	}
}

func TestDecimal_StringStatsIntTW(t *testing.T) {
	// 规则5 大数统计-带小数 中文
	tests := []struct {
		name string
		d    Decimal
		want string
	}{
		{"less than 10,000", Decimal{decimal.NewFromInt(9999)}, "9999"},
		{"equal to -10,000", Decimal{decimal.NewFromInt(-10000)}, "-1.00萬"},
		{"equal to 10,000", Decimal{decimal.NewFromInt(10000)}, "1.00萬"},
		{"between 10,000 and 100,000,000", Decimal{decimal.NewFromInt(50000)}, "5.00萬"},
		{"equal to 100,000,000", Decimal{decimal.NewFromInt(100000000)}, "1.00億"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.d.StringStatsIntTW(); got != tt.want {
				t.Errorf("input:%v StringStatsInt() = %v, want %v", tt.d, got, tt.want)
			}
		})
	}
}
