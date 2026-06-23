package cmd

import (
	"fmt"
	"sort"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var tagsCmd = &cobra.Command{
	Use:   "tags",
	Short: "列出所有标签",
	RunE: func(cmd *cobra.Command, args []string) error {
		counts, err := manager.ListTags()
		if err != nil {
			return err
		}

		if len(counts) == 0 {
			cmd.Println("没有标签")
			return nil
		}

		// Sort tags for consistent output
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
	},
}

func init() {
	rootCmd.AddCommand(tagsCmd)
}
