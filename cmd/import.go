package cmd

import (
	"github.com/spf13/cobra"
)

var importCmd = &cobra.Command{
	Use:   "import <file>",
	Short: "导入配置",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		file := args[0]
		merge, _ := cmd.Flags().GetBool("merge")

		count, err := manager.ImportFromFile(file, merge)
		if err != nil {
			return err
		}

		cmd.Printf("✓ 成功导入 %d 个连接\n", count)
		return nil
	},
}

func init() {
	importCmd.Flags().Bool("merge", false, "合并模式（保留现有连接）")
	rootCmd.AddCommand(importCmd)
}
