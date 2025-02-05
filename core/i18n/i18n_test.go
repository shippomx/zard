package i18n

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shippomx/zard/i18n/i18ngo"
	"github.com/stretchr/testify/assert"
)

func TestMiddlewareHandler(t *testing.T) {
	tests := []struct {
		name           string
		i18n           *I18n
		translator     i18ngo.Translator
		w              http.ResponseWriter
		r              *http.Request
		acceptLanguage string
	}{
		{
			name: "nil I18n instance",
			i18n: nil,
		},
		{
			name:       "nil translator",
			i18n:       &I18n{translator: nil},
			translator: nil,
		},
		{
			name: "nil ResponseWriter",
			i18n: &I18n{
				translator: i18ngo.GetTranslator(),
			},
			w: nil,
		},
		{
			name: "empty Accept-Language header",
			i18n: &I18n{
				translator: i18ngo.GetTranslator(),
			},
			w: httptest.NewRecorder(),
			r: &http.Request{
				Header: http.Header{},
			},
		},
		{
			name: "unknown language",
			i18n: &I18n{
				translator: i18ngo.GetTranslator(),
			},
			w: httptest.NewRecorder(),
			r: &http.Request{
				Header: http.Header{
					"Accept-Language": []string{"unknown"},
				},
			},
		},
		{
			name: "known language",
			i18n: &I18n{
				translator: i18ngo.GetTranslator(),
			},
			w: httptest.NewRecorder(),
			r: &http.Request{
				Header: http.Header{
					"Accept-Language": []string{"en"},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			next := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
			})
			handler := test.i18n.MiddlewareHandler(next)
			if test.w == nil {
				test.w = httptest.NewRecorder()
			}
			handler(test.w, test.r)
			if test.acceptLanguage != "" {
				ctx := test.r.Context()
				translator, ok := ctx.Value(TranslatorKey).(EmbeddedTranslator)
				if !ok {
					t.Errorf("expected EmbeddedTranslator in context, got %T", ctx.Value(TranslatorKey))
				}
				if translator.translator != test.translator {
					t.Errorf("expected translator %s, got %s", test.translator, translator.translator)
				}
			}
		})
	}
}

func TestI18n_SetLocaleMap(t *testing.T) {
	i := &I18n{}
	localeMap := map[string]i18ngo.LocaleT{
		"en": i18ngo.AR,
		"fr": i18ngo.FR,
	}
	// Test that the localeMap is set correctly
	i.SetLocaleMap(localeMap)
	assert.Equal(t, localeMap, i.localeMap)
	localeMap2 := map[string]i18ngo.LocaleT{}
	// Test that the localeMap is set correctly
	i.SetLocaleMap(localeMap2)
	assert.Equal(t, localeMap2, i.localeMap)
}

func TestI18n_SetHeaders(t *testing.T) {
	i := &I18n{}
	headers := []string{"Accept-Language", "Content-Language"}
	// Test that the headers are set correctly
	i.SetHeaders(headers)
	assert.Equal(t, headers, i.headers)
	// Test that the headers are set correctly
	i.SetHeaders([]string{})
	assert.Equal(t, []string{}, i.headers)
}

func TestI18n_SetHeader(t *testing.T) {
	i := &I18n{}
	header := "Accept-Language"
	// Test that the header is set correctly
	i.SetHeader(header)
	assert.Equal(t, []string{header}, i.headers)
	// Test that the header is set correctly
	i.SetHeader("test")
	assert.Equal(t, []string{"Accept-Language", "test"}, i.headers)
}

func TestI18n_SetHeaderAndLocaleMap(t *testing.T) {
	i := &I18n{}
	header := "Accept-Language"
	localeMap := map[string]i18ngo.LocaleT{
		"en": i18ngo.AR,
		"fr": i18ngo.FR,
	}
	// Test that the header is set correctly
	i.SetHeader(header)
	assert.Equal(t, []string{header}, i.headers)
	// Test that the localeMap is set correctly
	i.SetLocaleMap(localeMap)
	assert.Equal(t, localeMap, i.localeMap)
}

func TestI18nLocalString(t *testing.T) {
	assert.Equal(t, "en", string(i18ngo.LocaleT("en")))
	assert.Equal(t, "", string(i18ngo.LocaleT("")))
}
