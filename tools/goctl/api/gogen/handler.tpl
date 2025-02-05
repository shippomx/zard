package {{.PkgName}}

import (
	"net/http"
	{{if or  (.HasRequest) (not .HasResp) }}
	"github.com/shippomx/zard/rest/httpx"{{end}}
	xhttp "github.com/shippomx/zard/rest/http"

	{{.ImportPackages}}
)
{{if .HasDoc}}{{.Doc}}{{end}}
func {{.HandlerName}}(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		{{if .HasRequest}}var req types.{{.RequestType}}
		if err := httpx.Parse(r, &req); err != nil {
			xhttp.JsonErrResponseCtx(r.Context(), w, err)
			return
		}

		{{end}}l := {{.LogicName}}.New{{.LogicType}}(r.Context(), svcCtx)
		{{if .HasResp}}resp, {{end}}err := l.{{.Call}}({{if .HasRequest}}&req{{end}})
		if err != nil {
			// code-data 响应格式
			xhttp.JsonErrResponseCtx(r.Context(), w, err)
		} else {
			// code-data 响应格式
			{{if .HasResp}}xhttp.JsonBaseResponseCtx(r.Context(), w, resp){{else}}httpx.Ok(w){{end}}
		}
	}
}
