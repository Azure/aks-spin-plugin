package generate

import (
	"bytes"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	apps "k8s.io/client-go/applyconfigurations/apps/v1"
	core "k8s.io/client-go/applyconfigurations/core/v1"
	meta "k8s.io/client-go/applyconfigurations/meta/v1"
	node "k8s.io/client-go/applyconfigurations/node/v1"
	"sigs.k8s.io/yaml"
)

const (
	ymlSeparator = "---\n"
)

var (
	annotations = map[string]string{
		"spin.kubernetes.azure.com/created-by": "true",
	}
)

func Manifests(name, image string) ([]byte, error) {
	// define the objects we want to generate

	// using applyconfiguration types to generate yaml
	// means we only generate yaml with the fields we care about
	ns := core.Namespace(name).WithAnnotations(annotations)
	rc := node.RuntimeClass("wasmtime-spin-v1").
		WithAnnotations(annotations).
		WithHandler("spin").
		WithScheduling(node.Scheduling().WithNodeSelector(map[string]string{
			"kubernetes.azure.com/wasmtime-spin-v0-5-1": "true",
		}))
	appLabels := map[string]string{
		"app": name,
	}
	dep := apps.Deployment(name, *ns.Name).
		WithAnnotations(annotations).
		WithSpec(
			apps.DeploymentSpec().
				WithReplicas(3).
				WithSelector(meta.LabelSelector().WithMatchLabels(appLabels)).
				WithTemplate(core.PodTemplateSpec().
					WithLabels(appLabels).
					WithAnnotations(annotations).
					WithSpec(core.PodSpec().
						WithRuntimeClassName(*rc.Name).
						WithContainers(core.Container().
							WithName(name).
							WithImage(image).
							WithCommand("/"),
						),
					),
				),
		)
	service := core.Service(name, *ns.Name).
		WithAnnotations(annotations).
		WithSpec(core.ServiceSpec().
			WithSelector(appLabels).
			WithType(corev1.ServiceTypeLoadBalancer).
			WithPorts(core.ServicePort().
				WithProtocol(corev1.ProtocolTCP).
				WithPort(80).
				WithTargetPort(intstr.FromInt32(80)),
			),
		)

	objs := []interface{}{
		ns,
		rc,
		dep,
		service,
		// TODO: add kwasm operator deployment
	}

	// marshal to yaml
	var buf bytes.Buffer
	for i, obj := range objs {
		out, err := yaml.Marshal(obj)
		if err != nil {
			return nil, fmt.Errorf("error marshaling object; err: %s", err.Error())
		}

		if i != 0 {
			if _, err := buf.WriteString(ymlSeparator); err != nil {
				return nil, fmt.Errorf("writing separator: %w", err)
			}
		}

		if _, err := buf.Write(out); err != nil {
			return nil, fmt.Errorf("writing object: %w", err)
		}
	}

	return buf.Bytes(), nil
}
