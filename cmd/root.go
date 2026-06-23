package cmd

import (
	"fmt"
	"os"
	"sshbox/internal/keyring"
	"sshbox/internal/service"
	"sshbox/internal/storage"

	"github.com/spf13/cobra"
)

var (
	manager *service.ConnectionManager
	rootCmd = &cobra.Command{
		Use:   "sshbox",
		Short: "轻量级SSH连接管理工具",
		Long:  "sshbox - 基于命令行的轻量级SSH连接管理工具，支持Tag分组和快速连接。",
	}
)

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initManager)
}

func initManager() {
	store, err := storage.NewConfigStore()
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}

	masterKey, err := keyring.GetOrCreate()
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 获取主密钥失败: %v\n", err)
		os.Exit(1)
	}

	manager = service.NewConnectionManager(store, masterKey)
}
