package k8s

import (
	"fmt"
	devopsV1 "github.com/tomoncle/k8s-operator-nginx/api/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"strings"
)

type ResourceType string

const (
	Deployment = ResourceType("deployment")
	Service    = ResourceType("service")
	Ingress    = ResourceType("ingress")
)

func DefaultMap() map[string]string {
	return map[string]string{}
}

func MergeMap(a, b map[string]string) map[string]string {
	if a == nil {
		return b
	}
	for k, v := range b {
		a[k] = v
	}
	return a
}

func LabelsForNginx(name string) map[string]string {
	return map[string]string{
		MakeKeyForNginx("resource-name"): name,
		MakeKeyForNginx("app"):           strings.ToLower(devopsV1.Kind),
	}
}

func MakeKeyForNginx(key string) string {
	return fmt.Sprintf("%s/%s", devopsV1.GroupVersion.Group, key)
}

func GetTypeMeta(res ResourceType) metaV1.TypeMeta {
	switch res {
	case Deployment:
		return metaV1.TypeMeta{Kind: "Deployment", APIVersion: "apps/v1"}
	case Service:
		return metaV1.TypeMeta{Kind: "Service", APIVersion: "v1"}
	case Ingress:
		return metaV1.TypeMeta{Kind: "Ingress", APIVersion: "networking.k8s.io/v1"}
	default:
		var typeMeta metaV1.TypeMeta
		return typeMeta
	}
}

func GetObjectMeta(res ResourceType, n *devopsV1.Nginx, labels, annotations map[string]string) metaV1.ObjectMeta {
	var name string
	switch res {
	case Deployment:
		name = n.Name
	case Service:
		name = fmt.Sprintf("%s-service", n.Name)
	case Ingress:
		name = fmt.Sprintf("%s-ingress", n.Name)
	}
	return metaV1.ObjectMeta{
		Name:        name,
		Namespace:   n.Namespace,
		Labels:      labels,
		Annotations: annotations,
		OwnerReferences: []metaV1.OwnerReference{
			*metaV1.NewControllerRef(n, schema.GroupVersionKind{
				Group:   devopsV1.GroupVersion.Group,
				Version: devopsV1.GroupVersion.Version,
				Kind:    devopsV1.Kind,
			}),
		},
	}
}
