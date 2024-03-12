package cmd

import (
	"context"
	"log/slog"
	"net/url"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/loicalbertin/sweetcher/pkg/log"
	"github.com/loicalbertin/sweetcher/pkg/proxy"
)

var server *proxy.Server

func init() {
	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "runs a Sweetcher server",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			conf, err := initConfig()
			if err != nil {
				return err
			}
			profile, err := generateProfile(conf)
			if err != nil {
				return err
			}
			server = &proxy.Server{Addr: conf.Server.Address}
			server.SetupProfile(profile)

			viper.WatchConfig()
			viper.OnConfigChange(updateConfigOnChangeEvent)
			slog.Log(context.Background(), log.LevelTrace, "Running sweetcher server", "config", conf)
			// slog.Debug("Running sweetcher server", "config", conf)

			return server.ListenAndServe()

		},
	}
	RootCmd.AddCommand(serveCmd)
}

func updateConfigOnChangeEvent(e fsnotify.Event) {

	logger := slog.With(slog.String("file", e.Name))
	logger.Info("reloading config file")
	c := &Config{}
	err := viper.Unmarshal(c)
	if err != nil {
		logger.Error("Failed to read config file", "error", err)
		return
	}
	logger = logger.With(slog.String("profile", c.Server.Profile))
	err = log.SetupLogs(c.Server.Logs)
	if err != nil {
		os.Exit(1)
	}

	profile, err := generateProfile(c)
	if err != nil {
		logger.Error("Failed to create profile from config file", "error", err)
		return
	}
	server.SetupProfile(profile)
	logger.Info("Profile reloaded")
}

func initConfig() (*Config, error) {
	viper.SetConfigName("sweetcher")        // name of config file (without extension)
	viper.AddConfigPath(".")                // path to look for the config file in
	viper.AddConfigPath("$HOME/.sweetcher") // call multiple times to add many search paths
	viper.AddConfigPath("/etc/sweetcher/")  // optionally look for config in the working directory
	err := viper.ReadInConfig()             // Find and read the config file
	if err != nil {                         // Handle errors reading the config file
		return nil, errors.Errorf("Fatal error config file: %s", err)
	}
	conf := &Config{}
	viper.Unmarshal(conf)
	err = log.SetupLogs(conf.Server.Logs)
	if err != nil {
		os.Exit(1)
	}
	return conf, nil
}

func generateProfile(cfg *Config) (*proxy.Profile, error) {
	proxies := make(map[string]*url.URL)
	for proxyName, proxyURL := range cfg.Proxies {
		p, err := url.Parse(proxyURL)
		if err != nil {
			return nil, errors.Wrapf(err, "Malformed proxy definition for proxy %q", proxyName)
		}
		proxies[proxyName] = p
	}
	// Defaults to direct proxy
	profile := &proxy.Profile{}
	if cfg.Server.Profile == "direct" {
		return profile, nil
	}
	p, ok := cfg.Profiles[cfg.Server.Profile]
	if !ok {
		return nil, errors.Errorf("specified server profile %q not found", cfg.Server.Profile)
	}
	def, ok := proxies[p.Default]
	if !ok && p.Default != "direct" {
		return nil, errors.Errorf("specified default proxy %q not found for profile %q", p.Default, cfg.Server.Profile)
	}
	profile.Default = def
	for _, r := range p.Rules {
		rp, ok := proxies[r.Proxy]
		if !ok && r.Proxy != "direct" {
			return nil, errors.Errorf("specified proxy %q not found for rule %q in profile %q", r.Proxy, r.HostWildcard, cfg.Server.Profile)
		}
		profile.Rules = append(profile.Rules, proxy.Rule{Pattern: r.HostWildcard, Proxy: rp})
	}
	return profile, nil
}
