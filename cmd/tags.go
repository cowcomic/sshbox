package cmd

import (
	"fmt"
	"sort"
	"text/tabwriter"

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

	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "标签\t连接数")
	for _, k := range keys {
		fmt.Fprintf(w, "%s\t%d\n", k, counts[k])
	}
	w.Flush()
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
	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "名称\t主机\t端口\t用户")
	for _, c := range conns {
		fmt.Fprintf(w, "%s\t%s\t%d\t%s\n", c.Name, c.Host, c.Port, c.User)
	}
	w.Flush()
	return nil
}

func init() {
	rootCmd.AddCommand(tagsCmd)
}
