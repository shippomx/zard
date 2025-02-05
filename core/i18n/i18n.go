package i18n

import (
	"context"
	"net/http"
	"slices"
	"sync"

	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/shippomx/zard/core/logx"
	"github.com/shippomx/zard/i18n/i18ngo"
	"github.com/shippomx/zard/rest/errors"
)

var StatusI18nBaseErr = 494

type I18n struct {
	translator    i18ngo.Translator
	headers       []string
	mu            sync.Mutex
	localeMap     map[string]i18ngo.LocaleT
	defaultLocale i18ngo.LocaleT
}

func (i *I18n) SetLocaleMap(localeMap map[string]i18ngo.LocaleT) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.localeMap = localeMap
}

func (i *I18n) SetHeader(header string) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.headers = append(i.headers, header)
}

func (i *I18n) SetHeaders(headers []string) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.headers = headers
}

type translatorKeyType struct{}

var TranslatorKey = translatorKeyType{}

func MustNewI18n(conf I18nConf) *I18n {
	logx.Must(conf.Validate())
	if conf.Timeout == 0 {
		conf.Timeout = 5000
	}
	if !slices.Contains(conf.TeamIDs, conf.TeamID) {
		conf.TeamIDs = append(conf.TeamIDs, conf.TeamID)
	}
	IDslice := ToStringSlice(conf.TeamIDs)
	translatorConfig := i18ngo.TranslatorConf{
		ClientConfig: constant.ClientConfig{
			TimeoutMs:   uint64(conf.Timeout.Milliseconds()),
			Username:    conf.Username,
			Password:    conf.Password,
			LogLevel:    conf.LogLevel,
			NamespaceId: conf.NamespaceID,
		},
		Servers: []constant.ServerConfig{
			{
				IpAddr: conf.IPAddr,
				Port:   conf.Port,
			},
		},
		Groups:        IDslice,
		DefaultLocale: i18ngo.LocaleT(conf.DefaultLocale),
	}

	logx.Must(i18ngo.SetTranslatorByConfig(&translatorConfig))
	translator := i18ngo.GetTranslator()
	return &I18n{
		translator: translator,
		localeMap: map[string]i18ngo.LocaleT{
			i18ngo.AR: i18ngo.AR,
			i18ngo.BR: i18ngo.BR,
			i18ngo.DE: i18ngo.DE,
			i18ngo.ES: i18ngo.ES,
			i18ngo.FR: i18ngo.FR,
			i18ngo.ID: i18ngo.ID,
			i18ngo.IT: i18ngo.IT,
			i18ngo.JA: i18ngo.JA,
			i18ngo.KR: i18ngo.KR,
			i18ngo.NL: i18ngo.NL,
			i18ngo.PT: i18ngo.PT,
			i18ngo.RU: i18ngo.RU,
			i18ngo.TH: i18ngo.TH,
			i18ngo.TR: i18ngo.TR,
			i18ngo.UK: i18ngo.UK,
			i18ngo.VI: i18ngo.VI,
			i18ngo.ZH: i18ngo.ZH,
			i18ngo.TW: i18ngo.TW,
			i18ngo.EN: i18ngo.EN,
		},
		headers:       []string{"Accept-Language"},
		defaultLocale: i18ngo.LocaleT(conf.DefaultLocale),
	}
}

type EmbeddedTranslator struct {
	translator i18ngo.Translator
	localeT    i18ngo.LocaleT
}

func (e *EmbeddedTranslator) SetLocaleT(l i18ngo.LocaleT) {
	e.localeT = l
}

func (i *I18n) MiddlewareHandler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if i == nil || i.translator == nil {
			next(w, r)
			return
		}

		acceptLang := string(i.defaultLocale)
		for _, header := range i.headers {
			acceptLang = r.Header.Get(header)
			if acceptLang != "" {
				break
			}
		}
		if acceptLang != "" {
			lang := i.localeMap[acceptLang]
			if lang == "" {
				lang = i18ngo.LocaleT(acceptLang)
			}
			ctx := context.WithValue(r.Context(), TranslatorKey, EmbeddedTranslator{
				translator: i.translator,
				localeT:    lang,
			})
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			next(w, r)
		}
	}
}

func (e *EmbeddedTranslator) LocalizeText(text string) (string, error) {
	res, err := e.translator.LocalizeText(text, e.localeT)
	if err != nil {
		return text, NewI18nHTTPError(text, err)
	}
	return res, nil
}

func (e *EmbeddedTranslator) LocalizeSprintf(text string, args ...string) (string, error) {
	res, err := e.translator.LocalizeSprintf(text, e.localeT, args...)
	if err != nil {
		return text, NewI18nHTTPError(text, err)
	}
	return res, nil
}

func LocalizeText(ctx context.Context, text string) (string, error) {
	if t, ok := ctx.Value(TranslatorKey).(EmbeddedTranslator); ok {
		res, err := t.LocalizeText(text)
		if err != nil {
			return text, err
		}
		return res, nil
	}
	return text, nil
}

func SetLocaleTCtx(ctx context.Context, locale i18ngo.LocaleT) error {
	if t, ok := ctx.Value(TranslatorKey).(EmbeddedTranslator); ok {
		t.SetLocaleT(locale)
	} else {
		return errors.New(StatusI18nBaseErr, "no existing translator", errors.WithHTTPStatusCode(http.StatusInternalServerError))
	}
	return nil
}

func LocalizeSprintf(ctx context.Context, text string, args ...string) (string, error) {
	if t, ok := ctx.Value(TranslatorKey).(EmbeddedTranslator); ok {
		res, err := t.LocalizeSprintf(text, args...)
		if err != nil {
			return text, err
		}
		return res, nil
	}
	return text, nil
}

func NewI18nHTTPError(source string, err error) error {
	return &errors.CodeMsg{
		Code:    StatusI18nBaseErr,
		Message: source,
		Extra:   err.Error(),
	}
}
