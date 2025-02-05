package httpc

import "errors"

const (
	pathKey   = "path"
	formKey   = "form"
	headerKey = "header"
	jsonKey   = "json"
	slash     = "/"
	colon     = ':'
)

var ErrGetWithBody = errors.New("HTTP GET should not have body")
