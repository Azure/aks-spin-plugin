package prompt

import (
	"errors"
	"fmt"
	"os"

	"github.com/manifoldco/promptui"
)

type InputOpt struct {
	Default  string
	Validate func(string) error
}

func Input(label string, opt *InputOpt) (string, error) {
	p := promptui.Prompt{
		Label:    label,
		Validate: opt.Validate,
		Default:  opt.Default,
	}

	ret, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("running input: %w", err)
	}

	return ret, nil
}

func FileExists(path string) error {
	_, err := os.Stat(path)

	if errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("file %s doesn't exist", path)
	}

	if err != nil {
		return fmt.Errorf("checking if file exists: %w", err)
	}

	return nil
}
