package k8s

import (
	devopsV1 "github.com/tomoncle/k8s-operator-nginx/api/v1"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func GetServiceLabels(n *devopsV1.Nginx) map[string]string {
	labels := map[string]string{}
	if n.Spec.Service != nil {
		labels = n.Spec.Service.Labels
	}
	return MergeMap(labels, LabelsForNginx(n.Name))
}

func GetServiceSelector(n *devopsV1.Nginx) map[string]string {
	selector := LabelsForNginx(n.Name)
	if n.Spec.Service != nil {
		if n.Spec.Service.UsePodSelector != nil && !*n.Spec.Service.UsePodSelector {
			selector = nil
		}
	}
	return selector
}

func GetServiceAnnotations(n *devopsV1.Nginx) map[string]string {
	annotations := map[string]string{}
	if n.Spec.Service != nil {
		if n.Spec.Service.Annotations != nil {
			annotations = n.Spec.Service.Annotations
		}
	}
	return annotations
}

func GetExternalTrafficPolicy(n *devopsV1.Nginx) coreV1.ServiceExternalTrafficPolicyType {
	if n.Spec.Service == nil || n.Spec.Service.Type == coreV1.ServiceTypeClusterIP {
		return ""
	}
	var externalTrafficPolicy coreV1.ServiceExternalTrafficPolicyType
	if n.Spec.Service != nil {
		externalTrafficPolicy = n.Spec.Service.ExternalTrafficPolicy
	}
	return externalTrafficPolicy
}

func GetServiceType(n *devopsV1.Nginx) coreV1.ServiceType {
	if n == nil || n.Spec.Service == nil {
		return coreV1.ServiceTypeClusterIP
	}
	return n.Spec.Service.Type
}

func GetServicePorts() []coreV1.ServicePort {
	ports := []coreV1.ServicePort{
		{
			Name:       defaultHTTPPortName,
			Protocol:   coreV1.ProtocolTCP,
			TargetPort: intstr.FromInt(int(defaultHTTPPort)),
			Port:       int32(80),
		},
		{
			Name:       defaultHTTPSPortName,
			Protocol:   coreV1.ProtocolTCP,
			TargetPort: intstr.FromString(defaultHTTPSPortName),
			Port:       int32(443),
		},
	}
	return ports
}

func NewService(n *devopsV1.Nginx) *coreV1.Service {
	return &coreV1.Service{
		TypeMeta:   GetTypeMeta(Service),
		ObjectMeta: GetObjectMeta(Service, n, GetServiceLabels(n), GetServiceAnnotations(n)),
		Spec: coreV1.ServiceSpec{
			Ports:                 GetServicePorts(),
			Selector:              GetServiceSelector(n),
			Type:                  GetServiceType(n),
			ExternalTrafficPolicy: GetExternalTrafficPolicy(n),
		},
	}
}
