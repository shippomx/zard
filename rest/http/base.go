package http

import (
	"context"
	"encoding/xml"
	"net/http"
	"sync"
	"time"

	"github.com/shippomx/zard/rest/errors"
	"github.com/shippomx/zard/rest/httpx"
	"google.golang.org/grpc/status"
)

var once sync.Once

// BaseResponse is the base response struct.
type BaseResponse[T any] struct {
	// Code represents the business code, not the http status code.
	Code int `json:"code" xml:"code"`
	// Msg represents the business message, if Code = BusinessCodeOK,
	// and Msg is empty, then the Msg will be set to BusinessMsgOk.
	Message string `json:"message" xml:"message"`
	// Data represents the business data.
	Data T `json:"data,omitempty" xml:"data,omitempty"`
	// Time represents the timestamp.
	Timestamp int64 `json:"timestamp,omitempty" xml:"timestamp,omitempty"` // timestamp
	// apimV3/V2 历史规范兼容
	Label *string `json:"label,omitempty" xml:"label,omitempty"`
	// Page represents the current page.
	Page *int `json:"page,omitempty" xml:"page,omitempty"`
	// Limit represents the page size.
	Limit *int `json:"limit,omitempty" xml:"limit,omitempty"`
	// Total represents the total count.
	Total *int `json:"total,omitempty" xml:"total,omitempty"`
	// Extra represents the extra error information.
	Extra any `json:"extra,omitempty" xml:"extra,omitempty"`

	// Deprecated: PageSize represents the page size.
	PageSize *int `json:"pagesize,omitempty" xml:"pagesize,omitempty"`
	// Deprecated: Pagecount represents the total page count.
	Pagecount *int `json:"pagecount,omitempty" xml:"pagecount,omitempty"`
	// Deprecated: Totalcount represents the total count.
	Totalcount *int `json:"totalcount,omitempty" xml:"totalcount,omitempty"`
}

type Option func(response *BaseResponse[any])

type PagedData struct {
	Page  int
	Limit int
	Total int

	// Deprecated: PageSize represents the page size.
	PageSize int
	// Deprecated: Pagecount represents the total page count.
	Pagecount int
	// Deprecated: Totalcount represents the total count.
	Totalcount int
}

func WithPageData(pageData *PagedData) Option {
	if pageData == nil {
		return func(c *BaseResponse[any]) {}
	}
	return func(c *BaseResponse[any]) {
		c.Page = &pageData.Page
		c.Limit = &pageData.Limit
		if pageData.Total > 0 {
			c.Total = &pageData.Total
		}

		if pageData.PageSize > 0 {
			c.PageSize = &pageData.PageSize
		}
		if pageData.Pagecount > 0 {
			c.Pagecount = &pageData.Pagecount
		}
		if pageData.Totalcount > 0 {
			c.Totalcount = &pageData.Totalcount
		}
	}
}

func WithLabel(label string) Option {
	return func(c *BaseResponse[any]) {
		c.Label = &label
	}
}

func WithCode(code int) Option {
	return func(c *BaseResponse[any]) {
		c.Code = code
	}
}

type baseXmlResponse[T any] struct {
	XMLName  xml.Name `xml:"xml"`
	Version  string   `xml:"version,attr"`
	Encoding string   `xml:"encoding,attr"`
	BaseResponse[T]
}

// JsonBaseResponse writes v into w with http.StatusOK.
func JsonBaseResponse(w http.ResponseWriter, v any, opts ...Option) {
	httpx.OkJson(w, wrapBaseResponse(v, opts...))
}

// JsonBaseResponseCtx writes v into w with http.StatusOK.
func JsonBaseResponseCtx(ctx context.Context, w http.ResponseWriter, v any, opts ...Option) {
	httpx.OkJsonCtx(ctx, w, wrapBaseResponse(v, opts...))
}

// nolint: revive //Json prefix is default style
func JsonErrResponseCtx(ctx context.Context, w http.ResponseWriter, err error) {
	once.Do(initNewErrorHandle)
	httpx.ErrorCtx(ctx, w, err)
}

// XmlBaseResponse writes v into w with http.StatusOK.
func XmlBaseResponse(w http.ResponseWriter, v any) {
	OkXml(w, wrapXmlBaseResponse(v))
}

// XmlBaseResponseCtx writes v into w with http.StatusOK.
func XmlBaseResponseCtx(ctx context.Context, w http.ResponseWriter, v any) {
	OkXmlCtx(ctx, w, wrapXmlBaseResponse(v))
}

func wrapXmlBaseResponse(v any) baseXmlResponse[any] {
	base := wrapBaseResponse(v)
	return baseXmlResponse[any]{
		Version:      xmlVersion,
		Encoding:     xmlEncoding,
		BaseResponse: base,
	}
}

func wrapBaseResponse(v any, opts ...Option) BaseResponse[any] {
	var resp BaseResponse[any]
	switch data := v.(type) {
	case *errors.CodeMsg:
		resp.Code = data.Code
		resp.Message = data.Message
		if data.Label != "" {
			resp.Label = &data.Label
		}
		resp.Timestamp = time.Now().UnixMilli()
		if data.Extra != nil {
			resp.Extra = data.Extra
		}
	case errors.CodeMsg:
		resp.Code = data.Code
		resp.Message = data.Message
		if data.Label != "" {
			resp.Label = &data.Label
		}
		resp.Timestamp = time.Now().UnixMilli()
		if data.Extra != nil {
			resp.Extra = data.Extra
		}
	case *status.Status:
		resp.Code = int(data.Code())
		resp.Message = data.Message()
		resp.Timestamp = time.Now().UnixMilli()
	case error:
		resp.Code = BusinessCodeError
		resp.Message = data.Error()
		resp.Timestamp = time.Now().UnixMilli()
	default:
		resp.Code = BusinessCodeOK
		resp.Message = BusinessMsgOk
		resp.Timestamp = time.Now().UnixMilli()
		resp.Data = v
	}

	for _, opt := range opts {
		opt(&resp)
	}

	return resp
}
