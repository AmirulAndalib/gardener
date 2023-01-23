// Copyright (c) 2021 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bastion

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/utils/clock"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/ratelimiter"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	gardencorev1alpha1 "github.com/gardener/gardener/pkg/apis/core/v1alpha1"
	v1alpha1helper "github.com/gardener/gardener/pkg/apis/core/v1alpha1/helper"
	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	v1beta1constants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	v1beta1helper "github.com/gardener/gardener/pkg/apis/core/v1beta1/helper"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	operationsv1alpha1 "github.com/gardener/gardener/pkg/apis/operations/v1alpha1"
	"github.com/gardener/gardener/pkg/controllerutils"
	reconcilerutils "github.com/gardener/gardener/pkg/controllerutils/reconciler"
	"github.com/gardener/gardener/pkg/gardenlet/apis/config"
	kubernetesutils "github.com/gardener/gardener/pkg/utils/kubernetes"
)

const (
	// finalizerName is the Kubernetes finalizerName that is used to control the cleanup of
	// Bastion resources in the seed cluster.
	finalizerName = gardencorev1alpha1.GardenerName
)

// RequeueDurationWhenResourceDeletionStillPresent is the duration used for requeueing when owned resources are still in
// the process of being deleted when deleting a Bastion.
var RequeueDurationWhenResourceDeletionStillPresent = 5 * time.Second

// Reconciler reconciles Bastions and deploys them into the seed cluster.
type Reconciler struct {
	GardenClient client.Client
	SeedClient   client.Client
	Config       config.BastionControllerConfiguration
	Clock        clock.Clock
	// RateLimiter allows limiting exponential backoff for testing purposes
	RateLimiter ratelimiter.RateLimiter
}

// Reconcile reconciles Bastions and deploys them into the seed cluster.
func (r *Reconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	log := logf.FromContext(ctx)

	bastion := &operationsv1alpha1.Bastion{}
	if err := r.GardenClient.Get(ctx, request.NamespacedName, bastion); err != nil {
		if apierrors.IsNotFound(err) {
			log.V(1).Info("Object is gone, stop reconciling")
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, fmt.Errorf("error retrieving object from store: %w", err)
	}

	// get Shoot for the bastion
	shoot := gardencorev1beta1.Shoot{}
	shootKey := kubernetesutils.Key(bastion.Namespace, bastion.Spec.ShootRef.Name)
	if err := r.GardenClient.Get(ctx, shootKey, &shoot); err != nil {
		return reconcile.Result{}, fmt.Errorf("could not get shoot %v: %w", shootKey, err)
	}

	var err error
	if bastion.DeletionTimestamp != nil {
		err = r.cleanupBastion(ctx, log, bastion, &shoot)
	} else {
		err = r.reconcileBastion(ctx, log, bastion, &shoot)
	}

	if cause := reconcilerutils.ReconcileErrCause(err); cause != nil {
		log.Error(cause, "Reconciling failed")
	}

	return reconcilerutils.ReconcileErr(err)
}

func (r *Reconciler) reconcileBastion(
	ctx context.Context,
	log logr.Logger,
	bastion *operationsv1alpha1.Bastion,
	shoot *gardencorev1beta1.Shoot,
) error {
	if !controllerutil.ContainsFinalizer(bastion, finalizerName) {
		log.Info("Adding finalizer")
		if err := controllerutils.AddFinalizers(ctx, r.GardenClient, bastion, finalizerName); err != nil {
			return fmt.Errorf("failed to add finalizer: %w", err)
		}
	}

	extensionBastion := newBastionExtension(bastion, shoot)
	extensionIngress := make([]extensionsv1alpha1.BastionIngressPolicy, len(bastion.Spec.Ingress))
	for i, ingress := range bastion.Spec.Ingress {
		extensionIngress[i] = extensionsv1alpha1.BastionIngressPolicy{
			IPBlock: ingress.IPBlock,
		}
	}

	var (
		mustReconcileExtensionBastion = false
		lastObservedError             error
		extensionBastionSpec          = extensionsv1alpha1.BastionSpec{
			DefaultSpec: extensionsv1alpha1.DefaultSpec{
				Type: *bastion.Spec.ProviderType,
			},
			UserData: createUserData(bastion),
			Ingress:  extensionIngress,
		}
	)

	if err := r.SeedClient.Get(ctx, client.ObjectKeyFromObject(extensionBastion), extensionBastion); err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}
		// if the extension Bastion doesn't exist yet, create it
		mustReconcileExtensionBastion = true
	} else if !reflect.DeepEqual(extensionBastion.Spec, extensionBastionSpec) {
		// if the extensionBastionSpec has changed, reconcile it
		mustReconcileExtensionBastion = true
	} else if extensionBastion.Status.LastOperation == nil {
		// if the extension did not record a lastOperation yet, record it as error in the bastion status
		lastObservedError = fmt.Errorf("extension did not record a last operation yet")
	} else {
		lastOperationState := extensionBastion.Status.LastOperation.State
		if extensionBastion.Status.LastError != nil ||
			lastOperationState == gardencorev1beta1.LastOperationStateError ||
			lastOperationState == gardencorev1beta1.LastOperationStateFailed {
			if lastOperationState == gardencorev1beta1.LastOperationStateFailed {
				mustReconcileExtensionBastion = true
			}

			lastObservedError = fmt.Errorf("extension state is not Succeeded but %v", lastOperationState)
			if extensionBastion.Status.LastError != nil {
				lastObservedError = v1beta1helper.NewErrorWithCodes(fmt.Errorf("error during reconciliation: %s", extensionBastion.Status.LastError.Description), extensionBastion.Status.LastError.Codes...)
			}
		}
	}

	if lastObservedError != nil {
		message := fmt.Sprintf("Error while waiting for %s %s/%s to become ready", extensionsv1alpha1.BastionResource, extensionBastion.Namespace, extensionBastion.Name)
		err := v1beta1helper.NewErrorWithCodes(fmt.Errorf("%s: %w", message, lastObservedError), v1beta1helper.DeprecatedDetermineErrorCodes(lastObservedError)...)

		if patchErr := patchReadyCondition(ctx, r.GardenClient, bastion, gardencorev1alpha1.ConditionFalse, "FailedReconciling", err.Error()); patchErr != nil {
			log.Error(patchErr, "Failed patching ready condition")
		}
	}

	if mustReconcileExtensionBastion {
		if _, err := controllerutils.GetAndCreateOrMergePatch(ctx, r.SeedClient, extensionBastion, func() error {
			metav1.SetMetaDataAnnotation(&extensionBastion.ObjectMeta, v1beta1constants.GardenerOperation, v1beta1constants.GardenerOperationReconcile)
			metav1.SetMetaDataAnnotation(&extensionBastion.ObjectMeta, v1beta1constants.GardenerTimestamp, r.Clock.Now().UTC().String())

			extensionBastion.Spec = extensionBastionSpec
			return nil
		}); err != nil {
			if patchErr := patchReadyCondition(ctx, r.GardenClient, bastion, gardencorev1alpha1.ConditionFalse, "FailedReconciling", err.Error()); patchErr != nil {
				log.Error(patchErr, "Failed patching ready condition")
			}
			return fmt.Errorf("failed to ensure bastion extension resource: %w", err)
		}
		// return early here, the Bastion status will be updated by the reconciliation caused by the extension Bastion status update.
		return nil
	}

	if extensionBastion.Status.LastOperation != nil && extensionBastion.Status.LastOperation.State == gardencorev1beta1.LastOperationStateSucceeded {
		// copy over the extension's status to the operation bastion and set the condition
		patch := client.MergeFrom(bastion.DeepCopy())
		setReadyCondition(bastion, gardencorev1alpha1.ConditionTrue, "SuccessfullyReconciled", "The bastion has been reconciled successfully.")
		bastion.Status.Ingress = extensionBastion.Status.Ingress.DeepCopy()
		bastion.Status.ObservedGeneration = &bastion.Generation
		if err := r.GardenClient.Status().Patch(ctx, bastion, patch); err != nil {
			return fmt.Errorf("failed patching ready condition of Bastion: %w", err)
		}
	}

	return nil
}

func (r *Reconciler) cleanupBastion(
	ctx context.Context,
	log logr.Logger,
	bastion *operationsv1alpha1.Bastion,
	shoot *gardencorev1beta1.Shoot,
) error {
	if !sets.NewString(bastion.Finalizers...).Has(finalizerName) {
		return nil
	}

	if err := patchReadyCondition(ctx, r.GardenClient, bastion, gardencorev1alpha1.ConditionFalse, "DeletionInProgress", "The bastion is being deleted."); err != nil {
		return fmt.Errorf("failed patching ready condition of Bastion: %w", err)
	}

	// delete bastion extension resource in seed cluster
	extensionBastion := newBastionExtension(bastion, shoot)
	if err := r.SeedClient.Delete(ctx, extensionBastion); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("Successfully deleted")

			if controllerutil.ContainsFinalizer(bastion, finalizerName) {
				log.Info("Removing finalizer")
				if err := controllerutils.RemoveFinalizers(ctx, r.GardenClient, bastion, finalizerName); err != nil {
					return fmt.Errorf("failed to remove finalizer: %w", err)
				}
			}

			return nil
		}

		return fmt.Errorf("failed to delete bastion extension resource: %w", err)
	}

	// cleanup is now triggered on the seed, requeue to wait for it to happen
	return &reconcilerutils.RequeueAfterError{
		RequeueAfter: RequeueDurationWhenResourceDeletionStillPresent,
		Cause:        errors.New("bastion extension cleanup has not completed yet"),
	}
}

func newBastionExtension(bastion *operationsv1alpha1.Bastion, shoot *gardencorev1beta1.Shoot) *extensionsv1alpha1.Bastion {
	return &extensionsv1alpha1.Bastion{
		ObjectMeta: metav1.ObjectMeta{
			Name:      bastion.Name,
			Namespace: shoot.Status.TechnicalID,
		},
	}
}

func setReadyCondition(bastion *operationsv1alpha1.Bastion, status gardencorev1alpha1.ConditionStatus, reason string, message string) {
	condition := v1alpha1helper.GetOrInitCondition(bastion.Status.Conditions, operationsv1alpha1.BastionReady)
	condition = v1alpha1helper.UpdatedCondition(condition, status, reason, message)

	bastion.Status.Conditions = v1alpha1helper.MergeConditions(bastion.Status.Conditions, condition)
}

func patchReadyCondition(ctx context.Context, c client.StatusClient, bastion *operationsv1alpha1.Bastion, status gardencorev1alpha1.ConditionStatus, reason string, message string) error {
	patch := client.MergeFrom(bastion.DeepCopy())
	setReadyCondition(bastion, status, reason, message)
	return c.Status().Patch(ctx, bastion, patch)
}

func createUserData(bastion *operationsv1alpha1.Bastion) []byte {
	userData := fmt.Sprintf(`#!/bin/bash -eu

id gardener || useradd gardener -mU
mkdir -p /home/gardener/.ssh
echo "%s" > /home/gardener/.ssh/authorized_keys
chown gardener:gardener /home/gardener/.ssh/authorized_keys
echo "gardener ALL=(ALL) NOPASSWD:ALL" >/etc/sudoers.d/99-gardener-user
`, bastion.Spec.SSHPublicKey)

	return []byte(userData)
}
