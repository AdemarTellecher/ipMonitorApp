# Contexto Ativo (activeContext.md)

## Foco Atual
- Manter o aplicativo atualizado utilizando o **Wails v3** (Go + Webview2 nativo)
- Garantir que o binário gerado seja **único, enxuto (~13.6MB) e 100% portátil**
- Persistência com SQLite dinâmico (`model.ResolveDBPath`) na pasta de execução

## Decisões Recentes
- **Migração do Fyne para Wails v3**: Concluída com sucesso. Eliminadas todas as dependências do Fyne e as 5 DLLs externas que eram obrigatórias no Windows.
- **Estrutura do Projeto**:
  - `main.go`: Inicialização do Wails v3 e embutimento do frontend via `//go:embed all:frontend`.
  - `internal/model/`: Persistência SQLite3 com detecção dinâmica do caminho do banco.
  - `internal/service/`: `MonitorService` com rotas de API HTTP e IPC para o frontend.
  - `frontend/`: Interface moderna e refinada construída com HTML5, CSS3 e Vanilla JS puro (sem frameworks pesados).
- **Scripts de Build**: `make.bat` e `Makefile` simplificados para compilação direta via `go build -ldflags="-s -w -H=windowsgui"`.

## Próximos Passos
- Expandir configurações adicionais conforme necessidade do usuário
- Suporte a notificações nativas ou ícone na bandeja do sistema (Systray) via Wails v3 se desejado

---

*Este arquivo faz parte do Banco de Memória do projeto. Consulte também os demais arquivos para contexto completo.*
