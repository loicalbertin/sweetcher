package commands

import (
	"log"
	"net/url"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/loicalbertin/sweetcher/proxy"
)

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
			server := &proxy.Server{Addr: conf.Server.Address}
			server.SetupProfile(profile)

			viper.WatchConfig()
			viper.OnConfigChange(func(e fsnotify.Event) {
				log.Printf("reloading config file %q, (%q)", e.Name, e.Op)
				c := &Config{}
				err := viper.Unmarshal(c)
				if err != nil {
					log.Printf("Failed to read config file %q: %v", e.Name, err)
				}
				profile, err := generateProfile(c)
				if err != nil {
					log.Printf("Failed to create profile from config file %q: %v", e.Name, err)
				}
				server.SetupProfile(profile)
			})

			return server.ListenAndServe()

		},
	}
	RootCmd.AddCommand(serveCmd)
}

func initConfig() (*Config, error) {
	viper.SetConfigName("sweetcher")        // name of config file (without extension)
	viper.AddConfigPath("/etc/sweetcher/")  // path to look for the config file in
	viper.AddConfigPath("$HOME/.sweetcher") // call multiple times to add many search paths
	viper.AddConfigPath(".")                // optionally look for config in the working directory
	err := viper.ReadInConfig()             // Find and read the config file
	if err != nil {                         // Handle errors reading the config file
		return nil, errors.Errorf("Fatal error config file: %s", err)
	}
	conf := &Config{}
	viper.Unmarshal(conf)
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
