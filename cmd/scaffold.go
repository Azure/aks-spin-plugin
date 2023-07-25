package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(scaffold)
}

var scaffold = &cobra.Command{
	Use:   "scaffold",
	Short: "Generates required Dockerfile and Kubernetes manifests",
	Long:  "Creates Dockerfile and Kubernetes manifests required to run your application on AKS.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Scaffold")
	},
}
