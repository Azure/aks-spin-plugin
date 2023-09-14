package cmd

import (
	"fmt"

	"github.com/azure/spin-aks-plugin/pkg/spin"
	"github.com/azure/spin-aks-plugin/pkg/templates"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(newScaffoldCmd())
}
func newScaffoldCmd() *cobra.Command {
	sc := &templates.ScaffoldCmd{}

	var scaffold = &cobra.Command{
		Use:   "scaffold",
		Short: "Generates required Dockerfile and Kubernetes manifests",
		Long:  "Creates Dockerfile and Kubernetes manifests required to run your application on AKS.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := setConfig(sc); err != nil {
				return err
			}

			m, err := spin.Load()
			if err != nil {
				return fmt.Errorf("parsing spin: %w", err)
			}

			fmt.Println(m)

			return nil
		},
	}

	f := scaffold.Flags()
	f.StringVar(&sc.ManifestDestination, "k8s-dest", "", "the destination for the kubernetes manifests")
	f.StringVarP(&sc.Type, "type", "t", "", "the type of kubernetes manifest to use. Options are Helm, Kustomize, and Kube")
	f.BoolVar(&sc.Override, "override", false, "override existing Dockerfiles and kubernetes manifests")
	f.BoolVar(&sc.Defaults, "y", false, "accept all default values")
	f.StringVarP(&sc.SpinConfig, "config", "c", "", "the location of spin.toml configuration file")

	return scaffold
}

func setConfig(sc *templates.ScaffoldCmd) error {
	if err := sc.InitConfig(); err != nil {
		return err
	}

	if err := sc.ValidateConfig(); err != nil {
		return err
	}

	return nil
}
