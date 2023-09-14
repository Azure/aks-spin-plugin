package generate

import (
	"bytes"
	_ "embed"
	"fmt"
	"text/template"
)

var (
	//go:embed Dockerfile.tmpl
	dockerfileTmpl string
)

// DockerfileOpt is the options for the Dockerfile
type DockerfileOpt struct {
	// SpinManifest is the path to the spin manifest from the Dockerfile directory
	SpinManifest string
	// Sources is a list of sources and should be a cleaned path relative to the SpinManifest
	Sources []Source
}

type Source struct {
	// Path is a cleaned path of the source relative to the SpinManifest
	Path string
	// Relative is the relative path of the source relative to the Dockerfile
	Relative string
}

func Dockerfile(d DockerfileOpt) ([]byte, error) {
	if d.SpinManifest == "" {
		return nil, fmt.Errorf("no spin manifest provided")
	}
	if len(d.Sources) == 0 {
		return nil, fmt.Errorf("no sources provided")
	}

	tmpl, err := template.New("dockerfile").Parse(dockerfileTmpl)
	if err != nil {
		return nil, fmt.Errorf("creating template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, d); err != nil {
		return nil, fmt.Errorf("executing template: %w", err)
	}

	return buf.Bytes(), nil
}
