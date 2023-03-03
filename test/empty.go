package main

import (
	"fmt"
	devopsV1 "github.com/tomoncle/k8s-operator-nginx/api/v1"
)

func main() {
	fmt.Println(devopsV1.GroupVersion.Group)
}
