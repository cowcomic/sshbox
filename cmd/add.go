package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "添加SSH连接",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		host, _ := cmd.Flags().GetString("host")
		port, _ := cmd.Flags().GetInt("port")
		user, _ := cmd.Flags().GetString("user")
		password, _ := cmd.Flags().GetString("password")
		tagStr, _ := cmd.Flags().GetString("tag")
		notes, _ := cmd.Flags().GetString("notes")

		if host == "" {
			return fmt.Errorf("缺少必填参数 --host/-H")
		}
		if user == "" {
			return fmt.Errorf("缺少必填参数 --user/-u")
		}
		if password == "" {
			return fmt.Errorf("缺少必填参数 --password/-p")
		}

		var tags []string
		if tagStr != "" {
			for _, t := range strings.Split(tagStr, ",") {
				t = strings.TrimSpace(t)
				if t != "" {
					tags = append(tags, t)
				}
			}
		}

		if err := manager.AddConnection(name, host, port, user, password, tags, notes); err != nil {
			return err
		}

		cmd.Printf("✓ 连接 %s 已添加\n", name)
		return nil
	},
}

func init() {
	addCmd.Flags().StringP("host", "H", "", "主机地址（必填）")
	addCmd.Flags().IntP("port", "P", 22, "端口号")
	addCmd.Flags().StringP("user", "u", "", "登录用户名（必填）")
	addCmd.Flags().StringP("password", "p", "", "登录密码（必填）")
	addCmd.Flags().StringP("tag", "t", "", "标签，多个用逗号分隔")
	addCmd.Flags().StringP("notes", "n", "", "备注信息")
	rootCmd.AddCommand(addCmd)
}
