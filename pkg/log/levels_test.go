package log

import (
	"log/slog"
	"reflect"
	"testing"

	"gotest.tools/v3/assert"
)

func TestLevelFromString(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name    string
		args    args
		want    slog.Level
		wantErr bool
	}{
		{"TestLevelInfoLogs", args{"info"}, slog.LevelInfo, false},
		{"TestLevelInfoUpLogs", args{"InFo"}, slog.LevelInfo, false},
		{"TestLevelDebugLogs", args{"DEBUG"}, slog.LevelDebug, false},
		{"TestLevelWarnLogs", args{"WARN"}, slog.LevelWarn, false},
		{"TestLevelErrorLogs", args{"eRRoR"}, slog.LevelError, false},
		{"TestLevelTraceLogs", args{"trace"}, LevelTrace, false},
		{"TestLevelUnknownLogs", args{"Unknown"}, slog.LevelInfo, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LevelFromString(tt.args.str)
			if (err != nil) != tt.wantErr {
				t.Errorf("LevelFromString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, got.Level(), tt.want, "LevelFromString() = %v, want %v", got.Level(), tt.want)
		})
	}
}

func TestReplaceLevels(t *testing.T) {
	type args struct {
		groups []string
		a      slog.Attr
	}
	tests := []struct {
		name string
		args args
		want slog.Attr
	}{
		{"ReplaceTrace", args{groups: []string{}, a: slog.Attr{Key: slog.LevelKey, Value: slog.AnyValue(LevelTrace)}}, slog.Attr{Key: slog.LevelKey, Value: slog.StringValue("TRACE")}},
		{"ReplaceDebug", args{groups: []string{}, a: slog.Attr{Key: slog.LevelKey, Value: slog.AnyValue(slog.LevelDebug)}}, slog.Attr{Key: slog.LevelKey, Value: slog.StringValue("DEBUG")}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ReplaceLevels(tt.args.groups, tt.args.a); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReplaceLevels() = %v, want %v", got, tt.want)
			}
		})
	}
}
