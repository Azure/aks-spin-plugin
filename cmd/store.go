package cmd

import (
	"fmt"

	"github.com/azure/spin-aks-plugin/pkg/azure"
	"github.com/azure/spin-aks-plugin/pkg/config"
	"github.com/azure/spin-aks-plugin/pkg/logger"
	"github.com/azure/spin-aks-plugin/pkg/usererror"
	"github.com/spf13/cobra"
)

var (
	secretName  = ""
	secretValue = ""
)

func init() {
	rootCmd.AddCommand(storeCmd)

}

const ctxKeyAKV = "akv"

var storeCmd = &cobra.Command{
	Use:        "store",
	Short:      "Manages variables in the Spin AKS config file.",
	Long:       "Manages variables in the Spin AKS config file.",
	Args:       cobra.ExactArgs(2),
	ArgAliases: []string{"variableName", "variableValue"},
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		lgr := logger.FromContext(ctx)

		secretName = args[0]
		if secretName == "" {
			return fmt.Errorf("secret name cannot be empty")
		}
		secretValue = args[1]
		if secretValue == "" {
			return fmt.Errorf("secret value cannot be empty")
		}

		kv := config.GetKeyVault()

		if kv.Name == "" || kv.Subscription == "" || kv.ResourceGroup == "" {
			return usererror.New(fmt.Errorf("no keyvault found in config"), "no keyvault found in config, try running `spin aks init` first")
		}

		lgr.Info("starting to store variable in keyvault")
		akv, err := azure.GetKeyVault(ctx, kv.Subscription, kv.ResourceGroup, kv.Name)
		if err != nil {
			return fmt.Errorf("getting keyvault: %w", err)
		}
		if akv == nil {
			return fmt.Errorf("getting keyvault: keyvault was nil")
		}

		lgr.Info("starting to put variable")
		defer lgr.Info("finished putting variable")

		err = akv.PutSecretIfNewValue(ctx, secretName, secretValue)
		if err != nil {
			return fmt.Errorf("putting secret: %w", err)
		}

		return nil
	},
}
