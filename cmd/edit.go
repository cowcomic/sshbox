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

		var newName, host, user, password *string
		var port *int
		var tags *[]string
		var notes *string

		if cmd.Flags().Changed("name") {
			v, _ := cmd.Flags().GetString("name")
			newName = &v
		}
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

		if err := manager.UpdateConnection(name, newName, host, port, user, password, tags, notes); err != nil {
			return err
		}

		if newName != nil {
			cmd.Printf("✓ 连接 %s 已重命名为 %s\n", name, *newName)
		} else {
			cmd.Printf("✓ 连接 %s 已更新\n", name)
		}
		return nil
	},
}

func init() {
	editCmd.Flags().StringP("name", "N", "", "连接名称（重命名）")
	editCmd.Flags().StringP("host", "H", "", "主机地址")
	editCmd.Flags().IntP("port", "P", 22, "端口号")
	editCmd.Flags().StringP("user", "u", "", "登录用户名")
	editCmd.Flags().StringP("password", "p", "", "登录密码")
	editCmd.Flags().StringP("tag", "t", "", "标签（替换现有标签）")
	editCmd.Flags().StringP("notes", "n", "", "备注信息")
	rootCmd.AddCommand(editCmd)
}
