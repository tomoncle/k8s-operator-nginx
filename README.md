# k8s-operator-nginx

使用 kubebuilder 构建一个简单的 k8s-operator, 随便搞搞.

```
CRD -> Nginx (deploy - service - ingress)
```

## 软件版本

* golang     : `1.20.1`
* kubebuilder: `3.9.0`
* kubernetes : `1.22.4`

## 使用

* 安装CRD

```bash
# 安装
$ make install
# 查看
$ kubectl get crds
NAME                            CREATED AT
nginxes.devops.github.com       2023-03-01T10:22:32Z
webapis.kube-dev.tomoncle.com   2023-02-26T06:25:30Z
```

* 运行Controller

```bash
$ make run
2023-03-02T14:01:52+08:00	INFO	controller-runtime.metrics	Metrics server is starting to listen	{"addr": ":8080"}
2023-03-02T14:01:52+08:00	INFO	setup	starting manager
2023-03-02T14:01:52+08:00	INFO	Starting server	{"path": "/metrics", "kind": "metrics", "addr": "[::]:8080"}
2023-03-02T14:01:52+08:00	INFO	Starting server	{"kind": "health probe", "addr": "[::]:8081"}
2023-03-02T14:01:52+08:00	INFO	Starting EventSource	{"controller": "nginx", "controllerGroup": "devops.github.com", "controllerKind": "Nginx", "source": "kind source: *v1.Nginx"}
2023-03-02T14:01:52+08:00	INFO	Starting Controller	{"controller": "nginx", "controllerGroup": "devops.github.com", "controllerKind": "Nginx"}
2023-03-02T14:01:53+08:00	INFO	Starting workers	{"controller": "nginx", "controllerGroup": "devops.github.com", "controllerKind": "Nginx", "worker count": 1}
2023-03-02T14:01:53+08:00	INFO	controllers.nginx	
##    ## ##     ## ########  ######## ########  ##     ## #### ##       ########  ######## ########  
##   ##  ##     ## ##     ## ##       ##     ## ##     ##  ##  ##       ##     ## ##       ##     ## 
##  ##   ##     ## ##     ## ##       ##     ## ##     ##  ##  ##       ##     ## ##       ##     ## 
#####    ##     ## ########  ######   ########  ##     ##  ##  ##       ##     ## ######   ########  
##  ##   ##     ## ##     ## ##       ##     ## ##     ##  ##  ##       ##     ## ##       ##   ##   
##   ##  ##     ## ##     ## ##       ##     ## ##     ##  ##  ##       ##     ## ##       ##    ##  
##    ##  #######  ########  ######## ########   #######  #### ######## ########  ######## ##     ##

2023-03-02T14:01:53+08:00	INFO	controllers.nginx.Reconcile	查询CRD实例: 开始	{"命名空间/名称": "default/nginx-sample"}
2023-03-02T14:01:53+08:00	INFO	controllers.nginx.Reconcile	查询CRD实例: 结束	{"命名空间/名称": "default/nginx-sample", "详情": {"apiVersion": "devops.github.com/v1", "kind": "Nginx", "namespace": "default", "name": "nginx-sample"}}
2023-03-02T14:01:53+08:00	INFO	controllers.nginx.Reconcile	判断CRD实例是否匹配 AnnotationFilter: 开始	{"命名空间/名称": "default/nginx-sample"}
2023-03-02T14:01:53+08:00	INFO	controllers.nginx.shouldManageNginx	判断CRD实例是否匹配AnnotationFilter: 执行	{"命名空间": "default"}
2023-03-02T14:01:53+08:00	INFO	controllers.nginx.Reconcile	判断CRD实例是否匹配 AnnotationFilter: 结束	{"命名空间/名称": "default/nginx-sample"}
2023-03-02T14:01:53+08:00	INFO	controllers.nginx.Reconcile	处理CRD实例: 开始	{"命名空间/名称": "default/nginx-sample"}
```

* 安装CR

```bash
# 安装
$ kubectl apply -f config/samples/

# 查看
$ kubectl get nginx
NAME           CURRENT   DESIRED   AGE
nginx-sample   1                   19h
 
$ kubectl get hpa
NAME                               REFERENCE              TARGETS   MINPODS   MAXPODS   REPLICAS   AGE
default-auto-scaled-nginx-sample   Nginx/nginx-sample     33%/45%   1         10        1          21h

$ kubectl get deploy,pod,svc,ing -l devops.github.com/app=nginx
NAME                           		READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/nginx-sample   		1/1     1            1           23s
		
NAME                           	    READY   STATUS    RESTARTS   AGE
pod/nginx-sample-6cc8c7cfd9-jcgkp   1/1     Running   0          22s

NAME                           		TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)          AGE
service/nginx-sample-service   		ClusterIP   10.68.93.224   <none>        80/TCP,443/TCP   22s

NAME                                             CLASS   HOSTS                                     ADDRESS   PORTS     AGE
ingress.networking.k8s.io/nginx-sample-ingress   nginx   dev-01.devops.com,ops-01.devops.com                 80, 443   22s
```

* 测试

```bash
$ curl -i https://dev-01.devops.com/healthz
HTTP/1.1 200 OK
Date: Thu, 02 Mar 2023 06:07:32 GMT
Content-Type: text/plain
Content-Length: 8
Connection: keep-alive
Strict-Transport-Security: max-age=15724800; includeSubDomains

WORKING
```

## License

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

