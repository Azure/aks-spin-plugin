package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/olivermking/spin-aks-plugin/pkg/logger"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "aks", // this is really spin aks but must be defined in the templates
	Short: "A Spin Plugin for Azure Kubernetes Service",
	Long: `Spin AKS is a Spin Plugin that guides users through deploying Spin applications to Azure Kubernetes Service.
	
To walk through all deploy steps, run the 'spin aks up' command.

	$ spin aks up

For more information, please visit the GitHub page at https://github.com/OliverMKing/spin-aks-plugin.

Report any feature requests or issues at https://github.com/OliverMKing/spin-aks-plugin/issues.`,
}

// Config describes dynamic variables for the cmd which should be set at build time
type Config struct {
	// Version is the version of the tool
	Version string
}

func Execute(c Config) {
	rootCmd.Version = c.Version // if version is empty the only consequence should be the version command not working

	// need to prefix command use with "spin" because it's a spin plugin
	rootCmd.SetUsageTemplate(
		strings.NewReplacer(
			"{{.UseLine}}", "spin {{.UseLine}}",
			"{{.CommandPath}}", "spin {{.CommandPath}}",
		).Replace(rootCmd.UsageTemplate()))
	rootCmd.SetVersionTemplate(`{{printf "spin-aks-plugin %s" .Version}}
`) // new line is deliberate because it renders better

	// set verbose
	var verbose bool
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "print additional information typically useful for debugging")
	rootCmd.PersistentPreRun = func(_ *cobra.Command, _ []string) {
		logger.SetVerbose(verbose)
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
