package cmd

import (
	"os/exec"

	"github.com/azure/spin-aks-plugin/pkg/config"
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

		lgr.Info("Checking if Helm is installed...")
		checkHelmResult := exec.Command("helm", "version")
		_, err := checkHelmResult.Output()
		if err != nil {
			lgr.Info("Helm is not installed, please install it and try again")
			return nil
		}
		lgr.Info("Helm is installed")

		c := config.Get()

		lgr.Info("Deploying to AKS")
		commandResult := exec.Command("kubectl", "apply", "-f", c.K8sResources+"*.yaml")
		out, err := commandResult.Output()
		if err != nil {
			if len(out) != 0 {
				lgr.Error(string(out))
			}
		}
		lgr.Info("Deployed your spin app to AKS successfully 🥳")
		return nil
	},
}
