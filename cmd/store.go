package cmd

import (
	"fmt"

	"github.com/azure/spin-aks-plugin/pkg/azure"
	"github.com/azure/spin-aks-plugin/pkg/config"
	"github.com/azure/spin-aks-plugin/pkg/logger"
	"github.com/spf13/cobra"
)
var (
	secretName = ""
	secretValue = ""
)


func init() {
	rootCmd.AddCommand(storeCmd)

}

const ctxKeyAKV = "akv"

var storeCmd = &cobra.Command{
	Use:   "store",
	Short: "Manages variables in the Spin AKS config file.",
	Long:  "Manages variables in the Spin AKS config file.",
	Args: cobra.ExactArgs(2),
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

		lgr.Info("starting to store variable in keyvault")
		akv, err := azure.GetKeyVault(ctx,kv.Subscription,kv.ResourceGroup,kv.Name)
		if err != nil {
			return fmt.Errorf("getting keyvault: %w", err)
		}
		if akv == nil {
			return fmt.Errorf("getting keyvault: keyvault was nil")
		}

		//TODO: validate rg exists here

		//TODO: validate keyvault access here

		lgr.Info("starting to put variable")
		defer lgr.Info("finished putting variable")

		err = akv.PutSecret(ctx, secretName, secretValue)
		if err != nil {
			return fmt.Errorf("putting secret: %w", err)
		}

		//TODO: list keyvaults if one isn't specified
		// to do this we wouldn't require the kv flag and instead use Oliver's fancy selectors and the ListKeyVaults

		return nil
	},
}
