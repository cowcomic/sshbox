package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "列出所有SSH连接",
	RunE: func(cmd *cobra.Command, args []string) error {
		tag, _ := cmd.Flags().GetString("tag")
		search, _ := cmd.Flags().GetString("search")
		format, _ := cmd.Flags().GetString("format")

		conns, err := manager.ListConnections(tag, search)
		if err != nil {
			return err
		}

		if len(conns) == 0 {
			cmd.Println("没有找到连接")
			return nil
		}

		switch format {
		case "json":
			data, err := json.MarshalIndent(conns, "", "  ")
			if err != nil {
				return err
			}
			cmd.Println(string(data))
		default:
			headers := []string{"名称", "主机", "端口", "用户", "标签"}
			var rows [][]string
			for _, c := range conns {
				rows = append(rows, []string{c.Name, c.Host, fmt.Sprintf("%d", c.Port), c.User, joinTags(c.Tags)})
			}
			printTable(cmd.OutOrStdout(), headers, rows)
		}

		return nil
	},
}

func init() {
	listCmd.Flags().String("tag", "", "按标签筛选")
	listCmd.Flags().String("search", "", "搜索关键词")
	listCmd.Flags().String("format", "table", "输出格式（table/json）")
	rootCmd.AddCommand(listCmd)
}

func joinTags(tags []string) string {
	if len(tags) == 0 {
		return ""
	}
	result := ""
	for i, t := range tags {
		if i > 0 {
			result += ","
		}
		result += t
	}
	return result
}
