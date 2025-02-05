package i18n

import (
	"errors"
	"time"
)

var (
	ErrEmptyIPAddr = errors.New("empty ip addr")
	ErrEmptyTeamID = errors.New("empty team id")
)

type TeamID string

var (
	PHP      = TeamID("php")
	INFRA    = TeamID("infra")
	PAY      = TeamID("pay")
	ASSET    = TeamID("asset")
	COPY     = TeamID("copy")
	EARN     = TeamID("earn")
	MARKT    = TeamID("markt")
	PLATFORM = TeamID("platform")
	QUANT    = TeamID("quant")
	SOCIAL   = TeamID("social")
	SPOT     = TeamID("spot")
	FUTURES  = TeamID("futures")
	UNIFIED  = TeamID("unified")
	GROWTH   = TeamID("growth")
)

// nolint revive //go-zero default config styele
type I18nConf struct {
	IPAddr      string
	Port        uint64        `json:",default=8848"`
	NamespaceID string        `json:",default=i18n"`
	Timeout     time.Duration `json:",default=10000ms"`
	Username    string        `json:",default=nacos"`
	Password    string        `json:",optional"`
	// 业务id
	TeamID TeamID `json:",default=php"`
	// 业务id组
	TeamIDs       []TeamID `json:",optional"`
	DefaultLocale string   `json:",default=en"`
	LogLevel      string   `json:",default=info"`
}

func (ic I18nConf) Validate() error {
	if ic.IPAddr == "" {
		return ErrEmptyIPAddr
	}
	if len(ic.TeamIDs) == 0 && ic.TeamID == "" {
		return ErrEmptyTeamID
	}
	return nil
}

func (i TeamID) String() string {
	return string(i)
}
