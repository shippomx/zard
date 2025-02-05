package http

import (
	"context"
	"encoding/xml"
	"fmt"
	"go/types"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/shippomx/zard/core/i18n"
	"github.com/shippomx/zard/core/logx"
	"github.com/shippomx/zard/rest/errors"
	"github.com/shippomx/zard/rest/httpx"
)

var i18nWrap atomic.Bool

var ErrWrap atomic.Bool

func EnableI18nWrap() {
	i18nWrap.Store(true)
}

func DisableErrWrap() {
	ErrWrap.Store(true)
}

func initNewErrorHandle() {
	if httpx.IsErrorHandlerInited() {
		return
	}
	httpx.SetErrorHandlerCtx(func(ctx context.Context, err error) (int, any) {
		if !ErrWrap.Load() {
			err = errors.WrapErr(err)
		}
		if i18nWrap.Load() {
			err = errors.WrapErrWithI18n(err)
		}

		switch e := err.(type) {

		case *errors.CodeMsg:
			var label *string
			if e.Label != "" {
				label = &e.Label
			}
			statusCode := http.StatusOK
			if e.HTTPStatusCode != 0 {
				statusCode = e.HTTPStatusCode
			}
			message := e.Message
			if e.I18nEnable {
				newmessage, err := i18n.LocalizeSprintf(ctx, message, e.I18nArgs...)
				if err != nil {
					return statusCode, BaseResponse[types.Nil]{
						Code:      e.Code,
						Message:   message,
						Extra:     "i18n err: " + err.(*errors.CodeMsg).Extra.(string),
						Timestamp: time.Now().Unix(),
						Label:     label,
					}
				}
				message = newmessage
			}
			return statusCode, BaseResponse[types.Nil]{
				Code:      e.Code,
				Message:   message,
				Timestamp: time.Now().Unix(),
				Extra:     e.Extra,
				Label:     label,
			}
		default:
			return http.StatusOK, BaseResponse[types.Nil]{
				Code:      -1,
				Message:   err.Error(),
				Timestamp: time.Now().Unix(),
			}
		}
	})
}

// OkXml writes v into w with 200 OK.
func OkXml(w http.ResponseWriter, v any) {
	WriteXml(w, http.StatusOK, v)
}

// OkXmlCtx writes v into w with 200 OK.
func OkXmlCtx(ctx context.Context, w http.ResponseWriter, v any) {
	WriteXmlCtx(ctx, w, http.StatusOK, v)
}

// WriteXml writes v as xml string into w with code.
func WriteXml(w http.ResponseWriter, code int, v any) {
	if err := doWriteXml(w, code, v); err != nil {
		logx.Error(err)
	}
}

// WriteXmlCtx writes v as xml string into w with code.
func WriteXmlCtx(ctx context.Context, w http.ResponseWriter, code int, v any) {
	if err := doWriteXml(w, code, v); err != nil {
		logx.WithContext(ctx).Error(err)
	}
}

func doWriteXml(w http.ResponseWriter, code int, v any) error {
	bs, err := xml.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return fmt.Errorf("marshal xml failed, error: %w", err)
	}

	w.Header().Set(httpx.ContentType, XmlContentType)
	w.WriteHeader(code)

	if n, err := w.Write(bs); err != nil {
		// http.ErrHandlerTimeout has been handled by http.TimeoutHandler,
		// so it's ignored here.
		if err != http.ErrHandlerTimeout {
			return fmt.Errorf("write response failed, error: %w", err)
		}
	} else if n < len(bs) {
		return fmt.Errorf("actual bytes: %d, written bytes: %d", len(bs), n)
	}

	return nil
}
