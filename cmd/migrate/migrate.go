package migrate

import (
	"github.com/xxx-newbee/order/internal/config"
	"github.com/xxx-newbee/order/internal/model"
	"github.com/xxx-newbee/order/internal/svc"

	"github.com/spf13/cobra"
	"github.com/zeromicro/go-zero/core/conf"
)

var (
	configY  string
	StartCmd = &cobra.Command{
		Use:     "migrate",
		Short:   "Run migrations",
		Long:    "Run migrations",
		Example: "go run order.go migrate",
		RunE: func(cmd *cobra.Command, args []string) error {
			return run()
		},
	}
)

func init() {
	StartCmd.Flags().StringVarP(&configY, "config", "c", "etc/order.yaml", "config file")
}

func run() error {
	conf.MustLoad(configY, &config.C)
	db := svc.InitDB(config.C)
	err := db.AutoMigrate(model.OrderMain{}, model.OrderItem{}, model.SeckillStock{})
	return err
}
