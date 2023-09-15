package main

import (
	"context"
	"fmt"
	"github.com/azure/spin-aks-plugin/pkg/azure"
)

var version string // set through ldflags at build time like "go run -ldflags "-X main.version=0.0.1" main.go"

func main() {
	err := azure.LinkAcr(context.Background(), "26ad903f-2330-429d-8389-864ac35c4350", "jkatariyatest", "jkatariyatest5", "jkatariyatest", "jkatariyatest")
	fmt.Println(err)
}
