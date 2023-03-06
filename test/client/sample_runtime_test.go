package client

import (
	"context"
	devopsV1 "github.com/tomoncle/k8s-operator-nginx/api/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
)

func TestListNginxes(t *testing.T) {
	utilruntime.Must(devopsV1.AddToScheme(scheme.Scheme))
	nonNamespacedClient, _ := client.New(ctrl.GetConfigOrDie(), client.Options{Scheme: scheme.Scheme})
	runtimeClient := client.NewNamespacedClient(nonNamespacedClient, "default")

	var nginxList devopsV1.NginxList
	err := runtimeClient.List(context.TODO(), &nginxList, &client.ListOptions{})
	if err != nil {
		t.Error(err, "查询失败.")
	} else {
		t.Log("查询 Nginx 列表：成功", "count", len(nginxList.Items))
		for _, nginx := range nginxList.Items {
			t.Log("查询 Nginx 列表：成功", "nginx", nginx.Name)
		}
	}

}
