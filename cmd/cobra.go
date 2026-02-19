package cmd

import (
	"errors"
	"fmt"
	"order/cmd/api"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:          "order",
	Short:        "order",
	Long:         "order-srv",
	SilenceUsage: true,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("requires at least 1 arg")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		tips()
	},
}

func init() {
	rootCmd.AddCommand(api.StartCmd)
	rootCmd.AddCommand(migrate.StartCmd)
}

func tips() {
	usageStr := `欢迎使用服务，请使用 -h 查看命令`
	fmt.Printf("%s\n", usageStr)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}
