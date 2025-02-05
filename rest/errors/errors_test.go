package errors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	ast := assert.New(t)
	c := New(1, "test")
	cm, ok := c.(*CodeMsg)
	ast.True(ok)
	ast.NotNil(cm)
	ast.Equal(int(1), cm.Code)
	ast.Equal("test", cm.Message)
}

func TestCodeMsg_Error(t *testing.T) {
	ast := assert.New(t)
	c := New(1, "test")
	cm, ok := c.(*CodeMsg)
	ast.True(ok)
	ast.NotNil(cm)
	ast.NotEmpty(cm.Error())
}

func TestWithI18n(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected CodeMsg
	}{
		{
			name: "no arguments",
			args: []string{},
			expected: CodeMsg{
				I18nEnable: true,
				I18nArgs:   []string{},
			},
		},
		{
			name: "one argument",
			args: []string{"arg1"},
			expected: CodeMsg{
				I18nEnable: true,
				I18nArgs:   []string{"arg1"},
			},
		},
		{
			name: "multiple arguments",
			args: []string{"arg1", "arg2", "arg3"},
			expected: CodeMsg{
				I18nEnable: true,
				I18nArgs:   []string{"arg1", "arg2", "arg3"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			codeMsg := CodeMsg{}
			option := WithI18n(tt.args...)
			option(&codeMsg)

			if codeMsg.I18nEnable != tt.expected.I18nEnable {
				t.Errorf("I18nEnable = %v, want %v", codeMsg.I18nEnable, tt.expected.I18nEnable)
			}

			if !cmpSlice(codeMsg.I18nArgs, tt.expected.I18nArgs) {
				t.Errorf("I18nArgs = %v, want %v", codeMsg.I18nArgs, tt.expected.I18nArgs)
			}
		})
	}
}

func cmpSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func TestWithHTTPStatusCode(t *testing.T) {
	tests := []struct {
		name     string
		code     int
		expected CodeMsg
	}{
		{
			name: "sets HTTPStatusCode correctly",
			code: 200,
			expected: CodeMsg{
				HTTPStatusCode: 200,
			},
		},
		{
			name: "does not modify other fields",
			code: 404,
			expected: CodeMsg{
				Code:           0,
				Message:        "",
				Label:          "",
				Extra:          nil,
				I18nEnable:     false,
				I18nArgs:       nil,
				HTTPStatusCode: 404,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			codeMsg := CodeMsg{}
			option := WithHTTPStatusCode(tt.code)
			option(&codeMsg)
			if codeMsg.HTTPStatusCode != tt.expected.HTTPStatusCode {
				t.Errorf("expected HTTPStatusCode %d, got %d", tt.expected.HTTPStatusCode, codeMsg.HTTPStatusCode)
			}
		})
	}
}
