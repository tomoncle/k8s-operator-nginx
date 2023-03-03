package main

import (
	"fmt"
	devopsv1 "github.com/tomoncle/k8s-operator-nginx/api/v1"
)

func main() {
	fmt.Println(devopsv1.GroupVersion.Group)
}
