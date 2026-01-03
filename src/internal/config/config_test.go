package config

import (
	"os"
	"testing"
)

func TestValidate(t *testing.T) {
	// Setup
	tmp, err := os.CreateTemp("", "mpv")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())
	mpvPath := tmp.Name()

	validSettings := map[string]any{
		"roots": []any{"/tmp"},
	}

	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: Config{
				ActiveProfile: "local",
				Player: PlayerConfig{
					MPVPath:       mpvPath,
					InitialVolume: 50,
				},
				Profiles: []Profile{
					{
						ID:       "local",
						Provider: "filesystem",
						Enabled:  true,
						Settings: validSettings,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "missing active profile",
			cfg: Config{
				ActiveProfile: "missing",
				Player:        PlayerConfig{MPVPath: mpvPath},
				Profiles:      []Profile{},
			},
			wantErr: true,
		},
		{
			name: "disabled profile",
			cfg: Config{
				ActiveProfile: "local",
				Player:        PlayerConfig{MPVPath: mpvPath},
				Profiles: []Profile{
					{ID: "local", Enabled: false},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid mpv path",
			cfg: Config{
				ActiveProfile: "local",
				Player:        PlayerConfig{MPVPath: "/invalid/mpv/path"},
				Profiles: []Profile{
					{ID: "local", Enabled: true, Provider: "filesystem", Settings: validSettings},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
