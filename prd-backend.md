# Product Requirements Document (PRD) — Backend
## IP Monitor App (Go / Wails v3 / SQLite)

**Versão do Documento:** 1.0  
**Versão Atual da Aplicação:** `v2.5.4`  
**Autor:** Ademar Tellecher & Antigravity  
**Status:** Aprovado / Produção  
**Stack Principal:** Go 1.25+ (`CGO_ENABLED=0`), Wails v3 (`v3.0.0-beta.17`), `modernc.org/sqlite` (Pure Go), `github.com/go-ping/ping`.

---

## 1. Visão Geral do Produto (Executive Summary)

O **IP Monitor App** é uma aplicação desktop multiplataforma (Windows, macOS e Linux), ultra-leve e portátil, projetada para monitoramento e diagnóstico de conectividade de rede (endereços IPv4, IPv6, Hostnames e URLs) em tempo real.

O backend em Go atua como o **motor central e autoridade definitiva sobre os dados e regras de negócio** ("Golden Rule: Go owns data and business rules"), provendo:
- Persistência autocontida via SQLite em Go puro (sem runtime C ou CGO).
- Resolução dinâmica e resiliente do caminho do banco para compatibilidade com empacotamento desktop (macOS App Bundles, Gatekeeper Translocation e Windows Portable).
- Orquestração de rotinas periódicas de varredura concorrente com emissão de eventos reativos (`app.Event.Emit`).
- Motor híbrido de teste de conectividade (ICMP direto com privilégios adequados, fallback nativo do S.O. oculto e fallback de portas TCP).
- Exposição dual de serviços via **Native Wails v3 Bindings** e **HTTP REST API** (`/api/...`) consumível pelo WebView nativo.

---

## 2. Objetivos de Engenharia e Princípios de Design

1. **Ultra Portabilidade e Zero Dependências Externas:**
   - Binário único autocontido integrando assets e runtime.
   - Compilação sem dependência de toolchains C (GCC, MinGW-w64 ou WinLibs) usando driver SQLite 100% Go (`modernc.org/sqlite`).
2. **Frontend como Visão Burra (Dumb UI / Presentational View):**
   - Nenhuma lógica de ordenação de prioridade, métrica numérica agregada ou validação de rede reside no JavaScript/CSS.
   - O backend fornece objetos ricos prontos para exibição (`NetworkOverview`).
3. **Eliminação de Polling no Frontend:**
   - Proibido `setInterval` no frontend para consulta contínua de status. A comunicação de atualização periódica ocorre estritamente por barramento reativo de eventos (`ips-updated`).
4. **Resiliência Multiplataforma Específica:**
   - Tratamento de particularidades de permissões de sockets ICMP em Unix vs Windows.
   - Execução de comandos do sistema sem criar consoles visuais no Windows (`HideWindow` e `CREATE_NO_WINDOW`).
   - Respeito às políticas de permissão de escrita de disco do macOS Sandbox/App Bundle.

---

## 3. Arquitetura do Backend

### 3.1 Estrutura de Pacotes (`internal/`)

```
ipMonitorApp/
├── main.go                       # Entrypoint: Inicialização Wails v3, lifecycle hooks e goroutine periódica
├── internal/
│   ├── model/                    # Camada de Dados, Entidades e Persistência SQLite
│   │   ├── model.go              # Schema, migrações transparentes, repositório e resolução de path
│   │   └── model_test.go         # Testes de persistência
│   ├── service/                  # Camada de Domínio, Regras de Negócio e HTTP Handlers
│   │   ├── monitor_service.go    # Serviços de ping, concorrência, importação e validação
│   │   ├── ping_windows.go       # Implementação de fallback nativo para Windows (CREATE_NO_WINDOW)
│   │   ├── ping_other.go         # Implementação de fallback nativo para macOS/Linux (/sbin/ping)
│   │   └── service_test.go       # Testes unitários com banco in-memory/temp
│   ├── version/
│   │   └── version.go            # Constante centralizada de versão (v2.5.4)
│   └── assets/                   # Ícones e imagens embutidas (go:embed)
```

---

## 4. Modelagem de Dados e Persistência (SQLite)

### 4.1 Resolução Dinâmica de Diretório (`ResolveDBPath`)
- **macOS (`runtime.GOOS == "darwin"`):**
  - Diretório: `~/Library/Application Support/IPMonitor/` com permissões `0755`.
  - *Justificativa Técnica:* O bundle `/Applications/IP Monitor.app/Contents/MacOS/` é somente-leitura (`root:wheel`). Tentativas de gravar no diretório do binário provocam erros fatais de `permission denied` e travamento na inicialização.
- **Windows / Linux:**
  - Diretório: Pasta do executável (`filepath.Dir(os.Executable())`).
  - *Justificativa Técnica:* Garante portabilidade total (execução a partir de pendrives ou diretórios customizados).

### 4.2 Esquema do Banco de Dados (`ips`)

```sql
CREATE TABLE IF NOT EXISTS ips (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    ip TEXT NOT NULL UNIQUE,
    status TEXT,
    name TEXT DEFAULT '',
    method TEXT DEFAULT 'PING',
    threshold_ms INTEGER DEFAULT 2000,
    uuid TEXT DEFAULT ''
);
```

#### Migração Automática e Transparente
Na inicialização do repositório (`NewRepository`), o backend inspeciona a tabela via `pragma_table_info('ips')` e aplica comandos `ALTER TABLE ips ADD COLUMN ...` automaticamente para colunas ausentes (`name`, `method`, `threshold_ms`, `uuid`), assegurando retrocompatibilidade total com versões anteriores sem perda de dados.

### 4.3 Entidades do Domínio

#### `IPDevice`
```go
type IPDevice struct {
    ID          int    `json:"id"`
    IP          string `json:"ip"`
    Status      string `json:"status"`       // "Online", "Offline", "Desconhecido"
    Name        string `json:"name"`         // Nome amigável / Identificação
    Method      string `json:"method"`       // "PING" ou "TCP"
    ThresholdMs int    `json:"thresholdMs"`  // Timeout em milissegundos (default: 2000)
    UUID        string `json:"uuid"`         // Identificador externo / UUID importado
}
```

#### `NetworkOverview`
```go
type NetworkOverview struct {
    Total   int        `json:"total"`
    Online  int        `json:"online"`
    Offline int        `json:"offline"`
    Devices []IPDevice `json:"devices"`
}
```

### 4.4 Regra de Ordenação Canônica (SQL)
A ordenação de prioridade é imposta **diretamente no SQL** pelo backend para garantir visibilidade imediata de falhas:
```sql
ORDER BY 
    CASE status 
        WHEN 'Offline' THEN 0 
        WHEN 'Desconhecido' THEN 1 
        WHEN 'Online' THEN 2 
        ELSE 3 
    END, 
    ip ASC
```

---

## 5. Motor de Conectividade e Diagnóstico de Rede

O diagnóstico de conectividade avalia cada host em 3 camadas sucessivas com base nas preferências configuradas:

### 5.1 Fluxo de Avaliação por Método
1. **Método `TCP`:**
   - Realiza `net.DialTimeout` sequencial nas portas de serviço padrão: `443`, `80`, `8080`, `22`.
   - Se qualquer porta aceitar conexão antes de atingir `ThresholdMs`: Status = `Online`.
   - Se todas falharem ou expirarem: Status = `Offline`.

2. **Método `PING` (Padrão):**
   - **Camada 1: Driver ICMP Go (`github.com/go-ping/ping`):**
     - Windows: `pinger.SetPrivileged(true)` (utiliza API WinSock ICMP).
     - macOS / Linux: `pinger.SetPrivileged(false)` (utiliza sockets UDP ICMP desprivilegiados do kernel Darwin/Linux para evitar erro de `operation not permitted` por falta de root).
     - Se receber pelo menos 1 pacote de resposta: Status = `Online`.
   - **Camada 2: Fallback para Utilitário Nativo do S.O. (`pingSystemCommand`):**
     - Ativado caso o driver de socket direto falhe ou não tenha permissão.
     - *macOS:* Executa `/sbin/ping -c 1 -t 2 <target>` (utilitário nativo com `setuid-root`).
     - *Linux:* Executa `ping -c 1 -W 2 <target>`.
     - *Windows:* Executa `ping -n 1 -w 1500 <target>` com flags de processo oculto:
       ```go
       cmd.SysProcAttr = &syscall.SysProcAttr{
           HideWindow:    true,
           CreationFlags: 0x08000000, // CREATE_NO_WINDOW
       }
       ```
       *Objetivo Crítico:* Impede flashes de janela preta do prompt de comando (`cmd.exe`/`conhost.exe`) na tela do usuário durante varreduras.
   - **Camada 3: Fallback de Portas TCP:**
     - Último recurso: tenta `net.DialTimeout` nas portas `443`, `80`, `8080`, `22` com limite de 1000ms.
     - Se responder: Status = `Online`. Caso contrário: Status = `Offline`.

### 5.2 Concorrência e Performance nas Varreduras
No método `UpdateAllStatuses`:
- As checagens são executadas em Goroutines paralelas.
- O paralelismo é estrangulado por um canal semáforo com capacidade máxima de **20 workers simultâneos** (`sem := make(chan struct{}, 20)`).
- Isso previne esgotamento de descritores de arquivos (*file descriptors*), saturação de buffers de rede ou picos de CPU na máquina host.

---

## 6. Sanitização, Normalização e Validação

### 6.1 Sanitização de Hostnames e URLs (`cleanHostOrIP`)
O backend aceita entradas heterogêneas digitadas ou importadas e realiza normalização idempotente:
- URLs com protocolo (`http://`, `https://`): extrai estritamente o `u.Hostname()`.
- Endereços com porta (`192.168.1.1:8080`, `server.local:3000`): remove a porta via `net.SplitHostPort`.
- Endereços com caminhos (`meuhost.com/dashboard/api`): extrai o host pré-barra.
- Normalização de caixa: converte para minúsculas (`strings.ToLower`) e remove espaços (`TrimSpace`).

### 6.2 Validação Restrita (`isValidHostOrIP`)
- Rejeita strings vazias ou maiores que 253 caracteres.
- Se for IP válido (`net.ParseIP`), aprova imediatamente (IPv4 ou IPv6).
- Para Hostnames/FQDN: valida cada label (até 63 caracteres), proibindo hífens no início/fim e caracteres proibidos (`/: \t\r\n` ou símbolos não-alfanuméricos).

---

## 7. Interfaces de Exposição e Comunicação

O backend disponibiliza seus métodos através de duas vias integradas no Wails v3:

### 7.1 Métodos Nativos Go (Wails Bindings)
Expostos diretamente para a instância da aplicação:
- `s.GetNetworkOverview() (*model.NetworkOverview, error)`
- `s.ListIPs() ([]model.IPDevice, error)`
- `s.AddIPWithDetails(ip, name, method string, thresholdMs int, uuid string) Result`
- `s.EditIPWithDetails(id int, newIP, name, method string, thresholdMs int, uuid string) Result`
- `s.RemoveIP(ip string) Result`
- `s.UpdateAllStatuses() (*model.NetworkOverview, error)`
- `s.ImportJSONContent(content string) Result`
- `s.GetVersion() string`

### 7.2 Roteamento HTTP (`/api/...` via `ServeHTTP`)
Implementa `http.Handler` registrado no Wails na rota `/api`:

| Método | Endpoint | Parâmetros (JSON Body) | Retorno | Descrição |
| :--- | :--- | :--- | :--- | :--- |
| `GET` | `/api/list` | — | `NetworkOverview` | Lista consolidada e métricas |
| `POST` | `/api/add` | `{ ip, name, method, thresholdMs, uuid }` | `Result` | Adiciona e testa imediatamente |
| `POST` | `/api/remove` | `{ ip }` | `Result` | Remove o dispositivo |
| `POST` | `/api/edit` | `{ id, newIp, name, method, thresholdMs, uuid }` | `Result` | Edita e re-testa conectividade |
| `POST`/`GET` | `/api/update` | — | `NetworkOverview` | Dispara varredura paralela geral |
| `POST` | `/api/import` | Raw JSON string | `Result` | Importação em lote flexível |
| `GET` | `/api/version`| — | `{"version": "vX.Y.Z"}` | Versão da compilação |

### 7.3 Barramento de Eventos Reativos
- **Ticker em Background (1 Minuto):**
  - `main.go` dispara uma goroutine que executa `UpdateAllStatuses()` a cada 60 segundos.
  - Ao finalizar, emite:
    ```go
    app.Event.Emit("ips-updated", updatedOverview)
    ```
  - O frontend consome via `Events.On('ips-updated', ...)` e atualiza o estado sem disparar nenhuma requisição adicional.

---

## 8. Ciclo de Vida Nativo e Integração Desktop (macOS / Windows)

### 8.1 Gerenciamento de Janelas no macOS
- `ApplicationShouldTerminateAfterLastWindowClosed: false`
- **Hook `WindowClosing`:** Intercepta o clique no botão fechar (X vermelho), cancela o fechamento e executa `win.Hide()`.
- **Evento `ApplicationShouldHandleReopen`:** Reabre e coloca em foco a janela (`win.Show()`, `win.Focus()`) quando o usuário clica no ícone do aplicativo na Dock.

### 8.2 Requisitos de Segurança macOS (`Info.plist`)
O arquivo `build/darwin/Info.plist` possui as diretivas obrigatórias para chamadas de rede local no WebView:
```xml
<key>NSAppTransportSecurity</key>
<dict>
    <key>NSAllowsLocalNetworking</key>
    <true/>
    <key>NSAllowsArbitraryLoads</key>
    <true/>
</dict>
```

---

## 9. Mecanismo de Importação em Lote (JSON Parser)

O método `ImportJSONContent` suporta múltiplos formatos de schemas JSON com detecção polimórfica:
- **Array de Strings:** `["192.168.1.1", "google.com"]`
- **Array de Objetos:** `[{"url": "10.0.0.1", "name": "Servidor"}]`
- **Objetos Mapeados:** Objetos contendo chaves `"sites"`, `"ips"`, `"hosts"`, `"devices"` ou `"servers"`.
- **Mapeamento de Chaves de Host:** Suporta `"url"`, `"ip"`, `"host"`, `"address"` ou `"hostname"`.
- **Deduplicação Inteligente:** Endereços repetidos dentro do mesmo arquivo são deduplicados automaticamente mantendo a consistência no banco.
- **Trigger Assíncrono:** Ao término da importação, uma goroutine dispara automaticamente `UpdateAllStatuses()` em background para que todos os novos itens já tenham status calculado.

---

## 10. Requisitos Não-Funcionais e Métricas de Qualidade

1. **Pegada de Memória:** Consumo de RAM do backend Go inferior a **25 MB** em repouso.
2. **Tempo de Resposta em Varredura:** Uma lista de 50 hosts deve ser completamente varrida em menos de **3 segundos** graças ao pool de concorrência com 20 workers.
3. **Integridade de Compilação:** Compilável em qualquer máquina Windows moderna sem CGo (`CGO_ENABLED=0 go build`).
4. **Tamanho do Binário:** Executável consolidado (`ipMonitorApp.exe`) com tamanho médio de ~13 a 15 MB utilizando flags de redução de símbolos (`-ldflags="-s -w -H=windowsgui"`).
5. **Cobertura de Testes:** Suíte de testes unitários cobrindo sanitização, importação JSON, ordenação de prioridade e integridade do banco SQLite com `go test -v ./...`.

---

## 11. Política de Branches e Release

- **Branch `dev` (Local e Permanente):**
  - Ambiente de implementação de código, ajustes de banco e testes.
  - **Estritamente local:** proibido qualquer push da branch `dev` para o GitHub remoto. Nunca é excluída.
- **Branch `main` (Produção e Remoto):**
  - Única branch conectada com o repositório remoto.
  - Recebe commits apenas via merge expressamente validado e aprovado da branch `dev`.
  - Dispara o pipeline de CI/CD para compilação multiplataforma (Windows x64/x86, macOS Apple Silicon arm64 / Intel x86_64 e Linux).
