package cmd

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"

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
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "名称\t主机\t端口\t用户\t标签")
			for _, c := range conns {
				fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\n", c.Name, c.Host, c.Port, c.User, joinTags(c.Tags))
			}
			w.Flush()
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
