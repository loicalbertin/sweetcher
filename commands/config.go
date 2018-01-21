package commands

// Config is the root of a configuration file
type Config struct {
	Server   Server             `json:"server,omitempty" mapstructure:"server"`
	Proxies  map[string]string  `json:"proxies,omitempty" mapstructure:"proxies"`
	Profiles map[string]Profile `json:"profiles,omitempty" mapstructure:"profiles"`
}

// Server represents a Sweetcher server configuration file
type Server struct {
	Address string `json:"address,omitempty" mapstructure:"address"`
	Profile string `json:"profile,omitempty" mapstructure:"profile"`
}

// Profile represents a Profile definition
type Profile struct {
	Default string `json:"default,omitempty" mapstructure:"default"`
	Rules   []Rule `json:"rules,omitempty" mapstructure:"rules"`
}

// Rule is a routing rule to a proxy
type Rule struct {
	HostWildcard string `json:"host_wildcard,omitempty" mapstructure:"host_wildcard"`
	Proxy        string `json:"proxy,omitempty" mapstructure:"proxy"`
}
