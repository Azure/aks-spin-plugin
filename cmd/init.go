package cmd

import (
	"fmt"

	"github.com/olivermking/spin-aks-plugin/pkg/config"
	"github.com/olivermking/spin-aks-plugin/pkg/logger"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Creates the Spin AKS config describing how to deploy your application",
	Long:  "Generates the Spin AKS config based on guided user input. The AKS Spin config file is used to store the deployment targets of your application.",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		lgr := logger.FromContext(ctx)
		lgr.Debug("starting init command")

		if err := config.EnsureCluster(ctx); err != nil {
			return fmt.Errorf("ensuring config cluster: %w", err)
		}

		if err := config.Write(); err != nil {
			return fmt.Errorf("writting config: %w", err)
		}

		lgr.Debug("finished init command")
		return nil
	},
}
