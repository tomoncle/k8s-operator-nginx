package client

import (
	"context"
	devopsV1 "github.com/tomoncle/k8s-operator-nginx/api/v1"
	devopsClientV1 "github.com/tomoncle/k8s-operator-nginx/pkg/k8s/client"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"testing"
)

func TestGetNginx(t *testing.T) {
	utilruntime.Must(devopsV1.AddToScheme(scheme.Scheme))
	nginxClient, err := devopsClientV1.NewForConfig(ctrl.GetConfigOrDie())
	if err != nil {
		t.Error(err)
	}

	nginx, err := nginxClient.Nginxes("default").Get(context.TODO(), "nginx-sample", metaV1.GetOptions{})
	if err != nil {
		t.Error(err)
	}
	t.Log(nginx.Spec.Image)
}

func TestPutNginx(t *testing.T) {
	utilruntime.Must(devopsV1.AddToScheme(scheme.Scheme))
	nginxClient, err := devopsClientV1.NewForConfig(ctrl.GetConfigOrDie())
	if err != nil {
		t.Error(err)
	}

	nginx, err := nginxClient.Nginxes("default").Get(context.TODO(), "nginx-sample", metaV1.GetOptions{})
	if err != nil {
		t.Error(err)
	}

	nginx.Spec.Image = "nginx:latest"
	nginx, err = nginxClient.Nginxes("default").Update(context.TODO(), nginx, metaV1.UpdateOptions{})
	if err != nil {
		t.Error(err)
	}
	t.Log(nginx.Spec.Image)
}
