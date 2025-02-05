package errors

import (
	"fmt"
	"net/http"
)

// CodeMsg is a struct that contains a code and a message.
// It implements the error interface.
type CodeMsg struct {
	Code           int
	Message        string
	Label          string
	HTTPStatusCode int
	Extra          any
	I18nEnable     bool
	I18nArgs       []string
}

type Option func(*CodeMsg)

func WithLabel(label string) Option {
	return func(c *CodeMsg) {
		c.Label = label
	}
}

func WithI18n(args ...string) Option {
	return func(c *CodeMsg) {
		c.I18nEnable = true
		c.I18nArgs = args
	}
}

func WithHTTPStatusCode(code int) Option {
	return func(c *CodeMsg) {
		c.HTTPStatusCode = code
	}
}

func WithExtra(extra any) Option {
	return func(c *CodeMsg) {
		c.Extra = extra
	}
}

func (c *CodeMsg) Error() string {
	extraString := ""
	if c.Extra != nil {
		if extra, ok := c.Extra.(string); ok {
			extraString = ", extra: " + extra
		}
	}
	labelString := ""
	if c.Label != "" {
		labelString = ", label: " + c.Label
	}

	return fmt.Sprintf("code: %d, msg: %s", c.Code, c.Message) + labelString + extraString
}

// New creates a new CodeMsg.
func New(code int, msg string, opts ...Option) error {
	cm := &CodeMsg{Code: code, Message: msg}
	for _, opt := range opts {
		opt(cm)
	}
	return cm
}

func WrapErrWithI18n(err error, arg ...string) error {
	switch v := err.(type) {
	case *CodeMsg:
		if !v.I18nEnable {
			v.I18nEnable = true
			v.I18nArgs = arg
		}
		return v
	default:
		return &CodeMsg{
			Code:           -1,
			Message:        err.Error(),
			I18nEnable:     true,
			I18nArgs:       arg,
			HTTPStatusCode: http.StatusOK,
		}
	}
}

func WrapErr(err error) error {
	switch v := err.(type) {
	case *CodeMsg:
		return v
	case nil:
		return &CodeMsg{
			Code:           -1,
			HTTPStatusCode: http.StatusOK,
		}
	default:
		return &CodeMsg{
			Code:           -1,
			Message:        err.Error(),
			HTTPStatusCode: http.StatusOK,
		}
	}
}
