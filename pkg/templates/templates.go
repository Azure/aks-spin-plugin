package templates

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"text/template"

	"github.com/azure/spin-aks-plugin/pkg/prompt"
	"github.com/manifoldco/promptui"
)

const Helm string = "Helm"
const Kustomize string = "Kustomize"
const Kube string = "Kube"

var typeOptions = []string{Helm, Kustomize, Kube}

var (
	_, b, _, _ = runtime.Caller(0)
	basepath   = filepath.Dir(b)
)

type Config struct{}

type ScaffoldCmd struct {
	ManifestDestination string
	Type                string
	Override            bool
	DockerfilePath      string
	Defaults            bool
	SpinConfig          string
	Source              string
}

type Dockerfile struct {
	ConfigFile string
	Executable string
}

func (sc *ScaffoldCmd) InitConfig() error {
	if sc.Defaults {
		if sc.ManifestDestination == "" {
			sc.ManifestDestination = "."
		}

		if sc.Type == "" {
			sc.Type = Kube
		}

		if sc.SpinConfig == "" {
			sc.SpinConfig = "./aks-spin.toml"
		}
	}

	if sc.ManifestDestination == "" {
		sc.ManifestDestination = "."
	}

	if sc.Type == "" {
		k8sType, err := prompt.Select("Select your kubernetes manifest type", typeOptions, nil)
		if err != nil {
			return fmt.Errorf("error selecting manifest type; err: %s", err.Error())
		}
		sc.Type = k8sType
	}

	if sc.SpinConfig == "" {
		validate := func(input string) error {
			if input == "" {
				return fmt.Errorf("Invalid app name")
			}

			return sc.ValidateBuild()
		}

		prompt := promptui.Prompt{
			Label:    "Enter spin config location",
			Validate: validate,
		}

		result, err := prompt.Run()
		if err != nil {
			return fmt.Errorf("unable to save spin config; err: %s", err.Error())
		}

		sc.SpinConfig = result
	}
	return nil
}

func (sc *ScaffoldCmd) ValidateConfig() error {
	if !slices.Contains(typeOptions, sc.Type) {
		return fmt.Errorf("given kubernetes type is %s, must be %s, %s, or %s", sc.Type, Helm, Kustomize, Kube)
	}

	if sc.Override {
		validate := func(input string) error {
			_, err := os.Stat(input)
			return err
		}

		prompt := promptui.Prompt{
			Label:    "Enter Dockerfile location",
			Validate: validate,
		}

		result, err := prompt.Run()
		if err != nil {
			return fmt.Errorf("error with Dockerfile destination; err: %s", err)
		}

		sc.DockerfilePath = result
	}

	fi, err := os.Stat(sc.ManifestDestination)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		if err := os.MkdirAll(sc.ManifestDestination, 0755); err != nil {
			return fmt.Errorf("could not create %s: %s", sc.ManifestDestination, err)
		}
	} else if err != nil {
		return err
	} else if !fi.IsDir() {
		return fmt.Errorf("manifest destination must be a directory")
	}

	if err := sc.ValidateBuild(); err != nil {
		return err
	}

	return nil
}

// ValidateBuild checks that the spin build command has been run and successfully created the necessary build files. If not
// it will return an error
func (sc *ScaffoldCmd) ValidateBuild() error {
	// validate config exists
	fi, err := os.Stat(sc.SpinConfig)
	if err != nil {
		return fmt.Errorf("error finding spin config at path %s; err: %s", sc.SpinConfig, err.Error())
	} else if fi.IsDir() {
		return fmt.Errorf("%s cannot be a directory", sc.SpinConfig)
	}

	// validate source exe from config exists
	searchCmd := exec.Command(`grep`, `source ./aks-spin.toml`)
	extractCmd := exec.Command(`cut`, `-d ''"' -f 2`)

	pipe, _ := searchCmd.StdoutPipe()
	defer pipe.Close()

	extractCmd.Stdin = pipe
	err = searchCmd.Start()
	if err != nil {
		return fmt.Errorf("unable to find source property from %s; err: %s", sc.SpinConfig, err.Error())
	}

	out, err := extractCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("unable to extract source property from %s; err %s", sc.SpinConfig, err.Error())
	}
	source := string(out)
	fi, err = os.Stat(source)
	if err != nil {
		return fmt.Errorf("error finding build executable at %s", source)
	} else if fi.IsDir() {
		fmt.Errorf("%s cannot be a directory", source)
	}

	sc.Source = source
	return nil
}

func (sc *ScaffoldCmd) PopulateDockerfile() error {
	// create Dockerfile
	content, err := os.ReadFile("./Dockerfile_template")
	if err != nil {
		return fmt.Errorf("unable to read Dockerfile template; err: %s", err.Error())
	}

	tmpl, err := template.New("Dockerfile").Parse(string(content))
	if err != nil {
		return fmt.Errorf("unable to create dockerfile template; err: %s", err.Error())
	}

	var f *os.File
	if !sc.Override {
		f, err = os.Create("Dockerfile")
		if err != nil {
			return fmt.Errorf("unable to create Dockerfile; err: %s", err.Error())
		}
	} else {
		f, err = os.Open(sc.DockerfilePath)
		if err != nil {
			return fmt.Errorf("unable to open Dockerfile; err: %s", err.Error())
		}
	}
	defer f.Close()

	err = tmpl.Execute(f, Dockerfile{ConfigFile: sc.SpinConfig, Executable: sc.Source})
	if err != nil {
		return fmt.Errorf("unable to write information to Dockerfile; err: %s", err.Error())
	}

	return nil
}
