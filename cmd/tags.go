package cmd

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"
)

var tagsCmd = &cobra.Command{
	Use:   "tags [tag]",
	Short: "列出标签及其连接",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return listAllTags(cmd)
		}
		return listTagConnections(cmd, args[0])
	},
}

func listAllTags(cmd *cobra.Command) error {
	counts, err := manager.ListTags()
	if err != nil {
		return err
	}

	if len(counts) == 0 {
		cmd.Println("没有标签")
		return nil
	}

	keys := make([]string, 0, len(counts))
	for k := range counts {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	headers := []string{"标签", "连接数"}
	var rows [][]string
	for _, k := range keys {
		rows = append(rows, []string{k, fmt.Sprintf("%d", counts[k])})
	}
	printTable(cmd.OutOrStdout(), headers, rows)
	return nil
}

func listTagConnections(cmd *cobra.Command, tag string) error {
	conns, err := manager.ListConnections(tag, "")
	if err != nil {
		return err
	}

	if len(conns) == 0 {
		cmd.Printf("标签 %s 下没有连接\n", tag)
		return nil
	}

	cmd.Printf("标签 %s (%d 个连接):\n\n", tag, len(conns))
	headers := []string{"名称", "主机", "端口", "用户"}
	var rows [][]string
	for _, c := range conns {
		rows = append(rows, []string{c.Name, c.Host, fmt.Sprintf("%d", c.Port), c.User})
	}
	printTable(cmd.OutOrStdout(), headers, rows)
	return nil
}

func init() {
	rootCmd.AddCommand(tagsCmd)
}
