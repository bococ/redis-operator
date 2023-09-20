/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	// DefaultImageRepository is the default docker image repository for the operator
	DefaultImageRepository = "bitnami/redis"
	// DefaultImageTag is the default docker image tag for the operator
	DefaultImageTag = "7.0.12"
	// DefaultImagePullPolicy is the default docker image pull policy for the operator
	DefaultImagePullPolicy = corev1.PullIfNotPresent
	// DefaultServiceType is the default service type for the operator
	DefaultServiceType = corev1.ServiceTypeClusterIP
	// DefaultSize is the default size for the operator
	DefaultSize = "10Gi"
)

// RedisSpec defines the desired state of Redis
type RedisSpec struct {
	Image ImageSpec `json:"image"`

	// +kubebuilder:validation=Optional
	// +kubebuilder:default=1
	Replicas int32 `json:"replicas"`

	Resource *corev1.ResourceRequirements `json:"resource"`

	// +kubebuilder:validation:Optional
	SecurityContext *corev1.SecurityContext `json:"securityContext"`

	// +kubebuilder:validation:Optional
	PodSecurityContext *corev1.PodSecurityContext `json:"podSecurityContext"`

	// +kubebuilder:validation:Optional
	Service *ServiceSpec `json:"service"`

	// +kubebuilder:validation:Optional
	NodeSelector map[string]string `json:"nodeSelector"`

	// +kubebuilder:validation:Optional
	Tolerations []corev1.Toleration `json:"tolerations"`

	// +kubebuilder:validation:Optional
	Persistence *PersistenceSpec `json:"persistence"`
}

func (Redis *Redis) GetLabels() map[string]string {
	return map[string]string{
		"app": Redis.Name,
	}
}

type ImageSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=bitnami/redis
	Repository string `json:"repository,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=latest
	Tag string `json:"tag,omitempty"`

	// +kubebuilder:validation:Enum=Always;Never;IfNotPresent
	// +kubebuilder:default=IfNotPresent
	PullPolicy corev1.PullPolicy `json:"pullPolicy,omitempty"`
}

func (Redis *Redis) GetImageTag() string {
	image := Redis.Spec.Image.Repository
	if image == "" {
		image = DefaultImageRepository
	}
	tag := Redis.Spec.Image.Tag
	if tag == "" {
		tag = DefaultImageTag
	}
	return image + ":" + tag
}

func (Redis *Redis) GetImagePullPolicy() corev1.PullPolicy {
	pullPolicy := Redis.Spec.Image.PullPolicy
	if pullPolicy == "" {
		pullPolicy = DefaultImagePullPolicy
	}

	return pullPolicy
}

func (Redis *Redis) GetNameWithSuffix(name string) string {
	return Redis.GetName() + "-" + name
}

func (Redis *Redis) GetPvcName() string {
	if Redis.Spec.Persistence.Existing() {
		return *Redis.Spec.Persistence.ExistingClaim
	}
	return Redis.GetNameWithSuffix("-pvc")
}

func (Redis *Redis) GetSvcName() string {
	return Redis.GetNameWithSuffix("-svc")
}

type ServiceSpec struct {
	// +kubebuilder:validation:Optional
	Annotations map[string]string `json:"annotations,omitempty"`
	// +kubebuilder:validation:enum=ClusterIP;NodePort;LoadBalancer;ExternalName
	// +kubebuilder:default=ClusterIP
	Type corev1.ServiceType `json:"type"`

	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	// +kubebuilder:default=18080
	Port int32 `json:"port"`
}

func (Redis *Redis) GetServiceType() corev1.ServiceType {
	serviceType := Redis.Spec.Service.Type
	if serviceType == "" {
		serviceType = DefaultServiceType
	}
	return serviceType
}

type PersistenceSpec struct {
	// +kubebuilder:validation:Optional
	StorageClass *string `json:"storageClass,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default={ReadWriteOnce}
	AccessModes []corev1.PersistentVolumeAccessMode `json:"accessModes,omitempty"`

	// +kubebuilder:default="10Gi"
	Size string `json:"size,omitempty"`

	// +kubebuilder:validation:Optional
	ExistingClaim *string `json:"existingClaim,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=Filesystem
	VolumeMode *corev1.PersistentVolumeMode `json:"volumeMode,omitempty"`

	// +kubebuilder:validation:Optional
	Labels map[string]string `json:"labels,omitempty"`

	// +kubebuilder:validation:Optional
	Annotations map[string]string `json:"annotations,omitempty"`
}

func (p *PersistenceSpec) Existing() bool {
	return p.ExistingClaim != nil
}

// GetSize
// get persistence size from instance, if is "" use default value then
// return <Size>
func (p *PersistenceSpec) GetSize() resource.Quantity {
	size := p.Size
	if size == "" {
		return resource.MustParse(DefaultSize)
	}
	return resource.MustParse(size)

}

// RedisStatus defines the observed state of Redis
type RedisStatus struct {
	Nodes      []string                    `json:"nodes"`
	Conditions []corev1.ComponentCondition `json:"conditions"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Redis is the Schema for the redis API
type Redis struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RedisSpec   `json:"spec,omitempty"`
	Status RedisStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RedisList contains a list of Redis
type RedisList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Redis `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Redis{}, &RedisList{})
}
