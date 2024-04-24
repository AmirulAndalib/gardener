// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package containerruntime

import (
	"context"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"

	extensionspredicate "github.com/gardener/gardener/extensions/pkg/predicate"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/controllerutils/mapper"
)

const (
	// FinalizerName is the prefix name of the finalizer written by this controller.
	FinalizerName = "extensions.gardener.cloud/containerruntime"
	// ControllerName is the name of the controller.
	ControllerName = "containerruntime"
)

// AddArgs are arguments for adding an ContainerRuntime resources controller to a manager.
type AddArgs struct {
	// Actuator is an ContainerRuntime resource actuator.
	Actuator Actuator
	// FinalizerSuffix is the suffix for the finalizer name.
	FinalizerSuffix string
	// ControllerOptions are the controller options used for creating a controller.
	// The options.Reconciler is always overridden with a reconciler created from the
	// given actuator.
	ControllerOptions controller.Options
	// Predicates are the predicates to use.
	Predicates []predicate.Predicate
	// Resync determines the requeue interval.
	Resync time.Duration
	// Type is the type of the resource considered for reconciliation.
	Type string
	// IgnoreOperationAnnotation specifies whether to ignore the operation annotation or not.
	// If the annotation is not ignored, the extension controller will only reconcile
	// with a present operation annotation typically set during a reconcile (e.g in the maintenance time) by the Gardenlet
	IgnoreOperationAnnotation bool
}

// Add adds an ContainerRuntime controller to the given manager using the given AddArgs.
func Add(ctx context.Context, mgr manager.Manager, args AddArgs) error {
	args.ControllerOptions.Reconciler = NewReconciler(mgr, args.Actuator)
	return add(ctx, mgr, args)
}

// DefaultPredicates returns the default predicates for an containerruntime reconciler.
func DefaultPredicates(ctx context.Context, mgr manager.Manager, ignoreOperationAnnotation bool) []predicate.Predicate {
	return extensionspredicate.DefaultControllerPredicates(ignoreOperationAnnotation, extensionspredicate.ShootNotFailedPredicate(ctx, mgr))
}

func add(ctx context.Context, mgr manager.Manager, args AddArgs) error {
	ctrl, err := controller.New(ControllerName, mgr, args.ControllerOptions)
	if err != nil {
		return err
	}

	predicates := extensionspredicate.AddTypePredicate(args.Predicates, args.Type)

	if args.IgnoreOperationAnnotation {
		if err := ctrl.Watch(
			source.Kind(mgr.GetCache(), &extensionsv1alpha1.Cluster{}),
			mapper.EnqueueRequestsFrom(ctx, mgr.GetCache(), ClusterToContainerResourceMapper(mgr, predicates...), mapper.UpdateWithNew, ctrl.GetLogger()),
		); err != nil {
			return err
		}
	}

	return ctrl.Watch(source.Kind(mgr.GetCache(), &extensionsv1alpha1.ContainerRuntime{}), &handler.EnqueueRequestForObject{}, predicates...)
}
