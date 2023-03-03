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

package v1

import (
	appsV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// 权限配置，如果更新需要重新执行 make manifests
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.currentReplicas,selectorpath=.status.podSelector
// +kubebuilder:printcolumn:name="Current",type=integer,JSONPath=`.status.currentReplicas`
// +kubebuilder:printcolumn:name="Desired",type=integer,JSONPath=`.spec.replicas`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// Nginx is the Schema for the nginxes API
// 使用以上权限配置
type Nginx struct {
	metaV1.TypeMeta   `json:",inline"`
	metaV1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NginxSpec   `json:"spec,omitempty"`
	Status NginxStatus `json:"status,omitempty"`
}

// NginxStatus defines the observed state of Nginx
type NginxStatus struct {
	// CurrentReplicas is the last observed number from the NGINX object.
	CurrentReplicas int32 `json:"currentReplicas,omitempty"`
	// PodSelector is the Nginx pod label selector.
	PodSelector string `json:"podSelector,omitempty"`

	Deployments []DeploymentStatus `json:"deployments,omitempty"`
	Services    []ServiceStatus    `json:"services,omitempty"`
	Ingresses   []IngressStatus    `json:"ingresses,omitempty"`
}

//+kubebuilder:object:root=true

// NginxList contains a list of Nginx
// 使用默认权限即可：kubebuilder:object:root=true
type NginxList struct {
	metaV1.TypeMeta `json:",inline"`
	metaV1.ListMeta `json:"metadata,omitempty"`
	Items           []Nginx `json:"items"`
}

type ConfigKind string

const (
	// ConfigKindConfigMap 配置
	ConfigKindConfigMap = ConfigKind("ConfigMap")
	// ConfigKindInline 在Pod上设置为注释的配置, 并使用Downward API作为文件注入到容器中.
	ConfigKindInline = ConfigKind("Inline")
	Kind             = "Nginx"
)

type NginxIngress struct {
	// Annotations are extra annotations for the Ingress resource.
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
	// Labels are extra labels for the Ingress resource.
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
	// IngressClassName is the class to be set on Ingress.
	// +optional
	IngressClassName *string `json:"ingressClassName,omitempty"`
}

type NginxTLS struct {
	// SecretName is the name of the Secret which contains the certificate-key
	// pair. It must reside in the same Namespace as the Nginx resource.
	//
	// NOTE: The Secret should follow the Kubernetes TLS secrets type.
	// More info: https://kubernetes.io/docs/concepts/configuration/secret/#tls-secrets.
	SecretName string `json:"secretName"`
	// Hosts are a list of hosts included in the TLS certificate. Defaults to the
	// wildcard of hosts: "*".
	// +optional
	Hosts []string `json:"hosts,omitempty"`
}

type NginxService struct {
	// Type is the type of the service. Defaults to the default service type value.
	// +optional
	Type coreV1.ServiceType `json:"type,omitempty"`
	// LoadBalancerIP is an optional load balancer IP for the service.
	// +optional
	LoadBalancerIP string `json:"loadBalancerIP,omitempty"`
	// Labels are extra labels for the service.
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
	// Annotations are extra annotations for the service.
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
	// ExternalTrafficPolicy defines whether external traffic will be routed to
	// node-local or cluster-wide endpoints. Defaults to the default Service
	// externalTrafficPolicy value.
	// +optional
	ExternalTrafficPolicy coreV1.ServiceExternalTrafficPolicyType `json:"externalTrafficPolicy,omitempty"`
	// UsePodSelector defines whether Service should automatically map the
	// endpoints using the pod's label selector. Defaults to true.
	// +optional
	UsePodSelector *bool `json:"usePodSelector,omitempty"`
}

type PodTemplateSpec struct {
	// Affinity to be set on the nginx pod.
	// +optional
	Affinity *coreV1.Affinity `json:"affinity,omitempty"`
	// NodeSelector to be set on the nginx pod.
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	// Annotations are custom annotations to be set into Pod.
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
	// Labels are custom labels to be added into Pod.
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
	// HostNetwork enabled causes the pod to use the host's network namespace.
	// +optional
	HostNetwork bool `json:"hostNetwork,omitempty"`
	// Ports is the list of ports used by nginx.
	// +optional
	Ports []coreV1.ContainerPort `json:"ports,omitempty"`
	// TerminationGracePeriodSeconds defines the max duration seconds which the
	// pod needs to terminate gracefully. Defaults to pod's
	// terminationGracePeriodSeconds default value.
	// +optional
	TerminationGracePeriodSeconds *int64 `json:"terminationGracePeriodSeconds,omitempty"`
	// SecurityContext configures security attributes for the nginx pod.
	// +optional
	SecurityContext *coreV1.SecurityContext `json:"securityContext,omitempty"`
	// Volumes that will attach to nginx instances
	// +optional
	Volumes []coreV1.Volume `json:"volumes,omitempty"`
	// VolumeMounts will mount volume declared above in directories
	// +optional
	VolumeMounts []coreV1.VolumeMount `json:"volumeMounts,omitempty"`
	// InitContainers are executed in order prior to containers being started
	// More info: https://kubernetes.io/docs/concepts/workloads/pods/init-containers/
	// +optional
	InitContainers []coreV1.Container `json:"initContainers,omitempty"`
	// Containers are executed in parallel to the main nginx container
	// +optional
	Containers []coreV1.Container `json:"containers,omitempty"`
	// RollingUpdate defines params to control the desired behavior of rolling update.
	// +optional
	RollingUpdate *appsV1.RollingUpdateDeployment `json:"rollingUpdate,omitempty"`
	// Toleration defines list of taints that pod can tolerate.
	// +optional
	Toleration []coreV1.Toleration `json:"toleration,omitempty"`

	// ServiceAccountName is the name of the ServiceAccount to use to run this nginx instance.
	// +optional
	ServiceAccountName string `json:"serviceAccountName,omitempty"`
}

// ConfigRef is a reference to a config object.
type ConfigRef struct {
	// Kind of the config object. Defaults to "ConfigMap".
	Kind ConfigKind `json:"kind"`
	// Name of the ConfigMap object with "nginx.conf" key inside. It must reside
	// in the same Namespace as the Nginx resource. Required when Kind is "ConfigMap".
	//
	// It's mutually exclusive with Value field.
	// +optional
	Name string `json:"name,omitempty"`
	// Value is the raw Nginx configuration. Required when Kind is "Inline".
	//
	// It's mutually exclusive with Name field.
	// +optional
	Value string `json:"value,omitempty"`
}

// NginxSpec defines the desired state of Nginx
type NginxSpec struct {
	// Replicas是所需pod的数量。默认为 default deployment
	// replicas value.
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`
	// Image是容器镜像名称。默认值为 "nginx:latest".
	// +optional
	Image string `json:"image,omitempty"`
	// Config是对NGINX配置对象的引用，该对象存储NGINX配置文件。如果提供该文件，则将其装载在上的NGINX容器中
	// "/etc/nginx/nginx.conf".
	// +optional
	Config *ConfigRef `json:"config,omitempty"`
	// TLS configuration.
	// +optional
	TLS []NginxTLS `json:"tls,omitempty"`
	// Template used to configure the nginx pod.
	// +optional
	PodTemplate PodTemplateSpec `json:"podTemplate,omitempty"`
	// Service 服务配置
	// +optional
	Service *NginxService `json:"service,omitempty"`
	// Ingress 配置
	// +optional
	Ingress *NginxIngress `json:"ingress,omitempty"`
	// 健康检查路径
	// working or not.
	// +optional
	HealthcheckPath string `json:"healthcheckPath,omitempty"`
	// Resources 资源限制
	// +optional
	Resources coreV1.ResourceRequirements `json:"resources,omitempty"`
}

type DeploymentStatus struct {
	Name string `json:"name"`
}

type ServiceStatus struct {
	Name string `json:"name"`
}

type IngressStatus struct {
	Name string `json:"name"`
}

func init() {
	SchemeBuilder.Register(&Nginx{}, &NginxList{})
}
