package i18ngo

import (
	"encoding/json"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestInitCache_JSONFile(t *testing.T) {
	translator := NacosTranslator{
		cache: make([]sync.Map, len(Locales)),
	}
	dataID := "en.json"
	config := `{"key": "value"}`

	err := translator.InitCache(dataID, config)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	// Add assertions for checking if the locale is stored in the cache
}

func TestInitCache_UnknownFileType(t *testing.T) {
	translator := NacosTranslator{}
	dataID := "unknown.txt"
	config := "random data"
	err := translator.InitCache(dataID, config)
	if err == nil {
		t.Error("Expected error for unknown file type, got nil")
	} else if err.Error() != "unknown file type" {
		t.Errorf("Expected 'unknown file type' error, got %v", err.Error())
	}
}

func TestLocalizeText_Found(t *testing.T) {
	translator := nacosTranslator

	text := "Hello"
	locale := "en"

	err := translator.UpdateLocale(LocaleT(locale), map[string]string{
		text: "Goodbye",
	})
	assert.NoError(t, err)
	result, err := translator.LocalizeText(text, LocaleT(locale))
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result != "Goodbye" {
		t.Errorf("Expected 'Hello', got %s", result)
	}
}

func TestLocalizeText_NotFound(t *testing.T) {
	translator := nacosTranslator

	text := "Goodbye"
	locale := "en"

	_, err := translator.LocalizeText(text, LocaleT(locale))
	assert.Contains(t, err.Error(), "cache not found for text Goodbye")
}

func TestLocalizeText_LocaleNotFound(t *testing.T) {
	translator := nacosTranslator

	text := "Hello"
	locale := "en1"

	_, err := translator.LocalizeText(text, LocaleT(locale))

	assert.Contains(t, err.Error(), "cache not found for locale en")
}

func TestStoreLocale_Success(t *testing.T) {
	translator := nacosTranslator

	locale := LocaleT("en")
	m := &sync.Map{}
	m.Store("key", "value")

	err := translator.StoreLocale(locale, m)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Check if the value was stored correctly
	v, err := translator.LocalizeText("key", locale)
	assert.Equal(t, "value", v)
	assert.NoError(t, err)
}

func TestStoreLocale_LocaleNotFound(t *testing.T) {
	translator := NacosTranslator{
		cache: make([]sync.Map, 2),
	}

	locale := LocaleT("fr")
	m := &sync.Map{}
	m.Store("key", "value")

	err := translator.StoreLocale(locale, m)

	assert.Contains(t, err.Error(), "overflow t.cache")

	translator = nacosTranslator
	locale = LocaleT("fr1")

	err = translator.StoreLocale(locale, m)
	assert.Contains(t, err.Error(), "locale not support")
}

func TestLocalizeText(t *testing.T) {
	translator := &NacosTranslator{
		cache: make([]sync.Map, len(Locales)),
	}

	// Test successful localization with existing locale and text
	translator.cache[0].Store("hello", "bonjour")
	result, err := translator.LocalizeText("hello", Locales[0])
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result != "bonjour" {
		t.Errorf("expected 'bonjour', got '%s'", result)
	}

	// Test localization with non-existent locale
	_, err = translator.LocalizeText("hello", "non-existent-locale")
	if err == nil {
		t.Errorf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "cache not found for locale") {
		t.Errorf("expected error message to contain 'cache not found for locale', got '%s'", err.Error())
	}

	// Test localization with non-existent text
	_, err = translator.LocalizeText("non-existent-text", Locales[0])
	if err == nil {
		t.Errorf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "cache not found for text") {
		t.Errorf("expected error message to contain 'cache not found for text', got '%s'", err.Error())
	}

	// Test localization with default locale
	translator.defaultType = Locales[0]
	translator.cache[0].Store("hello", "bonjour")
	result, err = translator.LocalizeText("hello", Locales[1])
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result != "bonjour" {
		t.Errorf("expected 'bonjour', got '%s'", result)
	}

	// Test localization with empty default locale
	translator.defaultType = ""
	_, err = translator.LocalizeText("hello", Locales[1])
	if err == nil {
		t.Errorf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "cache not found for text") {
		t.Errorf("expected error message to contain 'cache not found for text', got '%s'", err.Error())
	}

	// Test localization with default locale and non-existent text
	translator.defaultType = Locales[0]
	_, err = translator.LocalizeText("non-existent-text", Locales[1])
	if err == nil {
		t.Errorf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "cache not found for text") {
		t.Errorf("expected error message to contain 'cache not found for text', got '%s'", err.Error())
	}
}

type TranslatorTestSuite struct {
	suite.Suite
	configs []TranslatorConf
	groups  []string
	mu      sync.Mutex
}

func (s *TranslatorTestSuite) SetupSuite() {
	s.groups = []string{"unified", "php", "pay", "infra", "asset", "copy", "earn", "market", "platform", "quant", "social", "spot", "futures", "growth"}
	for _, groupName := range s.groups {
		config := TranslatorConf{
			ClientConfig: constant.ClientConfig{
				NamespaceId:         "i18n",
				TimeoutMs:           5000,
				NotLoadCacheAtStart: true,
				CacheDir:            "/tmp/nacos/cache",
				LogLevel:            "debug",
				Username:            os.Getenv("NACOS_USER"),
				Password:            os.Getenv("NACOS_PASSWORD"),
			},
			Servers: []constant.ServerConfig{
				{
					IpAddr:      os.Getenv("NACOS_IP"),
					ContextPath: "/nacos",
					Port:        8848,
					Scheme:      "http",
				},
			},
			Groups: []string{groupName},
		}
		s.configs = append(s.configs, config)
	}
}

func (s *TranslatorTestSuite) TestLocalizeText() {
	t := NacosTranslator{
		cache: make([]sync.Map, len(Locales)),
	}
	for _, config := range s.configs {
		s.mu.Lock()
		err := t.InitByTranslatorConfig(&config)
		assert.NoError(s.T(), err)
		// get current dir
		dir, err := os.Getwd()
		assert.NoError(s.T(), err)
		dir = strings.TrimSuffix(dir, "i18ngo")

		err = filepath.WalkDir(path.Join(dir, "dist_json"), func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return err
			}
			if strings.HasSuffix(path, ".json") {
				baseName := filepath.Base(path)
				group := strings.Split(baseName, "_")[0]
				if config.Groups[0] != group {
					return err
				}
				content, err := os.ReadFile(path)
				assert.NoError(s.T(), err)

				m := make(map[string]string)
				err = json.Unmarshal(content, &m)
				assert.NoError(s.T(), err)

				lt := strings.Split(strings.TrimSuffix(baseName, ".json"), "_")[1]

				s.T().Logf("group %s localize %s", group, lt)
				for k, v := range m {
					value, err := t.LocalizeText(k, LocaleT(lt))
					assert.NoError(s.T(), err)
					assert.Equal(s.T(), v, value)
					if err != nil {
						return err
					}
				}
			}
			return nil
		})
		assert.NoError(s.T(), err)
		s.mu.Unlock()
	}
}

func TestTranslatorTestSuite(t *testing.T) {
	if os.Getenv("NACOS_IP") == "" {
		t.Skip("NACOS_IP is not set")
		return
	}
	suite.Run(t, new(TranslatorTestSuite))
}

func TestSprintf(t *testing.T) {
	// Test case 1: No placeholders
	text := "Hello, World!"
	expected := text
	result := Sprintf(text, nil)
	if result != expected {
		t.Errorf("Expected '%s', but got '%s'", expected, result)
	}

	// Test case 2: Single placeholder
	text = "Hello, %s!"
	args := []string{"World"}
	expected = "Hello, World!"
	result = Sprintf(text, map[string]string{}, args...)
	if result != expected {
		t.Errorf("Expected '%s', but got '%s'", expected, result)
	}

	// Test case 3: Multiple placeholders
	text = "Hello, %s! How are you, %s?"
	args = []string{"World", "fine"}
	expected = "Hello, World! How are you, fine?"
	result = Sprintf(text, map[string]string{}, args...)
	if result != expected {
		t.Errorf("Expected '%s', but got '%s'", expected, result)
	}

	// Test case 4: Placeholder with spaces
	text = "Hello, %s World!"
	args = []string{"World"}
	expected = "Hello, World World!"
	result = Sprintf(text, map[string]string{}, args...)
	if result != expected {
		t.Errorf("Expected '%s', but got '%s'", expected, result)
	}

	// Test case 5: Placeholder with newline
	text = "Hello, %s\nWorld!"
	args = []string{"World"}
	expected = "Hello, World\nWorld!"
	result = Sprintf(text, map[string]string{}, args...)
	if result != expected {
		t.Errorf("Expected '%s', but got '%s'", expected, result)
	}

	// Test case 6: Placeholder with tab
	text = "Hello, %s\tWorld!"
	args = []string{"World"}
	expected = "Hello, World\tWorld!"
	result = Sprintf(text, map[string]string{}, args...)
	if result != expected {
		t.Errorf("Expected '%s', but got '%s'", expected, result)
	}

	// Test case 7:  {0}
	text = "Hello, {0}!"
	args = []string{"World"}
	expected = "Hello, World!"
	result = Sprintf(text, map[string]string{}, args...)
	if result != expected {
		t.Errorf("Expected '%s', but got '%s'", expected, result)
	}

	// Test case 8:  {1}
	text = "Hello, {1}!"
	args = []string{"World"}
	expected = "Hello, {1}!"
	result = Sprintf(text, map[string]string{}, args...)
	if result != expected {
		t.Errorf("Expected '%s', but got '%s'", expected, result)
	}

	// Test case 9:  {2}
	text = "Hello, {2} {2}!"
	args = []string{"World"}
	expected = "Hello, {2} {2}!"
	result = Sprintf(text, map[string]string{}, args...)
	if result != expected {
		t.Errorf("Expected '%s', but got '%s'", expected, result)
	}

	// Test case 10:  {0} {0}
	text = "Hello, {0} {0}!"
	args = []string{"World"}
	expected = "Hello, World World!"
	result = Sprintf(text, map[string]string{}, args...)
	if result != expected {
		t.Errorf("Expected '%s', but got '%s'", expected, result)
	}

	// Test case 11:  {name}
	text = "Hello, {name}!"
	expected = "Hello, World!"
	result = Sprintf(text, map[string]string{"name": "World"}, args...)
	if result != expected {
		t.Errorf("Expected '%s', but got '%s'", expected, result)
	}

	// Test case 12:  {name}
	text = "Hello, {name}!"
	expected = "Hello, {name}!"
	result = Sprintf(text, nil, args...)
	if result != expected {
		t.Errorf("Expected '%s', but got '%s'", expected, result)
	}
}

func TestNacosTranslator_LocalizeInfof(t *testing.T) {
	locale := Locales[0]
	t.Run("test valid input", func(t *testing.T) {
		translator := &NacosTranslator{
			cache: make([]sync.Map, len(Locales)),
		}
		text := "hello {name}"
		err := translator.UpdateLocale(LocaleT(locale), map[string]string{
			text: text,
		})
		assert.NoError(t, err)
		name := "John"
		expected := "hello John"

		result, err := translator.LocalizeInfof(text, locale, "name", name)
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("test 2 valid input", func(t *testing.T) {
		translator := &NacosTranslator{
			cache: make([]sync.Map, len(Locales)),
		}
		text := "hello {name} and {name}"
		err := translator.UpdateLocale(LocaleT(locale), map[string]string{
			text: text,
		})
		assert.NoError(t, err)
		name := "John"
		expected := "hello John and John"

		result, err := translator.LocalizeInfof(text, locale, "name", name)
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("test 2 valid input with ._-", func(t *testing.T) {
		translator := &NacosTranslator{
			cache: make([]sync.Map, len(Locales)),
		}
		text := "hello {test.n-_ame} and {test.n-_ame}"
		err := translator.UpdateLocale(LocaleT(locale), map[string]string{
			text: text,
		})
		assert.NoError(t, err)
		name := "John"
		expected := "hello John and John"

		result, err := translator.LocalizeInfof(text, locale, "test.n-_ame", name)
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("test invalid locale", func(t *testing.T) {
		translator := &NacosTranslator{
			cache: make([]sync.Map, len(Locales)),
		}
		text := "hello {name}"
		err := translator.UpdateLocale(locale, map[string]string{
			text: text,
		})
		assert.NoError(t, err)

		_, err = translator.LocalizeInfof(text, "invalid-locale", "name", "John")
		assert.Error(t, err)
		assert.Equal(t, "cache not found for locale invalid-locale", err.Error())
	})

	t.Run("test invalid input (odd number of arguments)", func(t *testing.T) {
		translator := &NacosTranslator{
			cache: make([]sync.Map, len(Locales)),
		}
		text := "hello {name}"
		err := translator.UpdateLocale(locale, map[string]string{
			text: text,
		})
		assert.NoError(t, err)
		_, err = translator.LocalizeInfof(text, locale, "name")
		assert.Error(t, err)
		assert.Equal(t, "infof requires an even number of arguments", err.Error())
	})

	t.Run("test empty input", func(t *testing.T) {
		translator := &NacosTranslator{
			cache: make([]sync.Map, len(Locales)),
		}
		text := ""
		err := translator.UpdateLocale(locale, map[string]string{
			text: text,
		})
		assert.NoError(t, err)

		_, err = translator.LocalizeInfof(text, locale)
		assert.NoError(t, err)
		assert.Empty(t, "")
	})

	t.Run("test invalid locale", func(t *testing.T) {
		translator := &NacosTranslator{
			cache: make([]sync.Map, len(Locales)),
		}
		text := "hello {name}"
		err := translator.UpdateLocale(LocaleT("en"), map[string]string{
			text: text,
		})
		assert.NoError(t, err)
		name := "John"

		_, err = translator.LocalizeInfof(text, LocaleT("cc"), "name", name)
		assert.Error(t, err)
		assert.Equal(t, "cache not found for locale cc", err.Error())
	})
}

func Test_Es(t *testing.T) {
	s := `\{[a-zA-Z0-9_\-.\s]+?\}|\%s`
	jsonEscapedString := strconv.Quote(s)
	t.Log(jsonEscapedString)
	assert.Equal(t, `"\\{[a-zA-Z0-9_\\-.\\s]+?\\}|\\%s"`, jsonEscapedString)
}

func TestLocalizeSprintf(t *testing.T) {
	translator := &NacosTranslator{
		cache: make([]sync.Map, len(Locales)),
	}
	err := translator.UpdateLocale(LocaleT("en"), map[string]string{
		"Hello %s":                "Hello %s",
		"Hello":                   "Hello",
		"Hello %s and %s":         "Hello %s and %s",
		"Hello {name} and {name}": "Hello {name} and {name}",
	})
	assert.NoError(t, err)
	tests := []struct {
		name    string
		text    string
		locale  LocaleT
		args    []string
		want    string
		wantErr bool
	}{
		{
			name:    "Successful localization and formatting",
			text:    "Hello %s",
			locale:  "en",
			args:    []string{"World"},
			want:    "Hello World",
			wantErr: false,
		},
		{
			name:    "Error in localization",
			text:    "Non-existent text",
			locale:  "en",
			args:    []string{},
			want:    "",
			wantErr: true,
		},
		{
			name:    "No arguments provided",
			text:    "Hello",
			locale:  "en",
			args:    []string{},
			want:    "Hello",
			wantErr: false,
		},
		{
			name:    "Multiple arguments provided",
			text:    "Hello %s and %s",
			locale:  "en",
			args:    []string{"World", "Universe"},
			want:    "Hello World and Universe",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := translator.LocalizeSprintf(tt.text, tt.locale, tt.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("%s LocalizeSprintf() error = %v, wantErr %v", t.Name(), err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("LocalizeSprintf() got = %v, want %v", got, tt.want)
			}
		})
	}
}
