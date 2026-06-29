package cmd

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"sshbox/internal/ssh"

	"github.com/spf13/cobra"
)

var uploadCmd = &cobra.Command{
	Use:     "upload <name> <local_path> [remote_path]",
	Short:   "上传文件到远程服务器 (up)",
	Aliases: []string{"up"},
	Args:    cobra.RangeArgs(2, 3),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		localPath := args[1]
		recursive, _ := cmd.Flags().GetBool("recursive")

		conn, err := manager.GetConnection(name)
		if err != nil {
			return err
		}

		password, err := manager.GetDecryptedPassword(name)
		if err != nil {
			return fmt.Errorf("解密密码失败: %w", err)
		}

		// 确定远端路径
		remotePath := ""
		if len(args) >= 3 {
			remotePath = args[2]
		}

		// 检查本地路径
		info, err := os.Stat(localPath)
		if err != nil {
			return fmt.Errorf("本地路径不存在: %w", err)
		}

		if info.IsDir() && !recursive {
			return fmt.Errorf("%s 是目录，请使用 -r 参数", localPath)
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
		if remotePath == "" {
			remotePath, _ = sftpClient.Getwd()
		} else {
			remotePath = ssh.ResolveRemotePath(sftpClient, remotePath)
		}

		if info.IsDir() {
			// 目录上传：如果远端路径不存在，用目录名
			if ri, err := sftpClient.Stat(remotePath); err != nil || !ri.IsDir() {
				remotePath = path.Join(remotePath, filepath.Base(localPath))
			}
			fmt.Printf("上传目录 %s → %s:%s\n", localPath, name, remotePath)
			return ssh.UploadDir(sftpClient, localPath, remotePath)
		}

		fmt.Printf("上传文件 %s → %s:%s\n", localPath, name, remotePath)
		return ssh.UploadFile(sftpClient, localPath, remotePath)
	},
}

func init() {
	uploadCmd.Flags().BoolP("recursive", "r", false, "递归上传目录")
	rootCmd.AddCommand(uploadCmd)
}
