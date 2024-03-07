package log

import (
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
)

func SetupLogs(cfg LogsConfig) {
	level := cfg.Level
	if level == "" {
		level = "info"
	}
	l, err := LevelFromString(level)
	if err != nil {
		slog.Error("failed to parse config file log level", "error", err)
		os.Exit(1)
	}

	// set global logger with custom options
	slog.SetDefault(slog.New(getLogHandler(cfg, l)))
}

func getLogHandler(cfg LogsConfig, l *slog.Level) slog.Handler {
	if cfg.JSONOutput {
		return slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level:       l,
			ReplaceAttr: ReplaceLevels,
		})
	}
	return tint.NewHandler(os.Stdout, &tint.Options{
		Level:       l,
		TimeFormat:  "2006/01/02 15:04:05",
		NoColor:     !isatty.IsTerminal(os.Stdout.Fd()),
		ReplaceAttr: ReplaceLevels,
	})
}
