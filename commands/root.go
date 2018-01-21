package commands

import "github.com/spf13/cobra"

// RootCmd is the root cli command
var RootCmd = &cobra.Command{
	Use:   "sweetcher",
	Short: "Sweetcher is a system proxy switcher based on rules",
	Long: `For those who know the hell of enterprise proxies!
Sweetcher is inspired for web browsers proxy switcher plugins but allow to do it as 
your system proxy. Just configure it as your default system proxy and configure it 
using rules defined in profiles`,
}
