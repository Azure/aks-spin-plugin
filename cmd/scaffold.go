package cmd

import (
	"fmt"

	"github.com/olivermking/spin-aks-plugin/pkg/spin"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(scaffold)
}

var scaffold = &cobra.Command{
	Use:   "scaffold",
	Short: "Generates required Dockerfile and Kubernetes manifests",
	Long:  "Creates Dockerfile and Kubernetes manifests required to run your application on AKS.",
	RunE: func(cmd *cobra.Command, args []string) error {
		m, err := spin.Load()
		if err != nil {
			return fmt.Errorf("parsing spin: %w", err)
		}

		fmt.Println(m)

		return nil
	},
}
