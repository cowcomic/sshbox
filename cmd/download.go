package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"sshbox/internal/ssh"

	"github.com/spf13/cobra"
)

var downloadCmd = &cobra.Command{
	Use:     "download <name> <remote_path> [local_path]",
	Short:   "从远程服务器下载文件",
	Aliases: []string{"down"},
	Args:    cobra.RangeArgs(2, 3),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		remotePath := args[1]
		recursive, _ := cmd.Flags().GetBool("recursive")

		conn, err := manager.GetConnection(name)
		if err != nil {
			return err
		}

		password, err := manager.GetDecryptedPassword(name)
		if err != nil {
			return fmt.Errorf("解密密码失败: %w", err)
		}

		// 确定本地路径
		localPath := "."
		if len(args) >= 3 {
			localPath = args[2]
		}

		// 建立 SFTP 连接
		fmt.Printf("正在连接 %s (%s:%d)...\n", name, conn.Host, conn.Port)
		sftpClient, sshClient, err := ssh.DialSFTP(conn.Host, conn.Port, conn.User, password)
		if err != nil {
			return err
		}
		defer sftpClient.Close()
		defer sshClient.Close()

		// 解析远端路径
		remotePath = ssh.ResolveRemotePath(sftpClient, remotePath)

		// 获取远端路径信息
		info, err := sftpClient.Stat(remotePath)
		if err != nil {
			return fmt.Errorf("远端路径不存在: %w", err)
		}

		if info.IsDir() {
			if !recursive {
				return fmt.Errorf("%s 是目录，请使用 -r 参数", remotePath)
			}
			// 目录下载：如果本地路径是已存在的目录，在其下创建同名子目录
			if li, err := os.Stat(localPath); err == nil && li.IsDir() {
				localPath = filepath.Join(localPath, filepath.Base(remotePath))
			}
			fmt.Printf("下载目录 %s:%s → %s\n", name, remotePath, localPath)
			return ssh.DownloadDir(sftpClient, remotePath, localPath)
		}

		// 文件下载：如果本地路径是目录，自动拼接文件名
		if li, err := os.Stat(localPath); err == nil && li.IsDir() {
			localPath = filepath.Join(localPath, filepath.Base(remotePath))
		}
		fmt.Printf("下载文件 %s:%s → %s\n", name, remotePath, localPath)
		return ssh.DownloadFile(sftpClient, remotePath, localPath)
	},
}

func init() {
	downloadCmd.Flags().BoolP("recursive", "r", false, "递归下载目录")
	rootCmd.AddCommand(downloadCmd)
}
