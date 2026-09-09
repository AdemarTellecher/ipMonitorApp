//go:build !windows

package service

import (
	"os/exec"
	"runtime"
)

// pingSystemCommand executa o ping nativo em sistemas Unix-like (macOS / Linux)
func pingSystemCommand(target string) bool {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		// No macOS: -c 1 (1 pacote), -t 2 (timeout de 2s)
		cmd = exec.Command("ping", "-c", "1", "-t", "2", target)
	case "linux":
		// No Linux: -c 1 (1 pacote), -W 2 (timeout de 2s)
		cmd = exec.Command("ping", "-c", "1", "-W", "2", target)
	default:
		cmd = exec.Command("ping", "-c", "1", target)
	}

	err := cmd.Run()
	return err == nil
}
