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

package controllers

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	devopsV1 "github.com/tomoncle/k8s-operator-nginx/api/v1"
	"github.com/tomoncle/k8s-operator-nginx/pkg/k8s"
	appsV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	networkingV1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sort"
	"strings"
	"time"
)

const KubeBuilder = `
##    ## ##     ## ########  ######## ########  ##     ## #### ##       ########  ######## ########  
##   ##  ##     ## ##     ## ##       ##     ## ##     ##  ##  ##       ##     ## ##       ##     ## 
##  ##   ##     ## ##     ## ##       ##     ## ##     ##  ##  ##       ##     ## ##       ##     ## 
#####    ##     ## ########  ######   ########  ##     ##  ##  ##       ##     ## ######   ########  
##  ##   ##     ## ##     ## ##       ##     ## ##     ##  ##  ##       ##     ## ##       ##   ##   
##   ##  ##     ## ##     ## ##       ##     ## ##     ##  ##  ##       ##     ## ##       ##    ##  
##    ##  #######  ########  ######## ########   #######  #### ######## ########  ######## ##     ##
`

// 控制器最终会在群集上运行，因此 需要 RBAC 权限，使用控制器工具 RBAC 标记指定这些权限。
// 这些是运行所需的最低权限。有需要自己再添加。
// https://book.kubebuilder.io/reference/markers/rbac.html

// +kubebuilder:rbac:groups=devops.github.com,resources=nginxes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=devops.github.com,resources=nginxes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=devops.github.com,resources=nginxes/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch
// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch
// +kubebuilder:rbac:groups="",resources=events,verbs=create;update;patch

// NginxReconciler reconciles a Nginx object
type NginxReconciler struct {
	client.Client
	EventRecorder    record.EventRecorder
	Log              logr.Logger
	Scheme           *runtime.Scheme
	AnnotationFilter labels.Selector
}

func (r *NginxReconciler) listDeployments(ctx context.Context, obj *devopsV1.Nginx) ([]appsV1.Deployment, error) {
	logger := r.Log.WithName("listDeployments").WithValues("命名空间", obj.Namespace)
	logger.Info("查询 Nginx Deployments 列表(根据label筛选)")

	var deployList appsV1.DeploymentList
	err := r.Client.List(ctx, &deployList, &client.ListOptions{
		Namespace:     obj.Namespace,
		LabelSelector: labels.SelectorFromSet(k8s.LabelsForNginx(obj.Name)),
	})
	if err != nil {
		logger.Error(err, "查询 Nginx Deployments 列表：失败")
		return nil, err
	}

	deploys := deployList.Items
	for _, i := range deploys {
		logger.Info("查询 Nginx Deployment", "详情", i.Name, "状态", i.Status)
	}
	// 如果根据标签查询不到，就不根据标签查了
	if len(deploys) == 0 {
		err = r.Client.List(ctx, &deployList, &client.ListOptions{
			Namespace: obj.Namespace,
		})
		if err != nil {
			return nil, err
		}

		cr := *metaV1.NewControllerRef(obj, schema.GroupVersionKind{
			Group:   devopsV1.GroupVersion.Group,
			Version: devopsV1.GroupVersion.Version,
			Kind:    devopsV1.Kind,
		})

		for _, deploy := range deployList.Items {
			for _, owner := range deploy.OwnerReferences {
				if reflect.DeepEqual(owner, cr) {
					deploys = append(deploys, deploy)
				}
			}
		}
	}

	sort.Slice(deploys, func(i, j int) bool {
		return deploys[i].Name < deploys[j].Name
	})
	return deploys, nil
}

// listServices return all the services for the given nginx sorted by name
func (r *NginxReconciler) listServices(ctx context.Context, obj *devopsV1.Nginx) ([]devopsV1.ServiceStatus, error) {
	logger := r.Log.WithName("listServices").WithValues("命名空间", obj.Namespace)
	serviceList := &coreV1.ServiceList{}
	labelSelector := labels.SelectorFromSet(k8s.LabelsForNginx(obj.Name))
	listOps := &client.ListOptions{Namespace: obj.Namespace, LabelSelector: labelSelector}
	err := r.Client.List(ctx, serviceList, listOps)
	if err != nil {
		logger.Error(err, "查询 Nginx Services list 失败")
		return nil, err
	}

	var services []devopsV1.ServiceStatus
	for _, s := range serviceList.Items {
		logger.Info("查询 Nginx Service", "详情", s.Name, "状态", s.Status)
		services = append(services, devopsV1.ServiceStatus{
			Name: s.Name,
		})
	}

	sort.Slice(services, func(i, j int) bool {
		return services[i].Name < services[j].Name
	})

	return services, nil
}

func (r *NginxReconciler) listIngresses(ctx context.Context, obj *devopsV1.Nginx) ([]devopsV1.IngressStatus, error) {
	logger := r.Log.WithName("listIngresses").WithValues("命名空间", obj.Namespace)
	var ingressList networkingV1.IngressList

	options := &client.ListOptions{
		LabelSelector: labels.SelectorFromSet(k8s.LabelsForNginx(obj.Name)),
		Namespace:     obj.Namespace,
	}
	if err := r.Client.List(ctx, &ingressList, options); err != nil {
		logger.Error(err, "查询 Nginx Ingress list 失败")
		return nil, err
	}

	var ingresses []devopsV1.IngressStatus
	for _, i := range ingressList.Items {
		logger.Info("查询 Nginx Ingress", "详情", i.Name, "状态", i.Status)
		ingresses = append(ingresses, devopsV1.IngressStatus{Name: i.Name})
	}

	sort.Slice(ingresses, func(i, j int) bool {
		return ingresses[i].Name < ingresses[j].Name
	})

	return ingresses, nil
}

func (r *NginxReconciler) listPods(ctx context.Context, obj *devopsV1.Nginx) error {
	logger := r.Log.WithName("listPods").WithValues("命名空间", obj.Namespace)
	var podList coreV1.PodList
	err := r.Client.List(ctx, &podList, &client.ListOptions{
		Namespace:     obj.Namespace,
		LabelSelector: labels.SelectorFromSet(k8s.LabelsForNginx(obj.Name)),
	})
	if err != nil {
		logger.Error(err, "查询失败.")
		return err
	} else {
		logger.Info("查询 Nginx POD 列表：成功", "数量", len(podList.Items))
		for _, pod := range podList.Items {
			logger.Info("查询 Nginx POD 列表：成功", "详情", pod.ObjectMeta.Name, "状态", pod.Status.Phase)
		}
	}
	return nil
}

func (r *NginxReconciler) listNginx(ctx context.Context) {
	logger := r.Log.WithName("listNginx")
	var nginxList devopsV1.NginxList
	err := r.Client.List(ctx, &nginxList, &client.ListOptions{})
	if err != nil {
		logger.Error(err, "查询失败.")
	} else {
		logger.Info("查询 Nginx 列表：成功", "count", len(nginxList.Items))
		for _, nginx := range nginxList.Items {
			logger.Info("查询 Nginx 列表：成功", "nginx", nginx.Name)
		}
	}
}

func (r *NginxReconciler) refreshStatus(ctx context.Context, obj *devopsV1.Nginx) error {
	logger := r.Log.WithName("refreshStatus").WithValues("命名空间", obj.Namespace)

	logger.Info("查询 Deployment 列表")
	deploys, err := r.listDeployments(ctx, obj)
	if err != nil {
		return err
	}

	var deployStatuses []devopsV1.DeploymentStatus
	var replicas int32
	for _, d := range deploys {
		replicas += d.Status.Replicas
		deployStatuses = append(deployStatuses, devopsV1.DeploymentStatus{Name: d.Name})
	}

	logger.Info("查询 Service 列表")
	services, err := r.listServices(ctx, obj)
	if err != nil {
		return fmt.Errorf("failed to list services for nginx: %v", err)
	}

	logger.Info("查询 Ingress 列表")
	ingresses, err := r.listIngresses(ctx, obj)
	if err != nil {
		return fmt.Errorf("failed to list ingresses for nginx: %w", err)
	}

	sort.Slice(obj.Status.Services, func(i, j int) bool {
		return obj.Status.Services[i].Name < obj.Status.Services[j].Name
	})

	sort.Slice(obj.Status.Ingresses, func(i, j int) bool {
		return obj.Status.Ingresses[i].Name < obj.Status.Ingresses[j].Name
	})

	status := devopsV1.NginxStatus{
		CurrentReplicas: replicas,
		PodSelector:     labels.FormatLabels(k8s.LabelsForNginx(obj.Name)),
		Deployments:     deployStatuses,
		Services:        services,
		Ingresses:       ingresses,
	}

	if reflect.DeepEqual(obj.Status, status) {
		logger.Info("未检测到资源变化")
		return nil
	}

	logger.Info("更新资源状态")
	obj.Status = status
	err = r.Client.Status().Update(ctx, obj)
	if err != nil {
		return fmt.Errorf("failed to update nginx status: %v", err)
	}

	return nil
}

func (r *NginxReconciler) shouldManageNginx(obj *devopsV1.Nginx) bool {
	logger := r.Log.WithName("shouldManageNginx").WithValues("命名空间", obj.Namespace)
	logger.Info("判断CRD实例是否匹配AnnotationFilter: 执行")
	// empty filter matches all resources
	if r.AnnotationFilter == nil || r.AnnotationFilter.Empty() {
		return true
	}
	return r.AnnotationFilter.Matches(labels.Set(obj.Annotations))
}

func (r *NginxReconciler) reconcileNginx(ctx context.Context, obj *devopsV1.Nginx) error {
	logger := r.Log.WithName("reconcileNginx").WithValues("命名空间", obj.Namespace)
	logger.Info("处理CRD实例: 执行 -> step1. 处理 Deployment")
	if err := r.reconcileDeployment(ctx, obj); err != nil {
		return err
	}
	logger.Info("处理CRD实例: 执行 -> step2. 处理 Service")
	if err := r.reconcileService(ctx, obj); err != nil {
		return err
	}
	logger.Info("处理CRD实例: 执行 -> step3. 处理 Ingress")
	if err := r.reconcileIngress(ctx, obj); err != nil {
		return err
	}
	logger.Info("处理CRD实例: 结束")
	return nil
}

func (r *NginxReconciler) reconcileDeployment(ctx context.Context, obj *devopsV1.Nginx) error {
	logger := r.Log.WithName("reconcileDeployment").WithValues("命名空间", obj.Namespace)

	newDeploy, err := k8s.NewDeployment(obj)
	if err != nil {
		logger.Error(err, "构建 Nginx Deployment 失败: ")
		return fmt.Errorf("构建 Nginx Deployment 失败: %w", err)
	}

	logger.Info("查询 Nginx Deployment 实例: 开始")
	var currentDeploy appsV1.Deployment
	err = r.Client.Get(ctx, types.NamespacedName{Name: newDeploy.Name, Namespace: newDeploy.Namespace}, &currentDeploy)
	if errors.IsNotFound(err) {
		logger.Info("查询 Nginx Deployment 实例: 不存在")
		logger.Info("新建 Nginx Deployment 实例：开始")
		return r.Client.Create(ctx, newDeploy)
	}

	if err != nil {
		logger.Error(err, "查询 Nginx Deployment 实例: 失败")
		return fmt.Errorf("不能获取 Deployment: %w", err)
	}

	//logger.Info("查询 Nginx Deployment 实例: 成功","详情", currentDeploy)
	// 查询一下pod信息
	err = r.listPods(ctx, obj)
	if err != nil {
		logger.Error(err, "查询 Nginx Pod 列表: 失败")
	}

	replicas := currentDeploy.Spec.Replicas

	patch := client.StrategicMergeFrom(currentDeploy.DeepCopy())
	currentDeploy.Spec = newDeploy.Spec

	if newDeploy.Spec.Replicas == nil {
		if replicas == nil {
			defaultReplicas := int32(1)
			currentDeploy.Spec.Replicas = &defaultReplicas
		} else {
			currentDeploy.Spec.Replicas = replicas
		}
	}

	err = r.Client.Patch(ctx, &currentDeploy, patch)
	if err != nil {
		logger.Error(err, "Patch Nginx deployment: 失败")
		return fmt.Errorf("failed to patch Deployment: %w", err)
	}

	return nil
}

func (r *NginxReconciler) reconcileService(ctx context.Context, obj *devopsV1.Nginx) error {
	logger := r.Log.WithName("reconcileService").WithValues("命名空间", obj.Namespace)

	newService := k8s.NewService(obj)

	var currentService coreV1.Service
	var namespace = types.NamespacedName{Name: newService.Name, Namespace: newService.Namespace}

	logger.Info("查询 Nginx Service 实例: 开始")
	err := r.Client.Get(ctx, namespace, &currentService)

	if errors.IsNotFound(err) {
		logger.Info("查询 Nginx Service 实例: 不存在")
		logger.Info("新建 Nginx Service 实例：开始")

		err = r.Client.Create(ctx, newService)
		if errors.IsForbidden(err) && strings.Contains(err.Error(), "exceeded quota") {
			logger.Error(err, "新建 Nginx Service 实例：失败")
			r.EventRecorder.Eventf(obj, coreV1.EventTypeWarning, "ServiceQuotaExceeded", "创建服务失败: %s", err)
			return err
		}

		if err != nil {
			logger.Error(err, "新建 Nginx Service 实例：失败")
			r.EventRecorder.Eventf(obj, coreV1.EventTypeWarning, "ServiceCreationFailed", "创建服务失败: %s", err)
			return err
		}
		logger.Info("新建 Nginx Service 实例：成功")
		r.EventRecorder.Eventf(obj, coreV1.EventTypeNormal, "ServiceCreated", "创建服务成功")
		return nil
	}

	if err != nil {
		logger.Error(err, "查询 Nginx Service 实例: 失败")
		return fmt.Errorf("查询Service服务失败: %v", err)
	}
	//logger.Info("查询 Nginx Service 实例: 结束","详情", currentService)

	newService.ResourceVersion = currentService.ResourceVersion
	newService.Spec.ClusterIP = currentService.Spec.ClusterIP
	newService.Spec.HealthCheckNodePort = currentService.Spec.HealthCheckNodePort
	newService.Finalizers = currentService.Finalizers

	for annotation, value := range currentService.Annotations {
		if newService.Annotations[annotation] == "" {
			newService.Annotations[annotation] = value
		}
	}

	if newService.Spec.Type == coreV1.ServiceTypeNodePort || newService.Spec.Type == coreV1.ServiceTypeLoadBalancer {
		// avoid node port reallocation preserving the current ones
		for _, currentPort := range currentService.Spec.Ports {
			for index, newPort := range newService.Spec.Ports {
				if currentPort.Port == newPort.Port {
					newService.Spec.Ports[index].NodePort = currentPort.NodePort
				}
			}
		}
	}

	err = r.Client.Update(ctx, newService)
	if err != nil {
		r.EventRecorder.Eventf(obj, coreV1.EventTypeWarning, "ServiceUpdateFailed", "更新服务失败: %s", err)
		return err
	}

	r.EventRecorder.Eventf(obj, coreV1.EventTypeNormal, "ServiceUpdated", "更新服务成功")
	return nil
}

func shouldUpdateIngress(currentIngress, newIngress *networkingV1.Ingress) bool {
	if currentIngress == nil || newIngress == nil {
		return false
	}
	return !reflect.DeepEqual(currentIngress.Annotations, newIngress.Annotations) ||
		!reflect.DeepEqual(currentIngress.Labels, newIngress.Labels) ||
		!reflect.DeepEqual(currentIngress.Spec, newIngress.Spec)
}

func (r *NginxReconciler) reconcileIngress(ctx context.Context, obj *devopsV1.Nginx) error {
	logger := r.Log.WithName("reconcileIngress").WithValues("命名空间", obj.Namespace)

	if obj == nil {
		return fmt.Errorf("nginx cannot be nil")
	}

	logger.Info("查询 Nginx Ingress 实例: 开始")
	newIngress := k8s.NewIngress(obj)
	var currentIngress networkingV1.Ingress

	err := r.Client.Get(ctx, types.NamespacedName{Name: newIngress.Name, Namespace: newIngress.Namespace}, &currentIngress)
	if errors.IsNotFound(err) {
		logger.Info("查询 Nginx Ingress 实例: 不存在")
		if obj.Spec.Ingress == nil {
			logger.Info("CRD实例YAML配置文件未配置Ingress: 忽略Ingress的操作")
			return nil
		}

		logger.Info("创建 Nginx Ingress 实例")
		return r.Client.Create(ctx, newIngress)
	}

	if err != nil {
		logger.Error(err, "查询 Nginx Ingress 实例: 失败")
		return err
	}
	//logger.Info("查询 Nginx Ingress 实例: 结束","详情",currentIngress)

	logger.Info("查询 Nginx CRD 实例，是否配置了 Ingress")
	if obj.Spec.Ingress == nil {
		logger.Info("CRD实例YAML配置文件未配置Ingress: 删除多余的Ingress")
		return r.Client.Delete(ctx, &currentIngress)
	}

	logger.Info("验证 Nginx CRD 实例，是否更新了 Ingress")
	if !shouldUpdateIngress(&currentIngress, newIngress) {
		return nil
	}

	logger.Info("更新 Nginx Ingress")
	newIngress.ResourceVersion = currentIngress.ResourceVersion
	newIngress.Finalizers = currentIngress.Finalizers
	return r.Client.Update(ctx, newIngress)
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Nginx object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *NginxReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	//logger := log.FromContext(ctx)
	r.Log.Info(KubeBuilder)
	r.listNginx(ctx)
	logger := r.Log.WithName("Reconcile").WithValues("命名空间/名称", req.NamespacedName)

	var instance devopsV1.Nginx
	logger.Info("查询CRD实例: 开始")
	err := r.Client.Get(ctx, req.NamespacedName, &instance)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Error(err, "查询CRD实例：失败")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "获取不到Nginx资源")
		return ctrl.Result{}, err
	}
	logger.Info("查询CRD实例: 结束", "详情", &instance)

	logger.Info("判断CRD实例是否匹配 AnnotationFilter: 开始")
	if !r.shouldManageNginx(&instance) {
		logger.Error(err, "判断CRD实例是否匹配AnnotationFilter: 失败，终止继续执行")
		return ctrl.Result{Requeue: true, RequeueAfter: 5 * time.Minute}, nil
	}
	logger.Info("判断CRD实例是否匹配 AnnotationFilter: 结束")

	logger.Info("处理CRD实例: 开始")
	if err := r.reconcileNginx(ctx, &instance); err != nil {
		logger.Error(err, "处理CRD实例: 失败")
		return ctrl.Result{}, err
	}
	logger.Info("处理CRD实例: 结束")

	logger.Info("刷新CRD实例状态：开始")
	if err := r.refreshStatus(ctx, &instance); err != nil {
		logger.Error(err, "刷新CRD实例状态: 失败")
		return ctrl.Result{}, err
	}
	logger.Info("刷新CRD实例状态：结束")

	logger.Info("处理CRD实例: 结束", "详情", &instance)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
// owns的一般使用 将 deployment service ingress或者其他资源作为operator应用的子资源，进行生命周期管理
// 既删除 crd 实例 nginx 时，对应的 deployment service ingress 资源也会删除.
func (r *NginxReconciler) SetupWithManager(mgr ctrl.Manager) error {
	//return ctrl.NewControllerManagedBy(mgr).
	//	For(&devopsV1.Nginx{}).
	//	Owns(&appsV1.Deployment{}).
	//	Complete(r)
	return ctrl.NewControllerManagedBy(mgr).
		For(&devopsV1.Nginx{}).
		Owns(&appsV1.Deployment{}).
		Owns(&coreV1.Service{}).
		Owns(&networkingV1.Ingress{}).
		Complete(r)
}
