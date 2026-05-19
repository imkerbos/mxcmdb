package sshutil

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"golang.org/x/crypto/ssh"
)

// ExecResult 命令执行结果
type ExecResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Duration time.Duration
	Err      error
}

// RunCommand 在 SSH 连接上执行命令
func RunCommand(client *ssh.Client, command string, timeout time.Duration) ExecResult {
	start := time.Now()

	session, err := client.NewSession()
	if err != nil {
		return ExecResult{Err: fmt.Errorf("create session: %w", err), Duration: time.Since(start)}
	}
	defer session.Close()

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	// 使用 context 控制超时
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- session.Run(command)
	}()

	select {
	case <-ctx.Done():
		_ = session.Signal(ssh.SIGKILL)
		return ExecResult{
			Stdout:   stdout.String(),
			Stderr:   stderr.String(),
			ExitCode: -1,
			Duration: time.Since(start),
			Err:      fmt.Errorf("command timeout after %v", timeout),
		}
	case err := <-done:
		result := ExecResult{
			Stdout:   stdout.String(),
			Stderr:   stderr.String(),
			Duration: time.Since(start),
		}
		if err != nil {
			if exitErr, ok := err.(*ssh.ExitError); ok {
				result.ExitCode = exitErr.ExitStatus()
			} else {
				result.ExitCode = -1
				result.Err = err
			}
		}
		return result
	}
}
