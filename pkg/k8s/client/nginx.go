package client

import (
	"context"
	"encoding/json"
	"fmt"
	devopsV1 "github.com/tomoncle/k8s-operator-nginx/api/v1"
	"time"

	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	coreV1 "k8s.io/client-go/applyconfigurations/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

// NginxGetter has a method to return a NginxInterface.
// A group's client should implement this interface.
type NginxGetter interface {
	Nginxes(namespace string) NginxInterface
}

func (c *NginxClient) Nginxes(namespace string) NginxInterface {
	return newNginxes(c, namespace)
}

// NginxClient implements NginxGetter
type NginxClient struct {
	restClient rest.Interface
}

type NginxInterface interface {
	Create(ctx context.Context, nginx *devopsV1.Nginx, opts metaV1.CreateOptions) (*devopsV1.Nginx, error)
	Update(ctx context.Context, nginx *devopsV1.Nginx, opts metaV1.UpdateOptions) (*devopsV1.Nginx, error)
	UpdateStatus(ctx context.Context, nginx *devopsV1.Nginx, opts metaV1.UpdateOptions) (*devopsV1.Nginx, error)
	Delete(ctx context.Context, name string, opts metaV1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts metaV1.DeleteOptions, listOpts metaV1.ListOptions) error
	Get(ctx context.Context, name string, opts metaV1.GetOptions) (*devopsV1.Nginx, error)
	List(ctx context.Context, opts metaV1.ListOptions) (*devopsV1.NginxList, error)
	Watch(ctx context.Context, opts metaV1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metaV1.PatchOptions, subResources ...string) (result *devopsV1.Nginx, err error)
}

// nginxes implements NginxInterface
type nginxes struct {
	client rest.Interface
	ns     string
}

func newNginxes(c *NginxClient, namespace string) *nginxes {
	return &nginxes{
		client: c.restClient,
		ns:     namespace,
	}
}

// NewForConfig return NginxClient
func NewForConfig(c *rest.Config) (*NginxClient, error) {
	config := *c
	config.ContentConfig.GroupVersion = &devopsV1.GroupVersion
	config.APIPath = "/apis"
	config.NegotiatedSerializer = scheme.Codecs
	config.UserAgent = rest.DefaultKubernetesUserAgent()
	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	return &NginxClient{restClient: client}, nil
}

func (c *nginxes) Get(ctx context.Context, name string, options metaV1.GetOptions) (result *devopsV1.Nginx, err error) {
	result = &devopsV1.Nginx{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("nginxes").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

func (c *nginxes) List(ctx context.Context, opts metaV1.ListOptions) (result *devopsV1.NginxList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &devopsV1.NginxList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("nginxes").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

func (c *nginxes) Watch(ctx context.Context, opts metaV1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("nginxes").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

func (c *nginxes) Create(ctx context.Context, nginx *devopsV1.Nginx, opts metaV1.CreateOptions) (result *devopsV1.Nginx, err error) {
	result = &devopsV1.Nginx{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("nginxes").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(nginx).
		Do(ctx).
		Into(result)
	return
}

func (c *nginxes) Update(ctx context.Context, nginx *devopsV1.Nginx, opts metaV1.UpdateOptions) (result *devopsV1.Nginx, err error) {
	result = &devopsV1.Nginx{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("nginxes").
		Name(nginx.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(nginx).
		Do(ctx).
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *nginxes) UpdateStatus(ctx context.Context, nginx *devopsV1.Nginx, opts metaV1.UpdateOptions) (result *devopsV1.Nginx, err error) {
	result = &devopsV1.Nginx{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("nginxes").
		Name(nginx.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(nginx).
		Do(ctx).
		Into(result)
	return
}

func (c *nginxes) Delete(ctx context.Context, name string, opts metaV1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("nginxes").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

func (c *nginxes) DeleteCollection(ctx context.Context, opts metaV1.DeleteOptions, listOpts metaV1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("nginxes").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

func (c *nginxes) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metaV1.PatchOptions, subResources ...string) (result *devopsV1.Nginx, err error) {
	result = &devopsV1.Nginx{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("nginxes").
		Name(name).
		SubResource(subResources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}

func (c *nginxes) Apply(ctx context.Context, pod *coreV1.PodApplyConfiguration, opts metaV1.ApplyOptions) (result *devopsV1.Nginx, err error) {
	if pod == nil {
		return nil, fmt.Errorf("pod provided to Apply must not be nil")
	}
	patchOpts := opts.ToPatchOptions()
	data, err := json.Marshal(pod)
	if err != nil {
		return nil, err
	}
	name := pod.Name
	if name == nil {
		return nil, fmt.Errorf("pod.Name must be provided to Apply")
	}
	result = &devopsV1.Nginx{}
	err = c.client.Patch(types.ApplyPatchType).
		Namespace(c.ns).
		Resource("nginxes").
		Name(*name).
		VersionedParams(&patchOpts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}

// ApplyStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating ApplyStatus().
func (c *nginxes) ApplyStatus(ctx context.Context, pod *coreV1.PodApplyConfiguration, opts metaV1.ApplyOptions) (result *devopsV1.Nginx, err error) {
	if pod == nil {
		return nil, fmt.Errorf("pod provided to Apply must not be nil")
	}
	patchOpts := opts.ToPatchOptions()
	data, err := json.Marshal(pod)
	if err != nil {
		return nil, err
	}

	name := pod.Name
	if name == nil {
		return nil, fmt.Errorf("pod.Name must be provided to Apply")
	}

	result = &devopsV1.Nginx{}
	err = c.client.Patch(types.ApplyPatchType).
		Namespace(c.ns).
		Resource("nginxes").
		Name(*name).
		SubResource("status").
		VersionedParams(&patchOpts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}

func (c *nginxes) UpdateEphemeralContainers(ctx context.Context, podName string, nginx *devopsV1.Nginx, opts metaV1.UpdateOptions) (result *devopsV1.Nginx, err error) {
	result = &devopsV1.Nginx{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("nginxes").
		Name(podName).
		SubResource("ephemeralcontainers").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(nginx).
		Do(ctx).
		Into(result)
	return
}
