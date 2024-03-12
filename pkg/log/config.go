package log

type LogsConfig struct {
	Level      string `json:"level,omitempty" mapstructure:"level"`
	JSONOutput bool   `json:"json_output,omitempty" mapstructure:"json_output"`
}
