package i18n

import (
	"testing"
)

func TestI18nConfValidate(t *testing.T) {
	tests := []struct {
		name    string
		conf    I18nConf
		wantErr error
	}{
		{
			name: "empty IpAddr",
			conf: I18nConf{
				IPAddr: "",
			},
			wantErr: ErrEmptyIPAddr,
		},
		{
			name: "empty TeamIDs",
			conf: I18nConf{
				IPAddr: "example",
			},
			wantErr: ErrEmptyTeamID,
		},
		{
			name: "valid IpAddr and TeamIDs",
			conf: I18nConf{
				IPAddr: "example",
				TeamIDs: []TeamID{
					"php",
				},
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.conf.Validate(); err != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
