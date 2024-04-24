// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package shoot

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"

	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	e2e "github.com/gardener/gardener/test/e2e/gardener"
)

var _ = Describe("Shoot Tests", Label("Shoot", "default"), func() {
	test := func(shoot *gardencorev1beta1.Shoot) {
		f := defaultShootCreationFramework()
		f.Shoot = shoot

		It("Create and Delete Failed Shoot", Offset(1), func() {
			By("Create Shoot")
			if shoot.Namespace == "" {
				shoot.SetNamespace(f.ProjectNamespace)
			}
			f.Shoot = shoot

			ctx, cancel := context.WithTimeout(parentCtx, 2*time.Minute)
			defer cancel()

			Expect(f.GardenClient.Client().Create(ctx, shoot)).To(Succeed())

			By("Wait until last operation in Shoot is set to Failed")
			Eventually(func(g Gomega) {
				g.Expect(f.GetShoot(ctx, shoot)).To(Succeed())
				g.Expect(shoot.Status.LastOperation).ToNot(BeNil())
				g.Expect(shoot.Status.LastOperation.State).To(Equal(gardencorev1beta1.LastOperationStateFailed))
			}).WithTimeout(1 * time.Minute).Should(Succeed())

			By("Delete Shoot")
			ctx, cancel = context.WithTimeout(parentCtx, 2*time.Minute)
			defer cancel()
			Expect(f.DeleteShootAndWaitForDeletion(ctx, f.Shoot)).To(Succeed())
		})
	}

	Context("Shoot with invalid DNS configuration", func() {
		shoot := e2e.DefaultShoot("e2e-invalid-dns")
		shoot.Spec.DNS = &gardencorev1beta1.DNS{
			Domain: ptr.To("shoot.non-existing-domain"),
		}
		test(shoot)
	})
})
