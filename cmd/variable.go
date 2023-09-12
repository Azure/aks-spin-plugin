package cmd

import (
	"fmt"

	"github.com/azure/spin-aks-plugin/pkg/azure"
	"github.com/azure/spin-aks-plugin/pkg/logger"
	"github.com/spf13/cobra"
)

var kvName string
var secretName string

func init() {
	rootCmd.AddCommand(variableCmd)
	rootCmd.Flags().StringVar(&kvName, "kv-name", "", "Name of the keyvault to use")
	rootCmd.Flags().StringVar(&secretName, "secret-name", "", "Name of the secret to use")
}

var variableCmd = &cobra.Command{
	Use:   "variable",
	Short: "Manages variables in the Spin AKS config file.",
	Long:  "Manages variables in the Spin AKS config file.",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		lgr := logger.FromContext(ctx)


		subscriptionId := ""
		resourceGroup := ""

		lgr.Info("starting to list keyvaults")
		kvs, err := azure.ListKeyVaults(ctx, subscriptionId, resourceGroup)
		if err != nil {
			return fmt.Errorf("listing keyvaults: %w", err)
		}
		secretName = "test-secret"
		for _, kv := range kvs {
			lgr.Info("keyvault", "name",kv.GetId())
			err = kv.PutSecret(ctx, secretName, "test")
			if err != nil {
				return fmt.Errorf("putting secret: %w", err)
			}
		}

		return nil
	},
}
