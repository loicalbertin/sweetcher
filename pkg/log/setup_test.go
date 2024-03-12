package log

import (
	"context"
	"log/slog"
	"testing"
)

func TestSetupLogs(t *testing.T) {
	type args struct {
		cfg LogsConfig
	}
	tests := []struct {
		name           string
		args           args
		expectedLevels []slog.Level
		wantErr        bool
	}{
		{"DefaultConfig", args{LogsConfig{}}, []slog.Level{slog.LevelInfo, slog.LevelError, slog.LevelWarn}, false},
		{"DefaultConfigInJSON", args{LogsConfig{JSONOutput: true}}, []slog.Level{slog.LevelInfo, slog.LevelError, slog.LevelWarn}, false},
		{"ConfigDebug", args{LogsConfig{Level: slog.LevelDebug.String()}}, []slog.Level{slog.LevelDebug, slog.LevelInfo, slog.LevelError, slog.LevelWarn}, false},
		{"WrongLevel", args{LogsConfig{Level: "wrong"}}, []slog.Level{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SetupLogs(tt.args.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetupLogs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for _, level := range tt.expectedLevels {
				if !slog.Default().Handler().Enabled(context.Background(), level) {
					t.Errorf("level %q is expected to be enabled", level.String())
				}
			}

		})
	}
}
