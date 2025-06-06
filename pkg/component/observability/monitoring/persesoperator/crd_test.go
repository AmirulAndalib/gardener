// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package persesoperator_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/gardener/gardener/pkg/client/kubernetes"
	"github.com/gardener/gardener/pkg/component"
	. "github.com/gardener/gardener/pkg/component/observability/monitoring/persesoperator"
	"github.com/gardener/gardener/pkg/utils/test/matchers"
)

var _ = Describe("CRD", func() {
	var (
		ctx          = context.TODO()
		c            client.Client
		deployWaiter component.DeployWaiter
	)

	BeforeEach(func() {
		var err error
		c = fake.NewClientBuilder().WithScheme(kubernetes.SeedScheme).Build()

		mapper := meta.NewDefaultRESTMapper([]schema.GroupVersion{apiextensionsv1.SchemeGroupVersion})
		mapper.Add(apiextensionsv1.SchemeGroupVersion.WithKind("CustomResourceDefinition"), meta.RESTScopeRoot)
		applier := kubernetes.NewApplier(c, mapper)

		deployWaiter, err = NewCRDs(c, applier)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("#Deploy", func() {
		It("should deploy the CRD", func() {
			Expect(deployWaiter.Deploy(ctx)).To(Succeed())
			Expect(c.Get(ctx, client.ObjectKey{Name: "perses.perses.dev"}, &apiextensionsv1.CustomResourceDefinition{})).To(Succeed())
			Expect(c.Get(ctx, client.ObjectKey{Name: "persesdashboards.perses.dev"}, &apiextensionsv1.CustomResourceDefinition{})).To(Succeed())
			Expect(c.Get(ctx, client.ObjectKey{Name: "persesdatasources.perses.dev"}, &apiextensionsv1.CustomResourceDefinition{})).To(Succeed())
		})
	})

	Describe("#Destroy", func() {
		It("should delete the CRD", func() {
			Expect(deployWaiter.Destroy(ctx)).To(Succeed())
			Expect(c.Get(ctx, client.ObjectKey{Name: "perses.perses.dev"}, &apiextensionsv1.CustomResourceDefinition{})).To(matchers.BeNotFoundError())
			Expect(c.Get(ctx, client.ObjectKey{Name: "persesdashboards.perses.dev"}, &apiextensionsv1.CustomResourceDefinition{})).To(matchers.BeNotFoundError())
			Expect(c.Get(ctx, client.ObjectKey{Name: "persesdatasources.perses.dev"}, &apiextensionsv1.CustomResourceDefinition{})).To(matchers.BeNotFoundError())
		})
	})
})
