package k8s

import (
	"fmt"
	devopsV1 "github.com/tomoncle/k8s-operator-nginx/api/v1"
	networkingV1 "k8s.io/api/networking/v1"
)

func GetIngressLabels(n *devopsV1.Nginx) map[string]string {
	labels := LabelsForNginx(n.Name)
	if n.Spec.Ingress != nil {
		labels = MergeMap(n.Spec.Ingress.Labels, labels)
	}
	return labels
}

func GetIngressAnnotations(n *devopsV1.Nginx) map[string]string {
	var annotations map[string]string
	if n.Spec.Ingress != nil {
		annotations = MergeMap(n.Spec.Ingress.Annotations, annotations)
	}
	return annotations
}

func GetIngressClassName(n *devopsV1.Nginx) *string {
	if n.Spec.Ingress == nil {
		return nil
	}
	if n.Spec.Ingress.IngressClassName != nil {
		return n.Spec.Ingress.IngressClassName
	}
	defaultIngressClassName := "nginx"
	return &defaultIngressClassName
}

func GetIngressRule(n *devopsV1.Nginx, path, host string) networkingV1.IngressRule {
	return networkingV1.IngressRule{
		Host: host,
		IngressRuleValue: networkingV1.IngressRuleValue{
			HTTP: &networkingV1.HTTPIngressRuleValue{
				Paths: []networkingV1.HTTPIngressPath{
					{
						Path: path,
						PathType: func(pt networkingV1.PathType) *networkingV1.PathType {
							return &pt
						}(networkingV1.PathTypePrefix),
						Backend: networkingV1.IngressBackend{
							Service: &networkingV1.IngressServiceBackend{
								Name: fmt.Sprintf("%s-service", n.Name),
								Port: networkingV1.ServiceBackendPort{
									Name: defaultHTTPPortName,
								},
							},
						},
					},
				},
			},
		},
	}
}

func GetIngressRules(n *devopsV1.Nginx) []networkingV1.IngressRule {
	var rules []networkingV1.IngressRule
	for _, t := range n.Spec.TLS {
		hosts := t.Hosts
		if len(hosts) == 0 {
			// host为空，则会匹配所有的未知域名 或者 IP访问
			hosts = []string{""}
		}
		for _, host := range hosts {
			rules = append(rules, GetIngressRule(n, "/", host))
		}
	}
	return rules
}

func getIngressTLS(n *devopsV1.Nginx) []networkingV1.IngressTLS {
	var tls []networkingV1.IngressTLS
	for _, t := range n.Spec.TLS {
		hosts := t.Hosts
		if len(hosts) == 0 {
			// host为空，则会匹配所有的未知域名 或者 IP访问
			hosts = []string{""}
		}
		tls = append(tls, networkingV1.IngressTLS{
			SecretName: t.SecretName,
			Hosts:      t.Hosts,
		})
	}
	return tls
}

func GetIngressDefaultBackend(n *devopsV1.Nginx) *networkingV1.IngressBackend {
	defaultBackend := &networkingV1.IngressBackend{
		Service: &networkingV1.IngressServiceBackend{
			Name: fmt.Sprintf("%s-service", n.Name),
			Port: networkingV1.ServiceBackendPort{
				Name: defaultHTTPPortName,
			},
		},
	}
	return defaultBackend
}

func NewIngress(n *devopsV1.Nginx) *networkingV1.Ingress {
	return &networkingV1.Ingress{
		TypeMeta:   GetTypeMeta(Ingress),
		ObjectMeta: GetObjectMeta(Ingress, n, GetIngressLabels(n), GetIngressAnnotations(n)),
		Spec: networkingV1.IngressSpec{
			IngressClassName: GetIngressClassName(n),
			Rules:            GetIngressRules(n),
			TLS:              getIngressTLS(n),
			DefaultBackend:   GetIngressDefaultBackend(n),
		},
	}
}
