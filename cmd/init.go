package cmd

import (
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/subscription/armsubscription"
	"github.com/olivermking/spin-aks-plugin/pkg/azure"
	"github.com/olivermking/spin-aks-plugin/pkg/config"
	"github.com/olivermking/spin-aks-plugin/pkg/logger"
	"github.com/olivermking/spin-aks-plugin/pkg/prompt"
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

		subs, err := azure.ListSubscriptions(ctx)
		if err != nil {
			return fmt.Errorf("listing subscriptions: %w", err)
		}

		lgr.Debug("prompting for subscription")
		sub, err := prompt.Select("Select a Subscription", subs, &prompt.SelectOpt[armsubscription.Subscription]{
			Field: func(t armsubscription.Subscription) string {
				return *t.DisplayName
			},
		})
		if err != nil {
			return fmt.Errorf("selecting subscription: %w", err)
		}
		lgr.Debug("finished prompting for subscription")

		fmt.Println(*sub.DisplayName)

		if err := config.Write(); err != nil {
			return fmt.Errorf("writting config: %w", err)
		}

		lgr.Debug("finished init command")
		return nil
	},
}
