package cmd

import (
	"errors"
	"fmt"
	"os/exec"

	"github.com/azure/spin-aks-plugin/pkg/config"
	"github.com/azure/spin-aks-plugin/pkg/logger"
	"github.com/azure/spin-aks-plugin/pkg/usererror"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(buildCmd)
}

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Builds the Spin application",
	Long:  "Builds the Spin application to WASM",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		lgr := logger.FromContext(ctx)
		lgr.Debug("starting build command")

		spinManifest := config.Get().SpinManifest
		if spinManifest == "" {
			return usererror.New(errors.New("spin manifest not set in config"), "Spin manifest not set in config. Try running `spin aks init`.")
		}

		c := exec.Command("spin", "build", "-f", spinManifest)
		lgr.Debug("running command " + c.String())

		if log, err := c.Output(); err != nil {
			if len(log) != 0 {
				lgr.Info(string(log))
			}

			if exitErr, ok := err.(*exec.ExitError); ok {
				return fmt.Errorf("running spin build: %s", exitErr.Stderr)
			}

			return fmt.Errorf("running spin build: %w", err)
		}

		lgr.Debug("finished build command")
		return nil
	},
}
