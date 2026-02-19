package api

import (
	"fmt"
	"github.com/xxx-newbee/order/internal/config"
	"github.com/xxx-newbee/order/internal/server"
	"github.com/xxx-newbee/order/internal/svc"
	"github.com/xxx-newbee/order/order"

	"github.com/spf13/cobra"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	configY  string
	StartCmd = &cobra.Command{
		Use:     "service",
		Short:   "start api server",
		Example: "go run order.go service -c /your/config/file.yaml",
		PreRun: func(cmd *cobra.Command, args []string) {
			setup()
		},
		Run: func(cmd *cobra.Command, args []string) {
			run()
		},
	}
)

func init() {
	StartCmd.Flags().StringVarP(&configY, "config", "c", "etc/order.yaml", "the config file")
}

func setup() {
	conf.MustLoad(configY, &config.C)
}

func run() {
	ctx := svc.NewServiceContext(config.C)

	s := zrpc.MustNewServer(config.C.RpcServerConf, func(grpcServer *grpc.Server) {
		order.RegisterOrderServer(grpcServer, server.NewOrderServer(ctx))

		if config.C.Mode == service.DevMode || config.C.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", config.C.ListenOn)
	s.Start()
}
