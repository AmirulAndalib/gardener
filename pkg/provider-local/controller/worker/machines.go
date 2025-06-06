// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package worker

import (
	"context"
	"encoding/json"
	"fmt"

	machinev1alpha1 "github.com/gardener/machine-controller-manager/pkg/apis/machine/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gardener/gardener/extensions/pkg/controller/worker"
	genericworkeractuator "github.com/gardener/gardener/extensions/pkg/controller/worker/genericactuator"
	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	v1beta1constants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	extensionsv1alpha1helper "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1/helper"
	api "github.com/gardener/gardener/pkg/provider-local/apis/local"
	"github.com/gardener/gardener/pkg/provider-local/controller/infrastructure"
	"github.com/gardener/gardener/pkg/provider-local/local"
	machineproviderlocal "github.com/gardener/gardener/pkg/provider-local/machine-provider/local"
)

// DeployMachineClasses generates and creates the local provider specific machine classes.
func (w *workerDelegate) DeployMachineClasses(ctx context.Context) error {
	if w.machineClasses == nil {
		if err := w.generateMachineConfig(ctx); err != nil {
			return err
		}
	}

	for _, obj := range w.machineClassSecrets {
		if err := w.client.Patch(ctx, obj, client.Apply, local.FieldOwner, client.ForceOwnership); err != nil {
			return fmt.Errorf("failed to apply machine class secret %s: %w", obj.GetName(), err)
		}
	}

	for _, obj := range w.machineClasses {
		if err := w.client.Patch(ctx, obj, client.Apply, local.FieldOwner, client.ForceOwnership); err != nil {
			return fmt.Errorf("failed to apply machine class %s: %w", obj.GetName(), err)
		}
	}

	return nil
}

// GenerateMachineDeployments generates the configuration for the desired machine deployments.
func (w *workerDelegate) GenerateMachineDeployments(ctx context.Context) (worker.MachineDeployments, error) {
	if w.machineDeployments == nil {
		if err := w.generateMachineConfig(ctx); err != nil {
			return nil, err
		}
	}
	return w.machineDeployments, nil
}

func (w *workerDelegate) generateMachineConfig(ctx context.Context) error {
	var (
		machineClassSecrets []*corev1.Secret
		machineClasses      []*machinev1alpha1.MachineClass
		machineImages       []api.MachineImage
		machineDeployments  worker.MachineDeployments
	)

	for _, pool := range w.worker.Spec.Pools {
		workerPoolHash, err := worker.WorkerPoolHash(pool, w.cluster, nil, nil, nil)
		if err != nil {
			return err
		}

		image, err := w.findMachineImage(pool.MachineImage.Name, pool.MachineImage.Version)
		if err != nil {
			return err
		}
		machineImages = appendMachineImage(machineImages, api.MachineImage{
			Name:    pool.MachineImage.Name,
			Version: pool.MachineImage.Version,
			Image:   image,
		})

		userData, err := worker.FetchUserData(ctx, w.client, w.worker.Namespace, pool)
		if err != nil {
			return err
		}

		var (
			deploymentName = fmt.Sprintf("%s-%s", w.worker.Namespace, pool.Name)
			className      = fmt.Sprintf("%s-%s", deploymentName, workerPoolHash)
		)

		machineClassSecrets = append(machineClassSecrets, &corev1.Secret{
			TypeMeta: metav1.TypeMeta{
				APIVersion: corev1.SchemeGroupVersion.String(),
				Kind:       "Secret",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      className,
				Namespace: w.worker.Namespace,
				Labels:    map[string]string{v1beta1constants.GardenerPurpose: v1beta1constants.GardenPurposeMachineClass},
			},
			Type: corev1.SecretTypeOpaque,
			Data: map[string][]byte{"userData": userData},
		})

		providerConfig := map[string]interface{}{
			"image": image,
		}

		for _, ipFamily := range w.cluster.Shoot.Spec.Networking.IPFamilies {
			key := "ipPoolNameV4"
			if ipFamily == gardencorev1beta1.IPFamilyIPv6 {
				key = "ipPoolNameV6"
			}

			providerConfig[key] = infrastructure.IPPoolName(w.worker.Namespace, string(ipFamily))
		}

		providerConfigBytes, err := json.Marshal(providerConfig)
		if err != nil {
			return err
		}

		machineClasses = append(machineClasses, &machinev1alpha1.MachineClass{
			TypeMeta: metav1.TypeMeta{
				APIVersion: machinev1alpha1.SchemeGroupVersion.String(),
				Kind:       "MachineClass",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      className,
				Namespace: w.worker.Namespace,
			},
			SecretRef: &corev1.SecretReference{
				Name:      className,
				Namespace: w.worker.Namespace,
			},
			CredentialsSecretRef: &corev1.SecretReference{
				Name:      w.worker.Spec.SecretRef.Name,
				Namespace: w.worker.Spec.SecretRef.Namespace,
			},
			Provider:     local.Type,
			ProviderSpec: runtime.RawExtension{Raw: providerConfigBytes},
		})

		updateConfiguration := machinev1alpha1.UpdateConfiguration{
			MaxUnavailable: &pool.MaxUnavailable,
			MaxSurge:       &pool.MaxSurge,
		}

		machineDeploymentStrategy := machinev1alpha1.MachineDeploymentStrategy{
			Type: machinev1alpha1.RollingUpdateMachineDeploymentStrategyType,
			RollingUpdate: &machinev1alpha1.RollingUpdateMachineDeployment{
				UpdateConfiguration: updateConfiguration,
			},
		}

		switch ptr.Deref(pool.UpdateStrategy, "") {
		case gardencorev1beta1.AutoInPlaceUpdate:
			machineDeploymentStrategy = machinev1alpha1.MachineDeploymentStrategy{
				Type: machinev1alpha1.InPlaceUpdateMachineDeploymentStrategyType,
				InPlaceUpdate: &machinev1alpha1.InPlaceUpdateMachineDeployment{
					UpdateConfiguration: updateConfiguration,
					OrchestrationType:   machinev1alpha1.OrchestrationTypeAuto,
				},
			}
		case gardencorev1beta1.ManualInPlaceUpdate:
			machineDeploymentStrategy = machinev1alpha1.MachineDeploymentStrategy{
				Type: machinev1alpha1.InPlaceUpdateMachineDeploymentStrategyType,
				InPlaceUpdate: &machinev1alpha1.InPlaceUpdateMachineDeployment{
					UpdateConfiguration: updateConfiguration,
					OrchestrationType:   machinev1alpha1.OrchestrationTypeManual,
				},
			}
		}

		machineDeployments = append(machineDeployments, worker.MachineDeployment{
			Name:                         deploymentName,
			ClassName:                    className,
			SecretName:                   className,
			Minimum:                      pool.Minimum,
			Maximum:                      pool.Maximum,
			Strategy:                     machineDeploymentStrategy,
			PoolName:                     pool.Name,
			Priority:                     pool.Priority,
			Labels:                       pool.Labels,
			Annotations:                  pool.Annotations,
			Taints:                       pool.Taints,
			MachineConfiguration:         genericworkeractuator.ReadMachineConfiguration(pool),
			ClusterAutoscalerAnnotations: extensionsv1alpha1helper.GetMachineDeploymentClusterAutoscalerAnnotations(pool.ClusterAutoscaler),
		})
	}

	w.machineClassSecrets = machineClassSecrets
	w.machineClasses = machineClasses
	w.machineImages = machineImages
	w.machineDeployments = machineDeployments

	return nil
}

func (w *workerDelegate) PreReconcileHook(_ context.Context) error { return nil }
func (w *workerDelegate) PostReconcileHook(ctx context.Context) error {
	// Rewrite the /etc/os-release file for all machine pods to ensure that the version reflects the current machine image version for in-place update tests.
	// Overwrite only if Machine Image Version is not present to prevent overwriting the new version after an in-place update.

	podList := &corev1.PodList{}
	if err := w.client.List(ctx, podList, client.InNamespace(w.worker.Namespace), client.MatchingLabels{
		"app":              "machine",
		"machine-provider": "local",
	}); err != nil {
		return fmt.Errorf("failed to list machine pods: %w", err)
	}

	for _, pod := range podList.Items {
		_, _, err := w.podExecutor.Execute(ctx,
			pod.Namespace,
			pod.Name,
			machineproviderlocal.MachinePodContainerName,
			"sh",
			"-c",
			`if ! grep -q "Machine Image Version" /etc/os-release; then sed -i 's/^PRETTY_NAME="[^"]*"/PRETTY_NAME="Machine Image Version 1.0.0 (version overwritten for tests, check VERSION_ID for actual version)"/' /etc/os-release; fi`,
		)

		if err != nil {
			return fmt.Errorf("failed to execute command in machine pod %s: %w", pod.Name, err)
		}
	}

	return nil
}

func (w *workerDelegate) PreDeleteHook(_ context.Context) error  { return nil }
func (w *workerDelegate) PostDeleteHook(_ context.Context) error { return nil }
