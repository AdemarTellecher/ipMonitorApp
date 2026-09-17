package service

import (
	"testing"
	"time"

	"github.com/AdemarTellecher/ipmonitorapp/internal/model"
)

// TestWindowsPingOutputValidation valida o parser de saída do utilitário nativo do Windows ping.exe
func TestWindowsPingOutputValidation(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		target   string
		expected bool
	}{
		{
			name: "TTL expirado em trânsito (caso reportado pelo usuário com ExitCode 0)",
			output: `Disparando 168.195.155.240 com 32 bytes de dados:
Resposta de 10.128.16.105: A vida útil (TTL) expirou em trânsito.
Resposta de 10.128.16.105: A vida útil (TTL) expirou em trânsito.

Estatísticas do Ping para 168.195.155.240:
    Pacotes: Enviados = 2, Recebidos = 2, Perdidos = 0 (0% de perda),`,
			target:   "168.195.155.240",
			expected: false,
		},
		{
			name: "Host de destino inacessível (Português)",
			output: `Disparando 192.168.100.200 com 32 bytes de dados:
Resposta de 192.168.100.1: Host de destino inacessível.

Estatísticas do Ping para 192.168.100.200:
    Pacotes: Enviados = 1, Recebidos = 1, Perdidos = 0 (0% de perda),`,
			target:   "192.168.100.200",
			expected: false,
		},
		{
			name: "Esgotado o tempo limite do pedido (Português)",
			output: `Disparando 10.254.254.1 com 32 bytes de dados:
Esgotado o tempo limite do pedido.

Estatísticas do Ping para 10.254.254.1:
    Pacotes: Enviados = 1, Recebidos = 0, Perdidos = 1 (100% de perda),`,
			target:   "10.254.254.1",
			expected: false,
		},
		{
			name: "TTL expired in transit (Inglês)",
			output: `Pinging 168.195.155.240 with 32 bytes of data:
Reply from 10.128.16.105: TTL expired in transit.

Ping statistics for 168.195.155.240:
    Packets: Sent = 1, Received = 1, Lost = 0 (0% loss),`,
			target:   "168.195.155.240",
			expected: false,
		},
		{
			name: "Destination host unreachable (Inglês)",
			output: `Pinging 192.168.10.99 with 32 bytes of data:
Reply from 192.168.10.1: Destination host unreachable.

Ping statistics for 192.168.10.99:
    Packets: Sent = 1, Received = 1, Lost = 0 (0% loss),`,
			target:   "192.168.10.99",
			expected: false,
		},
		{
			name: "Request timed out (Inglês)",
			output: `Pinging 1.2.3.4 with 32 bytes of data:
Request timed out.

Ping statistics for 1.2.3.4:
    Packets: Sent = 1, Received = 0, Lost = 1 (100% loss),`,
			target:   "1.2.3.4",
			expected: false,
		},
		{
			name: "Resposta de Sucesso Legítima (Português)",
			output: `Disparando 8.8.8.8 com 32 bytes de dados:
Resposta de 8.8.8.8: bytes=32 tempo=18ms TTL=116

Estatísticas do Ping para 8.8.8.8:
    Pacotes: Enviados = 1, Recebidos = 1, Perdidos = 0 (0% de perda),
Aproximar um número redondo de vezes em milissegundos:
    Mínimo = 18ms, Máximo = 18ms, Média = 18ms`,
			target:   "8.8.8.8",
			expected: true,
		},
		{
			name: "Resposta de Sucesso Legítima (Inglês)",
			output: `Pinging 1.1.1.1 with 32 bytes of data:
Reply from 1.1.1.1: bytes=32 time=12ms TTL=57

Ping statistics for 1.1.1.1:
    Packets: Sent = 1, Received = 1, Lost = 0 (0% loss),
Approximate round trip times in milli-seconds:
    Minimum = 12ms, Maximum = 12ms, Average = 12ms`,
			target:   "1.1.1.1",
			expected: true,
		},
		{
			name: "Resposta com tempo inferior a 1ms (tempo<1ms)",
			output: `Disparando 127.0.0.1 com 32 bytes de dados:
Resposta de 127.0.0.1: bytes=32 tempo<1ms TTL=128

Estatísticas do Ping para 127.0.0.1:
    Pacotes: Enviados = 1, Recebidos = 1, Perdidos = 0 (0% de perda),`,
			target:   "127.0.0.1",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isWindowsPingOutputSuccess(tt.output, tt.target)
			if got != tt.expected {
				t.Errorf("isWindowsPingOutputSuccess() = %v, esperado %v para o caso '%s'", got, tt.expected, tt.name)
			}
		})
	}
}

// TestUnixPingOutputValidation valida o parser de saída Unix/Linux/macOS
func TestUnixPingOutputValidation(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		expected bool
	}{
		{
			name: "Sucesso no Linux",
			output: `PING 8.8.8.8 (8.8.8.8) 56(84) bytes of data.
64 bytes from 8.8.8.8: icmp_seq=1 ttl=116 time=18.2 ms

--- 8.8.8.8 ping statistics ---
1 packets transmitted, 1 received, 0% packet loss, time 0ms
rtt min/avg/max/mdev = 18.211/18.211/18.211/0.000 ms`,
			expected: true,
		},
		{
			name: "Sucesso no macOS",
			output: `PING 1.1.1.1 (1.1.1.1): 56 data bytes
64 bytes from 1.1.1.1: icmp_seq=0 ttl=57 time=14.120 ms

--- 1.1.1.1 ping statistics ---
1 packets transmitted, 1 packets received, 0.0% packet loss
round-trip min/avg/max/stddev = 14.120/14.120/14.120/0.000 ms`,
			expected: true,
		},
		{
			name: "100% packet loss no Linux",
			output: `PING 10.254.254.1 (10.254.254.1) 56(84) bytes of data.

--- 10.254.254.1 ping statistics ---
1 packets transmitted, 0 received, 100% packet loss, time 0ms`,
			expected: false,
		},
		{
			name: "Destination Host Unreachable no Linux",
			output: `PING 192.168.1.99 (192.168.1.99) 56(84) bytes of data.
From 192.168.1.1 icmp_seq=1 Destination Host Unreachable

--- 192.168.1.99 ping statistics ---
1 packets transmitted, 0 received, +1 errors, 100% packet loss, time 0ms`,
			expected: false,
		},
		{
			name: "Time to live exceeded no Unix",
			output: `PING 168.195.155.240 (168.195.155.240) 56(84) bytes of data.
From 10.128.16.105 icmp_seq=1 Time to live exceeded

--- 168.195.155.240 ping statistics ---
1 packets transmitted, 0 received, +1 errors, 100% packet loss, time 0ms`,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isUnixPingOutputSuccess(tt.output)
			if got != tt.expected {
				t.Errorf("isUnixPingOutputSuccess() = %v, esperado %v para o caso '%s'", got, tt.expected, tt.name)
			}
		})
	}
}

// TestCheckDeviceConnectivity_RealTargets testa a validação real com IPs
func TestCheckDeviceConnectivity_RealTargets(t *testing.T) {
	// IP reportado pelo usuário que dá "TTL expirou em trânsito"
	userProblematicDevice := model.IPDevice{
		IP:          "168.195.155.240",
		Method:      "PING",
		ThresholdMs: 2000,
	}

	status := checkDeviceConnectivity(userProblematicDevice)
	if status != "Offline" {
		t.Errorf("O IP 168.195.155.240 deveria ser reportado como 'Offline', mas retornou: %s", status)
	}

	// Host inalcançável de documentação RFC 5737
	rfcDevice := model.IPDevice{
		IP:          "192.0.2.1",
		Method:      "PING",
		ThresholdMs: 1000,
	}
	rfcStatus := checkDeviceConnectivity(rfcDevice)
	if rfcStatus != "Offline" {
		t.Errorf("O IP reservado 192.0.2.1 deveria ser 'Offline', mas retornou: %s", rfcStatus)
	}

	// Host local ativo
	localDevice := model.IPDevice{
		IP:          "127.0.0.1",
		Method:      "PING",
		ThresholdMs: 2000,
	}
	localStatus := checkDeviceConnectivity(localDevice)
	if localStatus != "Online" {
		t.Logf("Aviso: ping em 127.0.0.1 retornou %s (pode depender de permissões ICMP no ambiente)", localStatus)
	}
}

// TestCheckIPWithTimeout_EmptyOrWhitespace valida proteção de entradas vazias
func TestCheckIPWithTimeout_EmptyOrWhitespace(t *testing.T) {
	if checkIPWithTimeout("", 500*time.Millisecond) != "Offline" {
		t.Errorf("Target vazio deveria retornar Offline")
	}
	if checkIPWithTimeout("   ", 500*time.Millisecond) != "Offline" {
		t.Errorf("Target apenas com espaços deveria retornar Offline")
	}
}
