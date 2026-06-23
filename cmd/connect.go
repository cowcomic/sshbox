package cmd

import (
	"fmt"
	"sshbox/internal/ssh"

	"github.com/spf13/cobra"
)

var connectCmd = &cobra.Command{
	Use:     "connect <name>",
	Short:   "连接SSH服务器 (c)",
	Aliases: []string{"c"},
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		conn, err := manager.GetConnection(name)
		if err != nil {
			return err
		}

		password, err := manager.GetDecryptedPassword(name)
		if err != nil {
			return fmt.Errorf("解密密码失败: %w", err)
		}

		fmt.Printf("正在连接 %s (%s:%d)...\n", name, conn.Host, conn.Port)
		return ssh.Connect(name, conn.Host, conn.Port, conn.User, password)
	},
}

func init() {
	rootCmd.AddCommand(connectCmd)
}
