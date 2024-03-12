package log

import (
	"log/slog"
	"strings"
)

const LevelTrace = slog.Level(-8)

var LevelNames = map[slog.Leveler]string{
	LevelTrace: "TRACE",
}

func LevelFromString(str string) (*slog.Level, error) {
	for k, v := range LevelNames {
		if strings.TrimSpace(strings.ToUpper(str)) == v {
			l := k.Level()
			return &l, nil
		}
	}

	l := new(slog.Level)
	err := l.UnmarshalText([]byte(str))
	return l, err
}

func ReplaceLevels(groups []string, a slog.Attr) slog.Attr {
	if a.Key == slog.LevelKey {
		level := a.Value.Any().(slog.Level)
		levelLabel, exists := LevelNames[level]
		if !exists {
			levelLabel = level.String()
		}

		a.Value = slog.StringValue(levelLabel)
	}

	return a
}
