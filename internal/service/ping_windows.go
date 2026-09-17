//go:build windows

package service

import (
	"os/exec"
	"syscall"
)

// pingSystemCommand executa o ping.exe nativo do Windows sem abrir janela de prompt (cmd/conhost)
// e analisa a saída textual para evitar falsos positivos causados por códigos de saída 0 em erros ICMP (ex: TTL expirado).
func pingSystemCommand(target string) bool {
	cmd := exec.Command("ping", "-n", "1", "-w", "1500", target)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}
	return isWindowsPingOutputSuccess(string(out), target)
}


