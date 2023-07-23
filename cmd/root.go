package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "spin aks",
	Short: "A Spin Plugin for Azure Kubernetes Service",
	Long: `Spin AKS is a Spin Plugin that guides users through deploying Spin applications to Azure Kubernetes Service.
	
To walk through all deploy steps, run the 'spin aks up' command.

	$ spin aks up

For more information, please visit the GitHub page at https://github.com/OliverMKing/spin-aks-plugin.

Report any feature requests or issues at https://github.com/OliverMKing/spin-aks-plugin/issues.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
