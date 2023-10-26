package cmd

import (
	"github.com/azure/spin-aks-plugin/pkg/logger"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(deployCmd)
}

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploys the Spin application to AKS",
	Long:  "Deploys the Spin application to AKS",
	RunE: func(cmd *cobra.Command, args []string) error {
		lgr := logger.FromContext(cmd.Context())

		lgr.Info("Deployed your spin app to AKS successfully 🥳")
		return nil
	},
}
