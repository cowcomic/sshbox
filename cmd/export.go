package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "导出配置",
	RunE: func(cmd *cobra.Command, args []string) error {
		output, _ := cmd.Flags().GetString("output")
		tag, _ := cmd.Flags().GetString("tag")

		data, err := manager.ExportJSON(tag)
		if err != nil {
			return err
		}

		if output != "" {
			if err := os.WriteFile(output, data, 0600); err != nil {
				return fmt.Errorf("写入文件失败: %w", err)
			}
			cmd.Printf("✓ 配置已导出到 %s\n", output)
		} else {
			cmd.Println(string(data))
		}

		return nil
	},
}

func init() {
	exportCmd.Flags().String("output", "", "输出文件路径")
	exportCmd.Flags().String("tag", "", "只导出指定标签的连接")
	rootCmd.AddCommand(exportCmd)
}
