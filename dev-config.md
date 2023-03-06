# 开发环境配置

## 软件版本

* golang     : `1.20.1`
* kubebuilder: `3.9.0`
* kubernetes : `1.22.4`

## 初始化

* 创建项目，并初始化

```bash
$ mkdir -p github.com/tomoncle/k8s-operator-nginx
$ cd github.com/tomoncle/k8s-operator-nginx
# init repo
$ kubebuilder init --domain github.com --repo github.com/tomoncle/k8s-operator-nginx
# init api
$ kubebuilder create api --group devops --version v1 --kind Nginx
```

* 修改 go.mod 配置文件，替换 k8s.io/api 版本

```
// 修复 kubebuilder 3.9 和kubernetes 1.22.4 版本结合报错问题
replace k8s.io/api v0.26.0 => k8s.io/api v0.25.0
```

## CRD

* 生成deepcopy文件

```bash
$ make generate
```

* 生成crd、rbac、prometheus等配置文件

```bash
$ make manifests
```

* 安装

```bash
$ make install
```

* 运行controller

```bash
$ make run
```

* 卸载

```bash
$ make uninstall
```

## CR

编写：`config/samples/devops_v1_nginx.yaml`

* 安装: `$ kubectl apply -f config/samples/`
* 卸载: `$ kubectl delete -f config/samples/`

## 其他

* https://book.kubebuilder.io/reference/markers.html
```
//+kubebuilder:validation:Required 
// 用于标记该字段是必填项。

//+optional
// 用于标记该字段是可选项
```
