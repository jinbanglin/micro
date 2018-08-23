package main

import (
	"github.com/jinbanglin/micro/cmd"
	// etcdv3 service discover
	_ "github.com/jinbanglin/go-plugins/registry/etcdv3"
	// tcp transport
	_ "github.com/jinbanglin/go-plugins/transport/tcp"
)

func main() {
	cmd.Init()
}
