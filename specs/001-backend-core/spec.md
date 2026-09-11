# Feature Specification: Backend Core & Resilient Multiplatform Engine

**Feature Branch**: `001-backend-core`

**Created**: 2026-09-11

**Status**: Draft

**Input**: User description: "Definição formal e especificação técnica dos requisitos do Backend em Go para o IP Monitor App com Wails v3, SQLite e comunicação reativa baseada no PRD-Backend."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Real-time Network Reachability Monitoring (Priority: P1)

Como um administrador de infraestrutura ou usuário técnico, desejo cadastrar alvos de rede (IPv4, IPv6, URLs, Hostnames) e acompanhar seu status de conectividade em tempo real com métricas detalhadas (latência atual, mínima, média, máxima e perda de pacotes), sem sobrecarregar minha máquina nem abrir janelas pretas de terminal.

**Why this priority**: Este é o núcleo de valor do IP Monitor App. Sem uma rotina de monitoramento precisa, concorrente e silenciosa, a aplicação perde sua função primária.

**Independent Test**:
Pode ser testado de forma isolada submetendo uma lista de 5 alvos heterogêneos (um IP local ativo, um IP remoto, um domínio DNS com resolução pública, um alvo inexistente e uma URL HTTP/HTTPS) e verificando se o backend atualiza as estatísticas e emite o evento `ips-updated` com latências e status válidos sem falhas ou travamento de UI.

**Acceptance Scenarios**:
1. **Given** um alvo IPv4 ativo na rede, **When** o ciclo de varredura periódica for executado, **Then** o status deve ser atualizado para `Online`, com latência calculada e taxa de perda em 0%.
2. **Given** um alvo inexistente ou bloqueado, **When** o teste de ping esgotar o timeout estipulado (ex: 1500ms), **Then** o status deve transicionar para `Offline` com `packet_loss = 100%` e registrar timestamp de última falha.
3. **Given** a execução no Windows, **When** qualquer fallback de ping do sistema for acionado, **Then** nenhuma janela do console (`cmd.exe`/`conhost.exe`) deve piscar na tela do usuário.

---

### User Story 2 - Resilient Cross-Platform SQLite Persistence (Priority: P1)

Como usuário executando a aplicação no macOS (instalado em `/Applications` ou em DMG) ou no Windows (versão portátil), desejo que minhas configurações, alvos e histórico de logs persistam entre reinicializações sem erros de permissão de escrita ou travamentos no Dock.

**Why this priority**: Erros de caminho de persistência impedem a abertura do app no macOS devido às permissões restritas do `.app/Contents/MacOS`.

**Independent Test**:
Iniciar o backend tanto simulando o ambiente macOS sandbox quanto em modo portátil no Windows; certificar-se de que o diretório apropriado seja criado (`~/Library/Application Support/IPMonitor/` no Darwin ou diretório do executável no Windows) e as tabelas `hosts`, `settings` e `logs` sejam criadas e consultadas sem erros de `readonly database`.

**Acceptance Scenarios**:
1. **Given** execução em macOS (`runtime.GOOS == "darwin"`), **When** o banco SQLite é inicializado, **Then** o arquivo `.db` deve ser criado sob `~/Library/Application Support/IPMonitor/` com permissão `0755` para a pasta.
2. **Given** concorrência de escrita de métricas e leitura pela UI, **When** múltiplas goroutines acessam o banco, **Then** o modo WAL (`journal_mode=WAL`) e `busy_timeout=5000` devem garantir consistência sem erros de `database is locked`.

---

### User Story 3 - Reactive Event Dispatching & Dual Service Surface (Priority: P2)

Como engenheiro de frontend integrando a UI do Wails v3, desejo receber dados prontos para exibição via barramento reativo (`ips-updated`) e dispor de métodos nativos ou endpoints REST (`/api/...`) para CRUD de alvos e disparo manual de testes, sem necessidade de consultas contínuas via polling.

**Why this priority**: Assegura a separação estrita de responsabilidades ("Go owns data and business rules; Dumb UI") e fluidez na interface.

**Independent Test**:
Cadastrar um novo alvo via endpoint/binding e confirmar que o backend emite imediatamente o evento de atualização sem que o frontend execute nenhum `setInterval` ou loop de polling.

**Acceptance Scenarios**:
1. **Given** a conclusão de uma rodada de varreduras, **When** o serviço compilar o sumário de rede, **Then** deve emitir o evento nativo Wails `ips-updated` contendo a lista ordenada e métricas globais consolidadas.
2. **Given** uma requisição para adicionar um alvo inválido (ex: string não resolúvel), **When** processada pelo backend, **Then** deve retornar erro de validação sem alterar a lista ativa.

---

### User Story 4 - Multi-level Hybrid Connectivity Fallback (Priority: P2)

Como usuário em redes corporativas com políticas restritas de firewall (onde pacotes ICMP diretos são descartados ou filtrados), desejo que o aplicativo continue diagnosticando a saúde do host através de sondas TCP em portas comumente abertas.

**Why this priority**: Evita falsos-negativos (alvos marcados como `Offline` quando o serviço está operacional, porém ICMP está desabilitado na rede).

**Independent Test**:
Configurar um alvo que responda a HTTP na porta 80 ou HTTPS na porta 443, mas que bloqueie ICMP Echo Request. O sistema deve acionar o fallback TCP e confirmar o status `Online (TCP)`.

**Acceptance Scenarios**:
1. **Given** um alvo com ICMP bloqueado mas porta 80/443 acessível, **When** o ping ICMP falhar por timeout, **Then** o motor deve efetuar teste TCP com timeout curto (ex: 800ms) nas portas configuradas e atualizar o status adequadamente.

---

### Edge Cases

- O que acontece se o computador suspender/hibernar durante a execução periódica de varredura? As goroutines pendentes devem ser canceladas via contexto com timeout e a varredura reiniciada limpa após o retorno.
- O que acontece se a interface de rede for desconectada abruptamente? As requisições de DNS e sockets falharão rapidamente com `network unreachable`, marcando os alvos como inalcançáveis sem travar a aplicação.
- O que acontece quando há centenas de alvos cadastrados? A varredura deve utilizar um worker pool limitado (ex: concorrência máxima entre 20 a 50 workers) para não esgotar descritores de arquivo ou recursos do sistema operacional.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: O sistema DEVE persistir e gerenciar a lista de alvos de rede (`Host`) com campos de identificação, endereço (IP ou hostname), status de conectividade, métricas de latência e histórico recente.
- **FR-002**: O sistema DEVE fornecer resolução dinâmica e resiliente do caminho do banco SQLite (`ResolveDBPath`), garantindo que no macOS o banco fique em `~/Library/Application Support/IPMonitor/` e no Windows/Linux na pasta do executável.
- **FR-003**: O sistema DEVE rodar em Go puro com `CGO_ENABLED=0` utilizando exclusivamente o driver `modernc.org/sqlite`.
- **FR-004**: O sistema DEVE executar varreduras periódicas e pontuais de forma concorrente em segundo plano com cancelamento gracioso via `context.Context`.
- **FR-005**: O sistema DEVE implementar pipeline de teste de conectividade híbrido: ICMP com adaptação de privilégio por SO (`SetPrivileged(true)` no Windows e `SetPrivileged(false)` no Darwin/Linux), fallback via comando nativo com processo oculto (`CREATE_NO_WINDOW`) no Windows e fallback de portas TCP.
- **FR-006**: O sistema DEVE emitir eventos no barramento do Wails v3 (`app.Event.Emit("ips-updated", payload)`) após cada ciclo de verificação para atualizar a UI sem necessidade de polling.
- **FR-007**: O sistema DEVE expor uma API REST local (`/api/hosts`, `/api/ping`, `/api/settings`) e bindings nativos do Wails v3 para consumo pelo WebView.
- **FR-008**: O sistema DEVE controlar o ciclo de vida da janela no macOS interceptando o fechamento para ocultar (`win.Hide()`) em vez de encerrar a aplicação, restaurando ao clicar no ícone do Dock.

### Key Entities

- **Host (Alvo)**: Representa um destino de monitoramento. Contém ID, Label/Nome, Endereço (IPv4/IPv6/URL/Hostname), Porta opcional, Status atual (`online`, `offline`, `degraded`), Latência atual, Latência média, Pacotes transmitidos, Pacotes recebidos, Taxa de perda (`packet_loss`), Data do último sucesso e Data da última falha.
- **NetworkOverview (Sumário de Rede)**: Objeto agregado pronto para consumo pela UI contendo contadores de alvos (total, online, offline), latência média global, taxa de disponibilidade e lista ordenada de hosts.
- **AppSettings (Configurações)**: Entidade que armazena intervalo de varredura periódica (segundos), timeout de socket, concorrência máxima de workers e opções de notificação.
- **AuditLog / PingLog**: Registro temporal de transições de status e eventos de conectividade.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: O aplicativo deve inicializar e abrir a tela principal em menos de 1,5 segundos em sistemas operacionais suportados.
- **SC-002**: A varredura de até 50 alvos simultâneos deve ser concluída em menos de 3 segundos em conexões normais sem degradar a responsividade da UI.
- **SC-003**: Zero consumo de CGO: a aplicação deve compilar com sucesso com `CGO_ENABLED=0` para Windows (`amd64`), macOS (`darwin/arm64`, `darwin/amd64`) e Linux (`amd64`).
- **SC-004**: Zero falhas de permissão de escrita no macOS quando instalado no diretório `/Applications`.
- **SC-005**: Zero janelas visíveis ou piscantes de prompt de comando durante testes de rede no Windows.

## Assumptions

- O usuário possui permissões para executar aplicações de rede local no sistema operacional.
- O ambiente de compilação utiliza Go 1.25+ e Wails v3 CLI/SDK.
- A máquina de destino possui conectividade com a rede local ou internet conforme os alvos cadastrados.
