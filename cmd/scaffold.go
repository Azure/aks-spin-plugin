package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/azure/spin-aks-plugin/pkg/config"
	"github.com/azure/spin-aks-plugin/pkg/generate"
	"github.com/azure/spin-aks-plugin/pkg/logger"
	"github.com/azure/spin-aks-plugin/pkg/spin"
	"github.com/azure/spin-aks-plugin/pkg/usererror"
	"github.com/spf13/cobra"
)

var (
	k8sDest    string
	dockerDest string
	override   bool
)

func init() {
	addOverrideFlag(dockerfileCmd)
	addOverrideFlag(k8sCmd)
	dockerfileCmd.Flags().StringVarP(&dockerDest, "dest", "d", "./Dockerfile", "destination Dockerfile path")
	k8sCmd.Flags().StringVarP(&k8sDest, "dest", "d", "./manifests/manifests.yaml", "destination yaml file path")

	scaffoldCmd.AddCommand(dockerfileCmd)
	scaffoldCmd.AddCommand(k8sCmd)
	rootCmd.AddCommand(scaffoldCmd)
}

func addOverrideFlag(cmd *cobra.Command) {
	f := cmd.Flags()
	f.BoolVar(&override, "override", false, "override existing files")
}

var scaffoldCmd = &cobra.Command{
	Use:   "scaffold",
	Short: "Generates required files",
	Long:  "Creates files required to run your application on AKS",
}

var dockerfileCmd = &cobra.Command{
	Use:   "dockerfile",
	Short: "Generates Dockerfile",
	Long:  "Creates Dockerfile required to run your application on AKS",
	RunE: func(cmd *cobra.Command, args []string) error {
		lgr := logger.FromContext(cmd.Context())
		lgr.Debug("starting dockerfile command")

		spinManifest := config.Get().SpinManifest
		if spinManifest == "" {
			return usererror.New(errors.New("spin manifest not set in config"), "Spin manifest not set in config. Try running `spin aks init`.")
		}

		manifest, err := spin.Load(spinManifest)
		if err != nil {
			return fmt.Errorf("loading spin manifest: %w", err)
		}

		manifestRelativePath, err := filepath.Rel(filepath.Dir(dockerDest), spinManifest)
		if err != nil {
			return fmt.Errorf("getting relative path: %w", err)
		}
		diff := filepath.Dir(manifestRelativePath)

		sources := make([]generate.Source, 0, len(manifest.Components))
		for _, component := range manifest.Components {
			// guard against unsupported things
			if component.Source.URLSource.Url != "" {
				return usererror.New(
					errors.New("component source is a URL"),
					fmt.Sprintf("Component source %s is a URL which isn't currently supported.", component.Id),
				)
			}

			for _, file := range component.Files.MapFiles {
				if file.Source != "" {
					return usererror.New(
						errors.New("file source exists"),
						fmt.Sprintf("Component %s contains files which isn't currently supported.", component.Id),
					)
				}
			}

			for _, file := range component.Files.StringFiles {
				if file != "" {
					return usererror.New(
						errors.New("file source exists"),
						fmt.Sprintf("Component %s contains files which isn't currently supported.", component.Id),
					)
				}
			}

			cleaned := filepath.Clean(string(component.Source.StringSource))
			sources = append(sources, generate.Source{
				Path:     cleaned,
				Relative: filepath.Join(diff, cleaned),
			})
		}

		dockerfile, err := generate.Dockerfile(generate.DockerfileOpt{
			SpinManifest: manifestRelativePath,
			Sources:      sources,
		})
		if err != nil {
			return fmt.Errorf("generating Dockerfile: %w", err)
		}

		if _, err := os.Stat(dockerDest); err == nil && !override {
			return usererror.New(
				errors.New("file exists"),
				fmt.Sprintf("File %s already exists. Use --override to overwrite.", dockerDest),
			)
		} else if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("checking file existence: %w", err)
		}

		if err := os.MkdirAll(filepath.Dir(dockerDest), 0755); err != nil {
			return fmt.Errorf("creating directory: %w", err)
		}

		f, err := os.OpenFile(dockerDest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			return fmt.Errorf("opening file: %w", err)
		}
		defer f.Close()

		if _, err := f.Write(dockerfile); err != nil {
			return fmt.Errorf("writing file: %w", err)
		}

		lgr.Info("Dockerfile written to " + dockerDest)
		lgr.Debug("finished dockerfile command")
		return nil
	},
}

var k8sCmd = &cobra.Command{
	Use:   "k8s",
	Short: "Generates Kubernetes manifests",
	Long:  "Creates Kubernetes manifests required to run your application on AKS",
	RunE: func(cmd *cobra.Command, args []string) error {
		lgr := logger.FromContext(cmd.Context())
		lgr.Info("starting k8s command")

		spinManifest := config.Get().SpinManifest
		if spinManifest == "" {
			return usererror.New(errors.New("spin manifest not set in config"), "Spin manifest not set in config. Try running `spin aks init`.")
		}

		manifest, err := spin.Load(spinManifest)
		if err != nil {
			return fmt.Errorf("loading spin manifest: %w", err)
		}

		name := manifest.Name
		if name == "" {
			return usererror.New(errors.New("name not set in spin manifest"), "Name not set in spin manifest. Add a name to your spin manifest and try again.")
		}

		manifests, err := generate.Manifests(name, "placeholderimagefornow:latest")
		if err != nil {
			return fmt.Errorf("generating manifests: %w", err)
		}

		if _, err := os.Stat(k8sDest); err == nil && !override {
			return usererror.New(
				errors.New("file exists"),
				fmt.Sprintf("File %s already exists. Use --override to overwrite.", k8sDest),
			)
		}

		if err := os.MkdirAll(filepath.Dir(k8sDest), 0755); err != nil {
			return fmt.Errorf("creating directory: %w", err)
		}

		f, err := os.OpenFile(k8sDest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			return fmt.Errorf("opening file: %w", err)
		}
		defer f.Close()

		if _, err := f.Write(manifests); err != nil {
			return fmt.Errorf("writing file: %w", err)
		}

		lgr.Info("Kubernetes manifests written to " + k8sDest)
		lgr.Info("finished k8s command")
		return nil
	},
}
