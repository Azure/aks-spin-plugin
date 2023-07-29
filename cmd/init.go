package cmd

import (
	"fmt"

	"github.com/olivermking/spin-aks-plugin/pkg/config"
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
		if err := config.Load(config.Opts{}); err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		// TODO: prompt user for value

		if err := config.Write(); err != nil {
			return fmt.Errorf("writting config: %w", err)
		}

		return nil
	},
}
