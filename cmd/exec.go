package cmd

import (
	"fmt"
	"sshbox/internal/ssh"

	"github.com/spf13/cobra"
)

var execCmd = &cobra.Command{
	Use:   "exec <name> <command>",
	Short: "在远程服务器执行命令并返回结果",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		command := args[1]

		conn, err := manager.GetConnection(name)
		if err != nil {
			return err
		}

		password, err := manager.GetDecryptedPassword(name)
		if err != nil {
			return fmt.Errorf("解密密码失败: %w", err)
		}

		output, err := ssh.Exec(conn.Host, conn.Port, conn.User, password, command)
		if len(output) > 0 {
			cmd.Print(string(output))
		}
		return err
	},
}

func init() {
	rootCmd.AddCommand(execCmd)
}
