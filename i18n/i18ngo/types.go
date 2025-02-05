package i18ngo

import "github.com/nacos-group/nacos-sdk-go/v2/common/constant"

type NacosConf struct {
	Client  constant.ClientConfig
	Servers []constant.ServerConfig
}

type TranslatorConf struct {
	ClientConfig  constant.ClientConfig
	Servers       []constant.ServerConfig
	Groups        []string
	DataIDs       []LocaleT
	DefaultLocale LocaleT
}

type LocaleT string

const (
	AR = "ar"
	BR = "br"
	DE = "de"
	ES = "es"
	FR = "fr"
	ID = "id"
	IT = "it"
	JA = "ja"
	KR = "kr"
	NL = "nl"
	PT = "pt"
	RU = "ru"
	TH = "th"
	TR = "tr"
	UK = "uk"
	VI = "vi"
	ZH = "zh"
	TW = "tw"
	EN = "en"
)

var Locales = []LocaleT{
	AR, BR, DE, ES, FR, ID, IT, JA, KR, NL, PT, RU, TH, TR, UK, VI, ZH, TW, EN,
}
