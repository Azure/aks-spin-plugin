package generate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/azure/spin-aks-plugin/pkg/azure"
	"github.com/azure/spin-aks-plugin/pkg/config"
	"github.com/azure/spin-aks-plugin/pkg/spin"
	"github.com/azure/spin-aks-plugin/pkg/usererror"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	apps "k8s.io/client-go/applyconfigurations/apps/v1"
	core "k8s.io/client-go/applyconfigurations/core/v1"
	meta "k8s.io/client-go/applyconfigurations/meta/v1"
	node "k8s.io/client-go/applyconfigurations/node/v1"
	secv1 "sigs.k8s.io/secrets-store-csi-driver/apis/v1"
	"sigs.k8s.io/yaml"
)

const (
	ymlSeparator = "---\n"
	secretKey    = "secretKey"
	spcName      = "spin-aks-spc"
)

var (
	annotations = map[string]string{
		"spin.kubernetes.azure.com/created-by": "aks-spin-plugin",
	}
)

// TODO: replace image string once Dockerfile is created/scaffolding is tested
func Manifests(ctx context.Context, sm spin.Manifest, image string) ([]byte, error) {
	// define the objects we want to generate

	// using applyconfiguration types to generate yaml
	// means we only generate yaml with the fields we care about
	ns := core.Namespace(sm.Name).WithAnnotations(annotations)
	rc := node.RuntimeClass("wasmtime-spin-v1").
		WithAnnotations(annotations).
		WithHandler("spin").
		WithScheduling(node.Scheduling().WithNodeSelector(map[string]string{
			"kubernetes.azure.com/wasmtime-spin-v0-5-1": "true",
		}))
	appLabels := map[string]string{
		"app": sm.Name,
	}
	dep := apps.Deployment(sm.Name, *ns.Name).
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
							WithName(sm.Name).
							WithImage(image).
							WithCommand("/").
							WithVolumeMounts( // https://learn.microsoft.com/en-us/azure/aks/csi-secrets-store-identity-access#access-with-a-user-assigned-managed-identity
								&core.VolumeMountApplyConfiguration{
									Name:      to.StringPtr("spin-aks-spc-volume"),
									ReadOnly:  to.BoolPtr(true),
									MountPath: to.StringPtr("/mnt/secrets"), // this won't actually matter since secrets will be accessed via env
								}),
						).WithVolumes(
						&core.VolumeApplyConfiguration{
							Name: to.StringPtr("spin-aks-spc-volume"),
							VolumeSourceApplyConfiguration: core.VolumeSourceApplyConfiguration{
								CSI: &core.CSIVolumeSourceApplyConfiguration{
									Driver:           to.StringPtr("secrets-store.csi.k8s.io"),
									ReadOnly:         to.BoolPtr(true),
									VolumeAttributes: map[string]string{"secretProviderClass": spcName},
								},
							},
						}),
					),
				),
		)

	var envVars []*core.EnvVarApplyConfiguration
	var secretObjects []*secv1.SecretObject
	var paramArray []string

	for k, v := range sm.Variables {
		if v.Secret {
			envVars = append(envVars,
				&core.EnvVarApplyConfiguration{
					Name: to.StringPtr("SPIN_CONFIG_" + k),
					ValueFrom: &core.EnvVarSourceApplyConfiguration{
						SecretKeyRef: &core.SecretKeySelectorApplyConfiguration{
							LocalObjectReferenceApplyConfiguration: core.LocalObjectReferenceApplyConfiguration{
								Name: to.StringPtr(k),
							},
							Key: to.StringPtr(secretKey),
						},
					},
				},
			)

			secretObjects = append(secretObjects,
				&secv1.SecretObject{
					SecretName: k,
					Type:       "secret",
					Data: []*secv1.SecretObjectData{
						{
							ObjectName: k,
							Key:        secretKey,
						},
					},
				})

			p := map[string]interface{}{
				"objectName":    k,
				"objectType":    "secret",
				"objectVersion": "",
			}

			params, err := json.Marshal(p)
			if err != nil {
				return nil, usererror.New(fmt.Errorf("failed to generate SecretProviderClass"), fmt.Sprintf("failed to generate parameters for spc: %s", err))
			}

			paramArray = append(paramArray, string(params))
			continue
		}
		envVars = append(envVars,
			&core.EnvVarApplyConfiguration{
				Name:  to.StringPtr("SPIN_CONFIG_" + k),
				Value: to.StringPtr(v.Def),
			})
	}

	conf := config.Get()
	cluster, err := azure.GetManagedCluster(ctx, conf.Cluster.Subscription, conf.Cluster.ResourceGroup, conf.Cluster.Name)
	if err != nil {
		return nil, usererror.New(fmt.Errorf("failed to retrieve cluster information"), fmt.Sprintf("failed to retrieve cluster information while setting up manifests for keyvault: %s", err))
	}
	clusterId := *cluster.Identity.PrincipalID

	kv := config.GetKeyVault()
	if kv.Name == "" || kv.Subscription == "" || kv.ResourceGroup == "" {
		return nil, usererror.New(fmt.Errorf("no keyvault found in config"), "no keyvault found in config but secrets were detected. Try running `spin aks init` first")
	}

	objects, err := json.Marshal(map[string]interface{}{"array": paramArray})
	if err != nil {
		return nil, usererror.New(fmt.Errorf("failed to generate SecretProviderClass"), fmt.Sprintf("failed to generate objects for spc: %s", err))
	}

	// https://learn.microsoft.com/en-us/azure/aks/csi-secrets-store-identity-access#access-with-a-user-assigned-managed-identity
	spc := &secv1.SecretProviderClass{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "secrets-store.csi.x-k8s.io/v1",
			Kind:       "SecretProviderClass",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      spcName,
			Namespace: *ns.Name,
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion: *dep.APIVersion,
				Controller: to.BoolPtr(true),
				Kind:       *dep.Kind,
				Name:       *dep.Name,
				UID:        "uid",
			}},
		},
		Spec: secv1.SecretProviderClassSpec{
			Provider:      "azure",
			SecretObjects: secretObjects,
			// https://azure.github.io/secrets-store-csi-driver-provider-azure/docs/getting-started/usage/#create-your-own-secretproviderclass-object
			Parameters: map[string]string{
				"keyvaultName":           kv.Name,
				"useVMManagedIdentity":   "true",
				"userAssignedIdentityID": clusterId,
				"tenantId":               conf.TenantID,
				"objects":                string(objects),
			},
		},
	}

	dep.Spec.Template.Spec.Containers[0] = *dep.Spec.Template.Spec.Containers[0].WithEnv(envVars...)

	service := core.Service(sm.Name, *ns.Name).
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
		spc,
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
