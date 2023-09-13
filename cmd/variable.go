package cmd

import (
	"context"
	"fmt"

	"github.com/azure/spin-aks-plugin/pkg/azure"
	"github.com/azure/spin-aks-plugin/pkg/logger"
	"github.com/spf13/cobra"
)
var (
	subId = ""
	rgName = ""
	kvName = ""

	secretName = ""
	secretValue = ""
)


func init() {
	rootCmd.AddCommand(variableCmd)
	rootCmd.PersistentFlags().StringVarP(&subId, "subscription-id", "s", "", "subscriptionid to use")
	rootCmd.PersistentFlags().StringVarP(&rgName, "resource-group", "r", "", "resource group to use")
	rootCmd.PersistentFlags().StringVarP(&kvName, "keyvault-name", "k", "", "name of the keyvault to use")
	rootCmd.MarkPersistentFlagRequired("subscription-id")
	rootCmd.MarkPersistentFlagRequired("resource-group")
	rootCmd.MarkPersistentFlagRequired("keyvault-name")

	variableCmd.AddCommand(variablePutCmd)

	variablePutCmd.Flags().StringVar(&secretName, "secret-name", "", "Name of the secret to put")
	variablePutCmd.Flags().StringVar(&secretValue, "secret-value", "", "Value of the secret to put")
	variablePutCmd.MarkFlagRequired("n")
	variablePutCmd.MarkFlagRequired("v")

}

const ctxKeyAKV = "akv"

var variableCmd = &cobra.Command{
	Use:   "variable",
	Short: "Manages variables in the Spin AKS config file.",
	Long:  "Manages variables in the Spin AKS config file.",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		akv, err := azure.GetKeyVault(ctx, subId, rgName, kvName)
		if err != nil {
			return fmt.Errorf("getting keyvault: %w", err)
		}
		if akv == nil {
			return fmt.Errorf("getting keyvault: keyvault was nil")
		}
		ctx = context.WithValue(ctx,ctxKeyAKV, akv)
		cmd.SetContext(ctx)

		//TODO: validate rg exists here

		//TODO: validate keyvault access here

		//TODO: list keyvaults if one isn't specified
		// to do this we wouldn't require the kv flag and instead use Oliver's fancy selectors and the ListKeyVaults

		return nil
	},
}

var variablePutCmd = &cobra.Command{
	Use:   "put",
	Short: "Puts a variable in the Spin AKS config file.",
	Long:  "Puts a variable in the Spin AKS config file.",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		lgr := logger.FromContext(ctx)

		ctxAkv := ctx.Value(ctxKeyAKV)
		akv, ok := ctxAkv.(*azure.Akv)
		if !ok {
			return fmt.Errorf("casting keyvault from context value")
		}
		if akv == nil {
			return fmt.Errorf("getting keyvault from context")
		}
		lgr.Info("starting to put variable")
		defer lgr.Info("finished putting variable")

		err := akv.PutSecret(ctx, secretName, secretValue)
		if err != nil {
			return fmt.Errorf("putting secret: %w", err)
		}

		return nil
	},
}
