package ssh

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/sftp"
	gossh "golang.org/x/crypto/ssh"
)

// DialSFTP establishes an SFTP connection using SSH credentials.
func DialSFTP(host string, port int, user, password string) (*sftp.Client, *gossh.Client, error) {
	config := &gossh.ClientConfig{
		User:            user,
		Auth:            []gossh.AuthMethod{gossh.Password(password)},
		HostKeyCallback: gossh.InsecureIgnoreHostKey(),
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	sshClient, err := gossh.Dial("tcp", addr, config)
	if err != nil {
		return nil, nil, fmt.Errorf("SSH连接失败: %w", err)
	}

	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		sshClient.Close()
		return nil, nil, fmt.Errorf("SFTP连接失败: %w", err)
	}

	return sftpClient, sshClient, nil
}

// UploadFile uploads a single local file to the remote path.
func UploadFile(client *sftp.Client, local, remote string) error {
	info, err := os.Stat(local)
	if err != nil {
		return fmt.Errorf("读取本地文件失败: %w", err)
	}

	src, err := os.Open(local)
	if err != nil {
		return err
	}
	defer src.Close()

	// 如果 remote 是目录，自动拼接文件名
	if ri, err := client.Stat(remote); err == nil && ri.IsDir() {
		remote = filepath.Join(remote, filepath.Base(local))
	}

	dst, err := client.Create(remote)
	if err != nil {
		return fmt.Errorf("创建远端文件失败: %w", err)
	}
	defer dst.Close()

	return copyWithProgress(dst, src, info.Size(), filepath.Base(local))
}

// DownloadFile downloads a single remote file to the local path.
func DownloadFile(client *sftp.Client, remote, local string) error {
	info, err := client.Stat(remote)
	if err != nil {
		return fmt.Errorf("远端文件不存在: %w", err)
	}
	if info.IsDir() {
		return fmt.Errorf("%s 是目录，请使用 -r 参数", remote)
	}

	// 如果 local 是目录，自动拼接文件名
	if li, err := os.Stat(local); err == nil && li.IsDir() {
		local = filepath.Join(local, filepath.Base(remote))
	}

	src, err := client.Open(remote)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(local)
	if err != nil {
		return fmt.Errorf("创建本地文件失败: %w", err)
	}
	defer dst.Close()

	return copyWithProgress(dst, src, info.Size(), filepath.Base(remote))
}

// UploadDir recursively uploads a local directory to the remote path.
func UploadDir(client *sftp.Client, localDir, remoteDir string) error {
	localDir, err := filepath.Abs(localDir)
	if err != nil {
		return err
	}

	info, err := os.Stat(localDir)
	if err != nil {
		return fmt.Errorf("本地目录不存在: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s 不是目录", localDir)
	}

	// 创建远端根目录
	client.Mkdir(remoteDir)

	return filepath.Walk(localDir, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, _ := filepath.Rel(localDir, path)
		rel = filepath.ToSlash(rel) // 统一用 /
		remotePath := remoteDir + "/" + rel

		if fi.IsDir() {
			return client.Mkdir(remotePath)
		}

		return UploadFile(client, path, remotePath)
	})
}

// DownloadDir recursively downloads a remote directory to the local path.
func DownloadDir(client *sftp.Client, remoteDir, localDir string) error {
	info, err := client.Stat(remoteDir)
	if err != nil {
		return fmt.Errorf("远端目录不存在: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s 不是目录", remoteDir)
	}

	if err := os.MkdirAll(localDir, 0755); err != nil {
		return fmt.Errorf("创建本地目录失败: %w", err)
	}

	return downloadDirRecursive(client, remoteDir, localDir)
}

func downloadDirRecursive(client *sftp.Client, remoteDir, localDir string) error {
	entries, err := client.ReadDir(remoteDir)
	if err != nil {
		return fmt.Errorf("读取远端目录失败: %w", err)
	}

	for _, entry := range entries {
		remotePath := remoteDir + "/" + entry.Name()
		localPath := filepath.Join(localDir, entry.Name())

		if entry.IsDir() {
			if err := os.MkdirAll(localPath, 0755); err != nil {
				return err
			}
			if err := downloadDirRecursive(client, remotePath, localPath); err != nil {
				return err
			}
		} else {
			if err := DownloadFile(client, remotePath, localPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyWithProgress copies from src to dst with progress display.
func copyWithProgress(dst io.Writer, src io.Reader, total int64, name string) error {
	buf := make([]byte, 32*1024)
	var copied int64
	start := time.Now()
	lastPrint := start

	for {
		n, readErr := src.Read(buf)
		if n > 0 {
			written, writeErr := dst.Write(buf[:n])
			if writeErr != nil {
				return writeErr
			}
			copied += int64(written)

			now := time.Now()
			if now.Sub(lastPrint) >= 200*time.Millisecond || readErr != nil {
				printProgress(name, copied, total, start)
				lastPrint = now
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return readErr
		}
	}

	printProgress(name, total, total, start)
	fmt.Println() // 换行
	return nil
}

func printProgress(name string, current, total int64, start time.Time) {
	elapsed := time.Since(start).Seconds()
	if elapsed == 0 {
		elapsed = 0.001
	}

	speed := float64(current) / elapsed
	var percent float64
	if total > 0 {
		percent = float64(current) / float64(total) * 100
	}

	fmt.Printf("\r  %-20s  %s / %s  %.1f%%  %s/s",
		truncName(name, 20),
		humanSize(current),
		humanSize(total),
		percent,
		humanSize(int64(speed)),
	)
}

func truncName(name string, max int) string {
	if len(name) <= max {
		return name
	}
	return "..." + name[len(name)-max+3:]
}

func humanSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %c", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// ResolveRemotePath resolves a remote path, expanding ~ to the user's home directory.
func ResolveRemotePath(client *sftp.Client, path string) string {
	if !strings.HasPrefix(path, "~/") && path != "~" {
		return path
	}
	home, err := client.Getwd()
	if err != nil {
		return path
	}
	if path == "~" {
		return home
	}
	return home + "/" + path[2:]
}
