package main

import (
	"github.com/jinbanglin/micro/cmd"
	// tcp transport
	_ "github.com/jinbanglin/go-plugins/transport/tcp"
)

func main() {
	cmd.Init()
}
