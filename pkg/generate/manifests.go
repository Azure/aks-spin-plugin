package generate

import (
	"bytes"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"gopkg.in/yaml.v3"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	nodev1 "k8s.io/api/node/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/intstr"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
)

const (
	ymlSeparator = "\n---\n"
)

var (
	scheme      = runtime.NewScheme()
	annotations = map[string]string{
		"spin.kubernetes.azure.com/created-by": "true",
	}
)

func init() {
	// add any types used to scheme
	appsv1.AddToScheme(scheme)
	corev1.AddToScheme(scheme)
	nodev1.AddToScheme(scheme)
	metav1.AddMetaToScheme(scheme)
}

func Manifests(name, image string) ([]byte, error) {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Annotations: annotations,
		},
	}

	rc := &nodev1.RuntimeClass{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "wasmtime-spin-v1",
			Annotations: annotations,
		},
		Handler: "spin",
		Scheduling: &nodev1.Scheduling{
			NodeSelector: map[string]string{
				"kubernetes.azure.com/wasmtime-spin-v0-5-1": "true",
			},
		},
	}

	appLabels := map[string]string{
		"app": name,
	}

	// todo: add kwasm to these objs

	objs := []runtime.Object{
		ns,
		rc,
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:        name,
				Namespace:   ns.Name,
				Annotations: annotations,
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: to.Ptr(int32(3)),
				Selector: &metav1.LabelSelector{
					MatchLabels: appLabels,
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels:      appLabels,
						Annotations: annotations,
					},
					Spec: corev1.PodSpec{
						RuntimeClassName: to.Ptr(rc.Name),
						Containers: []corev1.Container{
							{
								Name:    name,
								Image:   image,
								Command: []string{"/"},
							},
						},
					},
				},
			},
		},
		&corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:        name,
				Namespace:   ns.Name,
				Annotations: annotations,
			},
			Spec: corev1.ServiceSpec{
				Selector: appLabels,
				Type:     corev1.ServiceTypeLoadBalancer,
				Ports: []corev1.ServicePort{
					{
						Protocol:   corev1.ProtocolTCP,
						Port:       80,
						TargetPort: intstr.FromInt32(80),
					},
				},
			},
		},
	}

	// encode to yaml
	var buf bytes.Buffer
	codec := serializer.NewCodecFactory(scheme).LegacyCodec(scheme.PreferredVersionAllGroups()...)
	for i, obj := range objs {
		json, err := runtime.Encode(codec, obj)
		if err != nil {
			return nil, fmt.Errorf("encoding object: %w", err)
		}

		var decoded map[string]interface{}
		if err := yamlutil.NewYAMLOrJSONDecoder(bytes.NewReader(json), 4096).Decode(&decoded); err != nil {
			return nil, fmt.Errorf("decoding object: %w", err)
		}

		if i != 0 {
			if _, err := buf.WriteString(ymlSeparator); err != nil {
				return nil, fmt.Errorf("writing separator: %w", err)
			}
		}

		out, err := yaml.Marshal(decoded)
		if err != nil {
			return nil, fmt.Errorf("marshaling object to yaml: %w", err)
		}

		if _, err := buf.Write(out); err != nil {
			return nil, fmt.Errorf("writing object %d: %w", i, err)
		}
	}

	return buf.Bytes(), nil
}
