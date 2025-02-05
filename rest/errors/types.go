package errors

// IsPassWord 交易密码
func IsPassWord(msg string, passType int) error {
	extra := map[string]interface{}{
		"pass_type": passType,
	}

	return New(101, msg, WithExtra(extra))
}

// IsBio 生物验证
func IsBio(msg string, qrid string) error {
	extra := map[string]interface{}{
		"qrid": qrid,
	}

	return New(102, msg, WithExtra(extra))
}

// TokenInvalid 动态mtoken失效 401
func TokenInvalid(msg string) error {
	return New(401, msg)
}

// PassWordNotSet 交易密码未设置
func PassWordNotSet(msg string) error {
	return New(999, msg)
}

// NotLogin 登录态丢失或未登录
func NotLogin(msg string) error {
	return New(1001, msg)
}

// SecondVerify 二次验证
func SecondVerify(msg string) error {
	return New(1200, msg)
}

// DefiTokenInvalid defi token失效
func DefiTokenInvalid(msg string) error {
	return New(11002, msg)
}
