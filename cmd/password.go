package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var passwordCmd = &cobra.Command{
	Use:   "password <name>",
	Short: "查看连接密码",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		pass, err := manager.GetDecryptedPassword(name)
		if err != nil {
			return err
		}

		fmt.Fprintln(cmd.OutOrStdout(), pass)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(passwordCmd)
}
