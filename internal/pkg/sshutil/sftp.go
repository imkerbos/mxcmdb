package sshutil

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// SFTPUpload 通过 SFTP 上传文件到远程主机
func SFTPUpload(client *ssh.Client, localPath, remotePath string, mode os.FileMode) error {
	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return fmt.Errorf("sftp client: %w", err)
	}
	defer sftpClient.Close()

	// 远程路径以 / 结尾或是已存在的目录时，拼接本地文件名
	if strings.HasSuffix(remotePath, "/") {
		remotePath = remotePath + filepath.Base(localPath)
	} else if info, err := sftpClient.Stat(remotePath); err == nil && info.IsDir() {
		remotePath = remotePath + "/" + filepath.Base(localPath)
	}

	// 确保远程目录存在
	remoteDir := filepath.Dir(remotePath)
	if err := sftpClient.MkdirAll(remoteDir); err != nil {
		return fmt.Errorf("mkdir %s: %w", remoteDir, err)
	}

	localFile, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("open local file: %w", err)
	}
	defer localFile.Close()

	remoteFile, err := sftpClient.Create(remotePath)
	if err != nil {
		return fmt.Errorf("create remote file: %w", err)
	}
	defer remoteFile.Close()

	if _, err := io.Copy(remoteFile, localFile); err != nil {
		return fmt.Errorf("copy file: %w", err)
	}

	if err := sftpClient.Chmod(remotePath, mode); err != nil {
		return fmt.Errorf("chmod: %w", err)
	}

	return nil
}
