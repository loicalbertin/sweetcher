package commands

import (
	"net/url"

	"github.com/loicalbertin/sweetcher/proxy"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	for profileName, p := range cfg.Profiles {
		if profileName == cfg.Server.Profile {
			profile.Default = proxies[p.Default]
			for _, r := range p.Rules {
				profile.Rules = append(profile.Rules, proxy.Rule{Pattern: r.HostWildcard, Proxy: proxies[r.Proxy]})
			}
			break
		}
	}
	return profile, nil
}
