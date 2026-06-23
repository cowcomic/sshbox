package cmd

import (
	"strings"

	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit <name>",
	Short: "编辑SSH连接",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		var host *string
		var port *int
		var user, password *string
		var tags *[]string
		var notes *string

		if cmd.Flags().Changed("host") {
			v, _ := cmd.Flags().GetString("host")
			host = &v
		}
		if cmd.Flags().Changed("port") {
			v, _ := cmd.Flags().GetInt("port")
			port = &v
		}
		if cmd.Flags().Changed("user") {
			v, _ := cmd.Flags().GetString("user")
			user = &v
		}
		if cmd.Flags().Changed("password") {
			v, _ := cmd.Flags().GetString("password")
			password = &v
		}
		if cmd.Flags().Changed("tag") {
			v, _ := cmd.Flags().GetString("tag")
			parsed := strings.Split(v, ",")
			tags = &parsed
		}
		if cmd.Flags().Changed("notes") {
			v, _ := cmd.Flags().GetString("notes")
			notes = &v
		}

		if err := manager.UpdateConnection(name, host, port, user, password, tags, notes); err != nil {
			return err
		}

		cmd.Printf("✓ 连接 %s 已更新\n", name)
		return nil
	},
}

func init() {
	editCmd.Flags().String("host", "", "主机地址")
	editCmd.Flags().Int("port", 22, "端口号")
	editCmd.Flags().String("user", "", "登录用户名")
	editCmd.Flags().String("password", "", "登录密码")
	editCmd.Flags().String("tag", "", "标签（替换现有标签）")
	editCmd.Flags().String("notes", "", "备注信息")
	rootCmd.AddCommand(editCmd)
}
