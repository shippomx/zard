package golang

import (
	goformat "go/format"

	gofumpt "mvdan.cc/gofumpt/format"
)

func FormatCode(code string) string {
	ret, err := goformat.Source([]byte(code))
	if err != nil {
		return code
	}

	return string(ret)
}

func FormatCodeByGofumpt(code, rootpkg string) (string, error) {
	got, err := gofumpt.Source([]byte(code), gofumpt.Options{LangVersion: "go1.21", ModulePath: rootpkg})
	if err != nil {
		return code, err
	}
	return string(got), err
}
