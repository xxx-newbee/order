package main

import (
	"flag"
	"fmt"

	"github.com/xxx-newbee/order/internal/config"
	"github.com/xxx-newbee/order/internal/server"
	"github.com/xxx-newbee/order/internal/svc"
	"github.com/xxx-newbee/order/order"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/order.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

}
