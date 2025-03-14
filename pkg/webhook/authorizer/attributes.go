/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

The utility functions in this file were copied from the kubernetes/kubernetes project
https://github.com/kubernetes/kubernetes/blob/v1.20.2/pkg/registry/authorization/util/helpers.go

Modifications Copyright 2024 SAP SE or an SAP affiliate company and Gardener contributors
*/

package authorizer

import (
	authorizationv1 "k8s.io/api/authorization/v1"
	"k8s.io/apiserver/pkg/authentication/user"
	auth "k8s.io/apiserver/pkg/authorization/authorizer"
)

// ResourceAttributesFrom combines the API object information and the user.Info from the context to build a full
// auth.AttributesRecord for resource access.
func ResourceAttributesFrom(user user.Info, in authorizationv1.ResourceAttributes) auth.AttributesRecord {
	return auth.AttributesRecord{
		User:            user,
		Verb:            in.Verb,
		Namespace:       in.Namespace,
		APIGroup:        in.Group,
		APIVersion:      in.Version,
		Resource:        in.Resource,
		Subresource:     in.Subresource,
		Name:            in.Name,
		ResourceRequest: true,
	}
}

// NonResourceAttributesFrom combines the API object information and the user.Info from the context to build a full
// auth.AttributesRecord for non resource access.
func NonResourceAttributesFrom(user user.Info, in authorizationv1.NonResourceAttributes) auth.AttributesRecord {
	return auth.AttributesRecord{
		User:            user,
		ResourceRequest: false,
		Path:            in.Path,
		Verb:            in.Verb,
	}
}

// AuthorizationAttributesFrom takes a spec and returns the proper authz attributes to check it.
func AuthorizationAttributesFrom(spec authorizationv1.SubjectAccessReviewSpec) auth.AttributesRecord {
	userToCheck := &user.DefaultInfo{
		Name:   spec.User,
		Groups: spec.Groups,
		UID:    spec.UID,
		Extra:  convertToUserInfoExtra(spec.Extra),
	}

	var authorizationAttributes auth.AttributesRecord
	if spec.ResourceAttributes != nil {
		authorizationAttributes = ResourceAttributesFrom(userToCheck, *spec.ResourceAttributes)
	} else if spec.NonResourceAttributes != nil {
		authorizationAttributes = NonResourceAttributesFrom(userToCheck, *spec.NonResourceAttributes)
	}

	return authorizationAttributes
}

func convertToUserInfoExtra(extra map[string]authorizationv1.ExtraValue) map[string][]string {
	if extra == nil {
		return nil
	}
	ret := make(map[string][]string, len(extra))
	for k, v := range extra {
		ret[k] = v
	}

	return ret
}
