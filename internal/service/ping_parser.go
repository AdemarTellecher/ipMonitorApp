package service

import "strings"

// isWindowsPingOutputSuccess verifica se a resposta do ping do Windows indica sucesso real de eco do host destino
func isWindowsPingOutputSuccess(output, target string) bool {
	lower := strings.ToLower(output)

	// Indicadores explícitos de falha/erro ICMP intermediário ou esgotamento
	// Mesmo com pacotes recebidos (de roteadores intermediários), estes são falhas:
	errorKeywords := []string{
		"ttl expirou",
		"ttl expired",
		"time to live exceeded",
		"inacessível",
		"inacessivel",
		"unreachable",
		"esgotado o tempo",
		"timed out",
		"falha geral",
		"general failure",
		"não foi possível encontrar",
		"could not find host",
		"100% de perda",
		"100% loss",
	}

	for _, kw := range errorKeywords {
		if strings.Contains(lower, kw) {
			return false
		}
	}

	// Para ser considerado sucesso real do host alvo:
	// 1. Deve conter tempo de resposta (tempo= / tempo< / time= / time<)
	hasTime := strings.Contains(lower, "tempo=") ||
		strings.Contains(lower, "tempo<") ||
		strings.Contains(lower, "time=") ||
		strings.Contains(lower, "time<")

	// 2. Deve conter contagem de bytes de payload (bytes=)
	hasBytes := strings.Contains(lower, "bytes=")

	return hasTime && hasBytes
}

// isUnixPingOutputSuccess valida se o utilitário nativo Unix/Linux/macOS realmente obteve Echo Reply
func isUnixPingOutputSuccess(output string) bool {
	lower := strings.ToLower(output)

	// Falhas explícitas em sistemas Unix
	errorKeywords := []string{
		"100% packet loss",
		"0 packets received",
		"0 packets transmitted",
		"destination host unreachable",
		"time to live exceeded",
		"request timeout",
		"unknown host",
	}

	for _, kw := range errorKeywords {
		if strings.Contains(lower, kw) {
			return false
		}
	}

	// Deve conter indicação de pacote recebido (ex: "1 packets received" ou "1 received")
	hasReceived := strings.Contains(lower, "1 packets received") ||
		strings.Contains(lower, "1 received") ||
		strings.Contains(lower, "bytes from")

	return hasReceived
}
