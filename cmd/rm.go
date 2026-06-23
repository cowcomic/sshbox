package cmd

import (
	"bufio"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var rmCmd = &cobra.Command{
	Use:   "rm <name>",
	Short: "删除SSH连接",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		// Confirm deletion
		cmd.Printf("确定要删除连接 %s 吗？(y/N): ", name)
		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "y" && answer != "yes" {
			cmd.Println("已取消")
			return nil
		}

		if err := manager.DeleteConnection(name); err != nil {
			return err
		}

		cmd.Printf("✓ 连接 %s 已删除\n", name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(rmCmd)
}
