package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:   "show <name>",
	Short: "查看连接详情",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		conn, err := manager.GetConnection(name)
		if err != nil {
			return err
		}

		out := cmd.OutOrStdout()
		fmt.Fprintf(out, "名称：    %s\n", conn.Name)
		fmt.Fprintf(out, "主机：    %s\n", conn.Host)
		fmt.Fprintf(out, "端口：    %d\n", conn.Port)
		fmt.Fprintf(out, "用户：    %s\n", conn.User)
		fmt.Fprintf(out, "密码：    ********\n")
		fmt.Fprintf(out, "标签：    %s\n", strings.Join(conn.Tags, ", "))
		fmt.Fprintf(out, "备注：    %s\n", conn.Notes)
		fmt.Fprintf(out, "创建时间：%s\n", conn.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Fprintf(out, "更新时间：%s\n", conn.UpdatedAt.Format("2006-01-02 15:04:05"))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(showCmd)
}
