package golang

import (
	"strings"
	"testing"
)

func TestFormatCodeByGofumpt(t *testing.T) {
	source := `package handler

import (
	"demo1/internal/logic"
	"demo1/internal/svc"
	"demo1/internal/types"
	"net/http"

	xhttp "github.com/shippomx/zard/rest/http"
	"github.com/shippomx/zard/rest/httpx"
)

func Demo1Handler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.Request
		if err := httpx.Parse(r, &req); err != nil {
			xhttp.JsonErrResponseCtx(r.Context(), w, err)
			return
		}

		l := logic.NewDemo1Logic(r.Context(), svcCtx)
		resp, err := l.Demo1(&req)
		if err != nil {
			// code-data 响应格式
			xhttp.JsonErrResponseCtx(r.Context(), w, err)
		} else {
			// code-data 响应格式
			xhttp.JsonBaseResponseCtx(r.Context(), w, resp)
		}
	}
}`
	res, err := FormatCodeByGofumpt(source, "demo1")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Index(res, "net/http") > strings.Index(res, "demo1/internal") {
		t.Fatal("format error")
		t.Log(res)
	}
}
