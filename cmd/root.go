package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/olivermking/spin-aks-plugin/pkg/config"
	"github.com/olivermking/spin-aks-plugin/pkg/logger"
	"github.com/olivermking/spin-aks-plugin/pkg/usererror"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	SilenceErrors: true,  // we handle printing error information ourselves in Execute fn
	Use:           "aks", // this is really spin aks but must be defined in the templates
	Short:         "A Spin Plugin for Azure Kubernetes Service",
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
	rootCmd.Version = c.Version // if version is empty the only consequence should be the version command not working. We shouldn't panic or fail because of that.

	// need to prefix command use with "spin" because it's a spin plugin
	rootCmd.SetUsageTemplate(
		strings.NewReplacer(
			"{{.UseLine}}", "spin {{.UseLine}}",
			"{{.CommandPath}}", "spin {{.CommandPath}}",
		).Replace(rootCmd.UsageTemplate()))
	rootCmd.SetVersionTemplate(`{{printf "spin-aks-plugin %s" .Version}}
`) // new line is deliberate because it renders better

	// set global flags
	var verbose bool
	var spinAksConfig string
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "print additional information typically useful for debugging")
	rootCmd.PersistentFlags().StringVarP(&spinAksConfig, "config", "c", "", "path to the spin aks config toml file")
	rootCmd.PersistentPreRunE = func(_ *cobra.Command, _ []string) error {
		logger.SetVerbose(verbose)
		if err := config.Load(config.Opts{
			Path: spinAksConfig,
		}); err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		return nil
	}

	ctx := context.Background()
	lgr := logger.FromContext(ctx)
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		errMsg := fmt.Sprintf("%s", err)

		if ue, ok := usererror.Is(err); ok {
			if verbose {
				lgr.Error(errMsg)
			}

			errMsg = ue.Msg()
		}

		lgr.Error(errMsg)
		os.Exit(1)
	}
}
