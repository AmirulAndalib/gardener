// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package gardener

import (
	"context"
	"errors"
	"fmt"
	"slices"

	apiequality "k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gardener/gardener/pkg/api"
	"github.com/gardener/gardener/pkg/apis/core"
	gardencorehelper "github.com/gardener/gardener/pkg/apis/core/helper"
	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	v1beta1constants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	gardencorev1beta1listers "github.com/gardener/gardener/pkg/client/core/listers/core/v1beta1"
	"github.com/gardener/gardener/pkg/features"
	"github.com/gardener/gardener/pkg/utils"
)

// ImagesContext is a helper struct to consume cloud profile images and their versions.
type ImagesContext[A any, B any] struct {
	Images map[string]A

	createVersionsMap func(A) map[string]B
	// imageVersions will be calculated lazily on first access of each key.
	imageVersions map[string]map[string]B
}

// GetCloudProfile determines whether the given shoot references a CloudProfile or a NamespacedCloudProfile and returns the appropriate object.
func GetCloudProfile(ctx context.Context, reader client.Reader, shoot *gardencorev1beta1.Shoot) (*gardencorev1beta1.CloudProfile, error) {
	cloudProfileReference := BuildV1beta1CloudProfileReference(shoot)
	if cloudProfileReference == nil {
		return nil, fmt.Errorf("could not determine cloudprofile from shoot")
	}
	var cloudProfile *gardencorev1beta1.CloudProfile
	switch cloudProfileReference.Kind {
	case v1beta1constants.CloudProfileReferenceKindCloudProfile:
		cloudProfile = &gardencorev1beta1.CloudProfile{}
		if err := reader.Get(ctx, client.ObjectKey{Name: cloudProfileReference.Name}, cloudProfile); err != nil {
			return nil, err
		}
	case v1beta1constants.CloudProfileReferenceKindNamespacedCloudProfile:
		namespacedCloudProfile := &gardencorev1beta1.NamespacedCloudProfile{}
		if err := reader.Get(ctx, client.ObjectKey{Name: cloudProfileReference.Name, Namespace: shoot.Namespace}, namespacedCloudProfile); err != nil {
			return nil, err
		}
		cloudProfile = &gardencorev1beta1.CloudProfile{
			ObjectMeta: metav1.ObjectMeta{
				Name:      cloudProfileReference.Name,
				Namespace: shoot.Namespace,
			},
			Spec: namespacedCloudProfile.Status.CloudProfileSpec,
		}
	}
	return cloudProfile, nil
}

// BuildV1beta1CloudProfileReference determines and returns the CloudProfile reference of the given shoot,
// depending on the availability of cloudProfileName and cloudProfile.
func BuildV1beta1CloudProfileReference(shoot *gardencorev1beta1.Shoot) *gardencorev1beta1.CloudProfileReference {
	if shoot == nil {
		return nil
	}
	if shoot.Spec.CloudProfile != nil {
		cloudProfileReference := shoot.Spec.CloudProfile.DeepCopy()
		if len(cloudProfileReference.Kind) == 0 {
			cloudProfileReference.Kind = v1beta1constants.CloudProfileReferenceKindCloudProfile
		}
		return cloudProfileReference
	}
	if len(ptr.Deref(shoot.Spec.CloudProfileName, "")) > 0 {
		return &gardencorev1beta1.CloudProfileReference{
			Name: *shoot.Spec.CloudProfileName,
			Kind: v1beta1constants.CloudProfileReferenceKindCloudProfile,
		}
	}
	return nil
}

// GetCloudProfileSpec determines whether the given shoot references a CloudProfile or a NamespacedCloudProfile and returns the appropriate CloudProfileSpec.
func GetCloudProfileSpec(cloudProfileLister gardencorev1beta1listers.CloudProfileLister, namespacedCloudProfileLister gardencorev1beta1listers.NamespacedCloudProfileLister, shoot *core.Shoot) (*gardencorev1beta1.CloudProfileSpec, error) {
	cloudProfileReference := BuildCoreCloudProfileReference(shoot)
	if cloudProfileReference == nil {
		return nil, fmt.Errorf("no cloudprofile reference has been provided")
	}
	switch cloudProfileReference.Kind {
	case v1beta1constants.CloudProfileReferenceKindNamespacedCloudProfile:
		namespacedCloudProfile, err := namespacedCloudProfileLister.NamespacedCloudProfiles(shoot.Namespace).Get(cloudProfileReference.Name)
		if err != nil {
			return nil, err
		}
		return &namespacedCloudProfile.Status.CloudProfileSpec, nil
	case v1beta1constants.CloudProfileReferenceKindCloudProfile:
		cloudProfile, err := cloudProfileLister.Get(cloudProfileReference.Name)
		if err != nil {
			return nil, err
		}
		return &cloudProfile.Spec, nil
	}
	return nil, fmt.Errorf("could not find referenced cloudprofile")
}

// ValidateCloudProfileChanges validates that the referenced CloudProfile only changes within the current profile hierarchy
// (i.e. between the parent CloudProfile and the descendant NamespacedCloudProfiles) and that upon changing the profile all
// current configurations still stay valid.
func ValidateCloudProfileChanges(cloudProfileLister gardencorev1beta1listers.CloudProfileLister, namespacedCloudProfileLister gardencorev1beta1listers.NamespacedCloudProfileLister, newShoot, oldShoot *core.Shoot) error {
	oldCloudProfileReference := BuildCoreCloudProfileReference(oldShoot)
	if oldCloudProfileReference == nil {
		return nil
	}
	newCloudProfileReference := BuildCoreCloudProfileReference(newShoot)
	if apiequality.Semantic.DeepEqual(oldCloudProfileReference, newCloudProfileReference) {
		return nil
	}

	newCloudProfileRoot, err := getRootCloudProfile(namespacedCloudProfileLister, newCloudProfileReference, newShoot.Namespace)
	if err != nil {
		return err
	}
	oldCloudProfileRoot, err := getRootCloudProfile(namespacedCloudProfileLister, oldCloudProfileReference, oldShoot.Namespace)
	if err != nil {
		return err
	}

	if !apiequality.Semantic.DeepEqual(oldCloudProfileRoot, newCloudProfileRoot) {
		fromProfile := fmt.Sprintf("%q", oldCloudProfileReference.Name)
		if oldCloudProfileReference.Kind != v1beta1constants.CloudProfileReferenceKindCloudProfile {
			fromProfile += fmt.Sprintf(" (root: %q)", oldCloudProfileRoot.Name)
		}
		toProfile := fmt.Sprintf("%q", newCloudProfileReference.Name)
		if newCloudProfileReference.Kind != v1beta1constants.CloudProfileReferenceKindCloudProfile {
			toProfile += fmt.Sprintf(" (root: %q)", newCloudProfileRoot.Name)
		}
		return fmt.Errorf("cloud profile reference change is invalid: cannot change from %s to %s. The cloud profile reference must remain within the same hierarchy", fromProfile, toProfile)
	}

	if !apiequality.Semantic.DeepEqual(newCloudProfileReference, oldCloudProfileReference) {
		newCloudProfileSpec, err := GetCloudProfileSpec(cloudProfileLister, namespacedCloudProfileLister, newShoot)
		if err != nil {
			return fmt.Errorf("could not find cloudProfileSpec from the shoot cloudProfile reference: %s", err.Error())
		}
		newCloudProfileSpecCore := &core.CloudProfileSpec{}
		if err := api.Scheme.Convert(newCloudProfileSpec, newCloudProfileSpecCore, nil); err != nil {
			return err
		}
		oldCloudProfileSpec, err := GetCloudProfileSpec(cloudProfileLister, namespacedCloudProfileLister, oldShoot)
		if err != nil {
			return fmt.Errorf("could not find cloudProfileSpec from the shoot cloudProfile reference: %s", err.Error())
		}
		oldCloudProfileSpecCore := &core.CloudProfileSpec{}
		if err := api.Scheme.Convert(oldCloudProfileSpec, oldCloudProfileSpecCore, nil); err != nil {
			return err
		}

		// Check that the target cloud profile still supports the currently used machine types, machine images and volume types.
		// No need to check for Kubernetes versions, as the NamespacedCloudProfile could have only extended a version so with the next maintenance the Shoot will be updated to a supported version.
		_, removedMachineImageVersions, _, _ := gardencorehelper.GetMachineImageDiff(oldCloudProfileSpecCore.MachineImages, newCloudProfileSpecCore.MachineImages)
		machineTypes := utils.CreateMapFromSlice(newCloudProfileSpec.MachineTypes, func(mt gardencorev1beta1.MachineType) string { return mt.Name })
		volumeTypes := utils.CreateMapFromSlice(newCloudProfileSpec.VolumeTypes, func(vt gardencorev1beta1.VolumeType) string { return vt.Name })

		for _, w := range newShoot.Spec.Provider.Workers {
			if len(removedMachineImageVersions) > 0 && w.Machine.Image != nil {
				if removedVersions, exists := removedMachineImageVersions[w.Machine.Image.Name]; exists {
					if removedVersions.Has(w.Machine.Image.Version) {
						return fmt.Errorf("newly referenced cloud profile does not contain the machine image version \"%s@%s\" currently in use by worker \"%s\"", w.Machine.Image.Name, w.Machine.Image.Version, w.Name)
					}
				}
			}

			if _, exists := machineTypes[w.Machine.Type]; !exists {
				return fmt.Errorf("newly referenced cloud profile does not contain the machine type %q currently in use by worker %q", w.Machine.Type, w.Name)
			}

			if w.Volume != nil && w.Volume.Type != nil {
				if _, exists := volumeTypes[*w.Volume.Type]; !exists {
					return fmt.Errorf("newly referenced cloud profile does not contain the volume type %q currently in use by worker %q", *w.Volume.Type, w.Name)
				}
			}
		}
	}
	return nil
}

// getRootCloudProfile determines the root CloudProfile from a CloudProfileReference containing any (Namespaced)CloudProfile
func getRootCloudProfile(namespacedCloudProfileLister gardencorev1beta1listers.NamespacedCloudProfileLister, cloudProfile *gardencorev1beta1.CloudProfileReference, namespace string) (*gardencorev1beta1.CloudProfileReference, error) {
	if cloudProfile == nil {
		return nil, errors.New("unexpected nil cloudprofile to get root of")
	}
	switch cloudProfile.Kind {
	case v1beta1constants.CloudProfileReferenceKindCloudProfile:
		return cloudProfile, nil
	case v1beta1constants.CloudProfileReferenceKindNamespacedCloudProfile:
		cp, err := namespacedCloudProfileLister.NamespacedCloudProfiles(namespace).Get(cloudProfile.Name)
		if err != nil {
			return nil, err
		}
		return getRootCloudProfile(namespacedCloudProfileLister, &cp.Spec.Parent, namespace)
	}
	return nil, fmt.Errorf("unexpected cloudprofile kind %s", cloudProfile.Kind)
}

// BuildCoreCloudProfileReference determines and returns the CloudProfile reference of the given shoot,
// depending on the availability of cloudProfileName and cloudProfile.
func BuildCoreCloudProfileReference(shoot *core.Shoot) *gardencorev1beta1.CloudProfileReference {
	if shoot == nil {
		return nil
	}
	if shoot.Spec.CloudProfile != nil {
		cloudProfileV1Beta1 := &gardencorev1beta1.CloudProfileReference{}
		if err := api.Scheme.Convert(shoot.Spec.CloudProfile, cloudProfileV1Beta1, nil); err != nil {
			return nil
		}
		if len(cloudProfileV1Beta1.Kind) == 0 {
			cloudProfileV1Beta1.Kind = v1beta1constants.CloudProfileReferenceKindCloudProfile
		}
		return cloudProfileV1Beta1
	}
	if len(ptr.Deref(shoot.Spec.CloudProfileName, "")) > 0 {
		return &gardencorev1beta1.CloudProfileReference{
			Name: *shoot.Spec.CloudProfileName,
			Kind: v1beta1constants.CloudProfileReferenceKindCloudProfile,
		}
	}
	return nil
}

// SyncCloudProfileFields handles the coexistence of a Shoot Spec's cloudProfileName and cloudProfile
// by making sure both fields are synced correctly and appropriate fallback cases are handled.
func SyncCloudProfileFields(oldShoot, newShoot *core.Shoot) {
	if newShoot.DeletionTimestamp != nil {
		return
	}

	// clear cloudProfile if namespacedCloudProfile is newly provided but feature toggle is disabled
	if newShoot.Spec.CloudProfile != nil && newShoot.Spec.CloudProfile.Kind == v1beta1constants.CloudProfileReferenceKindNamespacedCloudProfile && !utilfeature.DefaultFeatureGate.Enabled(features.UseNamespacedCloudProfile) &&
		(oldShoot == nil || oldShoot.Spec.CloudProfile == nil || oldShoot.Spec.CloudProfile.Kind != v1beta1constants.CloudProfileReferenceKindNamespacedCloudProfile) {
		newShoot.Spec.CloudProfile = nil
	}

	// fill empty cloudProfile field from cloudProfileName, if provided
	if newShoot.Spec.CloudProfile == nil && newShoot.Spec.CloudProfileName != nil {
		newShoot.Spec.CloudProfile = &core.CloudProfileReference{
			Kind: v1beta1constants.CloudProfileReferenceKindCloudProfile,
			Name: *newShoot.Spec.CloudProfileName,
		}
	}

	// default empty cloudprofile kind to CloudProfile
	if newShoot.Spec.CloudProfile != nil && newShoot.Spec.CloudProfile.Kind == "" {
		newShoot.Spec.CloudProfile.Kind = v1beta1constants.CloudProfileReferenceKindCloudProfile
	}

	// fill cloudProfileName from cloudProfile if provided and kind is CloudProfile
	// for backwards compatibility (esp. Dashboard), the CloudProfileName field is synced here with the referenced CloudProfile
	// TODO(LucaBernstein): Remove this block as soon as the CloudProfileName field is deprecated.
	if newShoot.Spec.CloudProfile != nil && newShoot.Spec.CloudProfile.Kind == v1beta1constants.CloudProfileReferenceKindCloudProfile {
		newShoot.Spec.CloudProfileName = &newShoot.Spec.CloudProfile.Name
	}

	// if other than CloudProfile is provided, unset cloudProfileName
	if newShoot.Spec.CloudProfile != nil && newShoot.Spec.CloudProfile.Kind != v1beta1constants.CloudProfileReferenceKindCloudProfile {
		newShoot.Spec.CloudProfileName = nil
	}
}

// SyncArchitectureCapabilityFields syncs the architecture capabilities and the architecture fields.
func SyncArchitectureCapabilityFields(newCloudProfileSpec core.CloudProfileSpec, oldCloudProfileSpec core.CloudProfileSpec) {
	hasCapabilities := len(newCloudProfileSpec.Capabilities) > 0
	if !hasCapabilities || !gardencorehelper.HasCapability(newCloudProfileSpec.Capabilities, v1beta1constants.ArchitectureName) {
		return
	}

	isInitialMigration := hasCapabilities && len(oldCloudProfileSpec.Capabilities) == 0

	// During the initial migration to capabilities, synchronize the architecture fields with the capability definitions.
	// After the migration, only sync architectures from the capability definitions back to the architecture fields.
	// This approach ensures that capabilities are consistently used once defined.
	// Any mismatch between capabilities and architecture fields will result in a validation error.
	syncMachineImageArchitectureCapabilities(newCloudProfileSpec.MachineImages, oldCloudProfileSpec.MachineImages, isInitialMigration)
	syncMachineTypeArchitectureCapabilities(newCloudProfileSpec.MachineTypes, oldCloudProfileSpec.MachineTypes, isInitialMigration)
}

func syncMachineImageArchitectureCapabilities(newMachineImages, oldMachineImages []core.MachineImage, isInitialMigration bool) {
	oldMachineImagesMap := NewCoreImagesContext(oldMachineImages)

	for imageIdx, image := range newMachineImages {
		for versionIdx, version := range newMachineImages[imageIdx].Versions {
			oldMachineImageVersion, oldVersionExists := oldMachineImagesMap.GetImageVersion(image.Name, version.Version)
			capabilityArchitectures := gardencorehelper.ExtractArchitecturesFromCapabilitySets(version.CapabilitySets)

			// Skip any architecture field syncing if
			// - architecture field has been modified and changed to any value other than empty.
			architecturesFieldHasBeenChanged := oldVersionExists && len(version.Architectures) > 0 &&
				!apiequality.Semantic.DeepEqual(oldMachineImageVersion.Architectures, version.Architectures)

			// - both the architecture field and the architecture capability are empty or filled equally.
			if architecturesFieldHasBeenChanged || slices.Equal(capabilityArchitectures, version.Architectures) {
				continue
			}

			// Sync architecture field to capabilities if filled on initial migration.
			if isInitialMigration && len(version.Architectures) > 0 && len(version.CapabilitySets) == 0 {
				for _, architecture := range version.Architectures {
					newMachineImages[imageIdx].Versions[versionIdx].CapabilitySets = append(newMachineImages[imageIdx].Versions[versionIdx].CapabilitySets,
						core.CapabilitySet{
							Capabilities: core.Capabilities{
								v1beta1constants.ArchitectureName: []string{architecture},
							},
						})
				}
				continue
			}

			// Sync capability architectures to architectures field.
			if len(capabilityArchitectures) > 0 {
				newMachineImages[imageIdx].Versions[versionIdx].Architectures = capabilityArchitectures
			}
		}
	}
}

func syncMachineTypeArchitectureCapabilities(newMachineTypes, oldMachineTypes []core.MachineType, isInitialMigration bool) {
	oldMachineTypesMap := utils.CreateMapFromSlice(oldMachineTypes, func(machineType core.MachineType) string { return machineType.Name })

	for i, machineType := range newMachineTypes {
		oldMachineType, oldMachineTypeExists := oldMachineTypesMap[machineType.Name]
		architectureValue := ptr.Deref(machineType.Architecture, "")
		oldArchitectureValue := ptr.Deref(oldMachineType.Architecture, "")
		capabilityArchitectures := machineType.Capabilities[v1beta1constants.ArchitectureName]

		// Skip any architecture field syncing if
		// - architecture field has been modified and changed to any value other than empty.
		architectureFieldHasBeenChanged := oldMachineTypeExists && architectureValue != "" &&
			(oldArchitectureValue == "" || oldArchitectureValue != architectureValue)
		// - both the architecture field and the architecture capability are empty or filled equally.
		architecturesInSync := len(capabilityArchitectures) == 0 && architectureValue == "" ||
			len(capabilityArchitectures) == 1 && capabilityArchitectures[0] == architectureValue
		if architectureFieldHasBeenChanged || architecturesInSync {
			continue
		}

		// Sync architecture field to capabilities if filled on initial migration.
		if isInitialMigration && architectureValue != "" && len(capabilityArchitectures) == 0 {
			if newMachineTypes[i].Capabilities == nil {
				newMachineTypes[i].Capabilities = make(core.Capabilities)
			}
			newMachineTypes[i].Capabilities[v1beta1constants.ArchitectureName] = []string{architectureValue}
			continue
		}

		// Sync capability architecture to architecture field.
		if len(capabilityArchitectures) == 1 {
			newMachineTypes[i].Architecture = ptr.To(capabilityArchitectures[0])
		}
	}
}

// GetImage returns the image with the given name.
func (vc *ImagesContext[A, B]) GetImage(imageName string) (A, bool) {
	o, exists := vc.Images[imageName]
	return o, exists
}

// GetImageVersion returns the image version with the given name and version.
func (vc *ImagesContext[A, B]) GetImageVersion(imageName string, version string) (B, bool) {
	o, exists := vc.getImageVersions(imageName)[version]
	return o, exists
}

func (vc *ImagesContext[A, B]) getImageVersions(imageName string) map[string]B {
	if versions, exists := vc.imageVersions[imageName]; exists {
		return versions
	}
	vc.imageVersions[imageName] = vc.createVersionsMap(vc.Images[imageName])
	return vc.imageVersions[imageName]
}

// NewImagesContext creates a new generic ImagesContext.
func NewImagesContext[A any, B any](images map[string]A, createVersionsMap func(A) map[string]B) *ImagesContext[A, B] {
	return &ImagesContext[A, B]{
		Images:            images,
		createVersionsMap: createVersionsMap,
		imageVersions:     make(map[string]map[string]B),
	}
}

// NewCoreImagesContext creates a new ImagesContext for core.MachineImage.
func NewCoreImagesContext(profileImages []core.MachineImage) *ImagesContext[core.MachineImage, core.MachineImageVersion] {
	return NewImagesContext(
		utils.CreateMapFromSlice(profileImages, func(mi core.MachineImage) string { return mi.Name }),
		func(mi core.MachineImage) map[string]core.MachineImageVersion {
			return utils.CreateMapFromSlice(mi.Versions, func(v core.MachineImageVersion) string { return v.Version })
		},
	)
}

// NewV1beta1ImagesContext creates a new ImagesContext for gardencorev1beta1.MachineImage.
func NewV1beta1ImagesContext(parentImages []gardencorev1beta1.MachineImage) *ImagesContext[gardencorev1beta1.MachineImage, gardencorev1beta1.MachineImageVersion] {
	return NewImagesContext(
		utils.CreateMapFromSlice(parentImages, func(mi gardencorev1beta1.MachineImage) string { return mi.Name }),
		func(mi gardencorev1beta1.MachineImage) map[string]gardencorev1beta1.MachineImageVersion {
			return utils.CreateMapFromSlice(mi.Versions, func(v gardencorev1beta1.MachineImageVersion) string { return v.Version })
		},
	)
}
