package k8s

import (
	"encoding/json"
	"fmt"
	devopsV1 "github.com/tomoncle/k8s-operator-nginx/api/v1"
	appsV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"math"
	"strings"
)

const (
	defaultImage                = "tomoncle/webhook:latest"
	defaultHTTPPort             = int32(80)
	defaultHTTPHostNetworkPort  = int32(80)
	defaultHTTPPortName         = "http"
	defaultHTTPSPort            = int32(443)
	defaultHTTPSHostNetworkPort = int32(443)
	defaultHTTPSPortName        = "https"
	curlProbeCommand            = "curl -m%d -kfsS -o /dev/null %s"
	configMountPath             = "/etc/nginx"
	configFileName              = "nginx.conf"
)

func findContainerPort(podSpec *devopsV1.PodTemplateSpec, name string) *coreV1.ContainerPort {
	for i, port := range podSpec.Ports {
		if port.Name == name {
			return &podSpec.Ports[i]
		}
	}
	return nil
}

func makeContainerPort(name string, port int32) coreV1.ContainerPort {
	return coreV1.ContainerPort{
		Name:          name,
		ContainerPort: port,
		Protocol:      coreV1.ProtocolTCP,
	}
}

func getContainerPorts(n *devopsV1.Nginx) []coreV1.ContainerPort {
	podSpec := &n.Spec.PodTemplate
	if findContainerPort(podSpec, defaultHTTPPortName) == nil {
		httpPort := defaultHTTPPort
		if podSpec.HostNetwork {
			httpPort = defaultHTTPHostNetworkPort
		}
		podSpec.Ports = append(podSpec.Ports, makeContainerPort(defaultHTTPPortName, httpPort))
	}
	if findContainerPort(podSpec, defaultHTTPSPortName) == nil {
		httpsPort := defaultHTTPSPort
		if podSpec.HostNetwork {
			httpsPort = defaultHTTPSHostNetworkPort
		}
		podSpec.Ports = append(podSpec.Ports, makeContainerPort(defaultHTTPSPortName, httpsPort))
	}
	return podSpec.Ports
}

func hasLowPort(ports []coreV1.ContainerPort) bool {
	for _, port := range ports {
		if port.ContainerPort < 1024 {
			return true
		}
	}
	return false
}

func getDeploymentAnnotations(spec devopsV1.NginxSpec) map[string]string {
	origSpec, err := json.Marshal(spec)
	if err != nil {
		return DefaultMap()
	}
	return map[string]string{MakeKeyForNginx("generated-from"): string(origSpec)}
}

func setConfigRef(conf *devopsV1.ConfigRef, deploy *appsV1.Deployment) {
	if conf == nil {
		return
	}
	volumeName := "nginx-config"

	containerVolumeMounts := deploy.Spec.Template.Spec.Containers[0].VolumeMounts
	deploy.Spec.Template.Spec.Containers[0].VolumeMounts = append(
		containerVolumeMounts,
		coreV1.VolumeMount{
			Name:      volumeName,
			MountPath: fmt.Sprintf("%s/%s", configMountPath, configFileName),
			SubPath:   configFileName,
			ReadOnly:  true,
		})

	deploymentVolumes := deploy.Spec.Template.Spec.Volumes
	switch conf.Kind {
	case devopsV1.ConfigKindConfigMap:
		deploy.Spec.Template.Spec.Volumes = append(deploymentVolumes,
			coreV1.Volume{
				Name: volumeName,
				VolumeSource: coreV1.VolumeSource{
					ConfigMap: &coreV1.ConfigMapVolumeSource{
						LocalObjectReference: coreV1.LocalObjectReference{
							Name: conf.Name,
						},
						Optional: func(b bool) *bool { return &b }(false),
					},
				},
			})
	case devopsV1.ConfigKindInline:
		if deploy.Spec.Template.Annotations == nil {
			deploy.Spec.Template.Annotations = make(map[string]string)
		}

		key := MakeKeyForNginx("custom-nginx-config")
		deploy.Spec.Template.Annotations[key] = conf.Value

		deploy.Spec.Template.Spec.Volumes = append(deploymentVolumes,
			coreV1.Volume{
				Name: volumeName,
				VolumeSource: coreV1.VolumeSource{
					DownwardAPI: &coreV1.DownwardAPIVolumeSource{
						Items: []coreV1.DownwardAPIVolumeFile{
							{
								Path: "nginx.conf",
								FieldRef: &coreV1.ObjectFieldSelector{
									FieldPath: fmt.Sprintf("metadata.annotations['%s']", key),
								},
							},
						},
					},
				},
			})
	}
}

// 健康检查1
func getContainerProbes(n *devopsV1.Nginx) *coreV1.Probe {
	httpPort := findContainerPort(&n.Spec.PodTemplate, defaultHTTPPortName)
	//  没有找到端口，不配置
	if httpPort == nil {
		return nil
	}
	return &coreV1.Probe{
		TimeoutSeconds:      int32(1), // 超时时间
		InitialDelaySeconds: int32(5), // 延迟5秒触发
		ProbeHandler: coreV1.ProbeHandler{
			HTTPGet: &coreV1.HTTPGetAction{
				Path: n.Spec.HealthcheckPath,
				Port: intstr.FromInt(int(httpPort.ContainerPort)),
			},
		},
	}
}

// 健康检查2
func _(spec devopsV1.NginxSpec, dep *appsV1.Deployment) {
	httpPort := findContainerPort(&spec.PodTemplate, defaultHTTPPortName)
	cmdTimeoutSec := int32(1)
	var commands []string
	if httpPort != nil {
		httpURL := fmt.Sprintf("http://localhost:%d%s", httpPort.ContainerPort, spec.HealthcheckPath)
		commands = append(commands, fmt.Sprintf(curlProbeCommand, cmdTimeoutSec, httpURL))
	}
	if len(spec.TLS) > 0 {
		httpsPort := findContainerPort(&spec.PodTemplate, defaultHTTPSPortName)
		if httpsPort != nil {
			httpsURL := fmt.Sprintf("https://localhost:%d%s", httpsPort.ContainerPort, spec.HealthcheckPath)
			commands = append(commands, fmt.Sprintf(curlProbeCommand, cmdTimeoutSec, httpsURL))
		}
	}
	if len(commands) == 0 {
		return
	}
	dep.Spec.Template.Spec.Containers[0].ReadinessProbe = &coreV1.Probe{
		TimeoutSeconds:      cmdTimeoutSec * int32(len(commands)),
		InitialDelaySeconds: int32(5), // 延迟5秒触发
		ProbeHandler: coreV1.ProbeHandler{
			Exec: &coreV1.ExecAction{
				Command: []string{"sh", "-c", strings.Join(commands, " && ")},
			},
		},
	}

}

func getSecurityContext(n *devopsV1.Nginx) *coreV1.SecurityContext {
	securityContext := n.Spec.PodTemplate.SecurityContext
	if hasLowPort(n.Spec.PodTemplate.Ports) {
		if securityContext == nil {
			securityContext = &coreV1.SecurityContext{}
		}
		if securityContext.Capabilities == nil {
			securityContext.Capabilities = &coreV1.Capabilities{}
		}
		securityContext.Capabilities.Add = append(securityContext.Capabilities.Add, "NET_BIND_SERVICE")
	}
	return securityContext
}

func getDeploymentStrategy(n *devopsV1.Nginx) appsV1.DeploymentStrategy {
	var maxSurge, maxUnavailable *intstr.IntOrString
	if n.Spec.PodTemplate.HostNetwork {
		replicas := int32(1)
		if n.Spec.Replicas != nil && *n.Spec.Replicas > int32(0) {
			replicas = *n.Spec.Replicas
		}
		// 计算期望值
		wishNumber := intstr.FromInt(int(math.Ceil(float64(replicas) * 0.25)))
		maxUnavailable = &wishNumber
		maxSurge = &wishNumber
	}

	if ru := n.Spec.PodTemplate.RollingUpdate; ru != nil {
		maxSurge, maxUnavailable = ru.MaxSurge, ru.MaxUnavailable
	}
	return appsV1.DeploymentStrategy{
		Type: appsV1.RollingUpdateDeploymentStrategyType,
		RollingUpdate: &appsV1.RollingUpdateDeployment{
			MaxUnavailable: maxUnavailable,
			MaxSurge:       maxSurge,
		},
	}
}

func NewDeployment(n *devopsV1.Nginx) (*appsV1.Deployment, error) {
	deployment := appsV1.Deployment{
		TypeMeta:   GetTypeMeta(Deployment),
		ObjectMeta: GetObjectMeta(Deployment, n, LabelsForNginx(n.Name), getDeploymentAnnotations(n.Spec)),
		Spec: appsV1.DeploymentSpec{
			Strategy: getDeploymentStrategy(n),
			Replicas: n.Spec.Replicas,
			Selector: &metaV1.LabelSelector{MatchLabels: LabelsForNginx(n.Name)},
			Template: coreV1.PodTemplateSpec{
				ObjectMeta: metaV1.ObjectMeta{
					Namespace:   n.Namespace,
					Annotations: n.Spec.PodTemplate.Annotations,
					Labels:      MergeMap(LabelsForNginx(n.Name), n.Spec.PodTemplate.Labels),
				},
				Spec: coreV1.PodSpec{
					ServiceAccountName:            n.Spec.PodTemplate.ServiceAccountName,
					EnableServiceLinks:            func(b bool) *bool { return &b }(false),
					InitContainers:                n.Spec.PodTemplate.InitContainers,
					Affinity:                      n.Spec.PodTemplate.Affinity,
					NodeSelector:                  n.Spec.PodTemplate.NodeSelector,
					HostNetwork:                   n.Spec.PodTemplate.HostNetwork,
					TerminationGracePeriodSeconds: n.Spec.PodTemplate.TerminationGracePeriodSeconds,
					Volumes:                       n.Spec.PodTemplate.Volumes,
					Tolerations:                   n.Spec.PodTemplate.Toleration,
					Containers: append([]coreV1.Container{
						{
							Name:            n.Name,
							Image:           NewDefaultStringUtils(n.Spec.Image, defaultImage).ValueOrDefault(),
							Command:         nil,
							Resources:       n.Spec.Resources,
							SecurityContext: getSecurityContext(n),
							Ports:           getContainerPorts(n),
							VolumeMounts:    n.Spec.PodTemplate.VolumeMounts,
							ReadinessProbe:  getContainerProbes(n),
						}}, n.Spec.PodTemplate.Containers...),
				},
			},
		},
	}

	setConfigRef(n.Spec.Config, &deployment)
	return &deployment, nil
}
