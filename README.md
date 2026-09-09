# IP Monitor App

Aplicativo moderno, ultra-leve e portátil em Go para monitoramento de IPs e conectividade de rede em tempo real, construído com interface nativa em [Wails v3](https://v3.wails.io) e persistência local com SQLite3.

Inspirado no design minimalista e moderno do Microsoft PC Manager, o **IP Monitor App** entrega uma experiência de alto padrão visual com temas Claro e Escuro, animações fluidas e zero dependências externas.

---

## 📸 Interface do Aplicativo

<div align="center">
  <table>
    <tr>
      <td align="center"><b>Tema Escuro (Dark Mode)</b></td>
      <td align="center"><b>Tema Claro (Light Mode)</b></td>
    </tr>
    <tr>
      <td><img src="internal/assets/screenshot-dark.png" alt="IP Monitor - Tema Escuro" width="370" /></td>
      <td><img src="internal/assets/screenshot-light.png" alt="IP Monitor - Tema Claro" width="370" /></td>
    </tr>
  </table>
</div>

---

## ✨ Funcionalidades Principais

- **Executável Único e Portátil**: Frontend e backend 100% integrados em um binário autocontido (`ipMonitorApp.exe` de ~13 MB).
- **Zero DLLs ou Pastas Externas**: Utiliza o WebView2 nativo do Windows, sem necessidade de arquivos `.dll` adicionais na pasta.
- **SQLite Dinâmico e Portátil**: Detecta automaticamente a pasta de execução e cria/utiliza o banco `ipmonitor.db` localmente com migração transparente de schema.
- **Painel Expansível de Detalhes**: Cada cartão possui uma seta (chevron) que expande uma gaveta com informações completas:
  - **Nome / Identificação** amigável (ex: *RADAR-PA255 - KM3*)
  - **Método de Teste** (ICMP Ping / TCP)
  - **Limite de Timeout** configurado (ms)
  - **ID do Sistema / UUID** (gerado ou importado)
- **Filtro Inteligente de Prioridade**: Hosts **Offline** aparecem automaticamente no topo da lista para atenção e diagnóstico imediatos.
- **Identificação Visual Suave**: Indicadores de status refinados com animação de pulso e destaque visual em tom vermelho-alaranjado para hosts inacessíveis.
- **Cadastro e Edição com Formulário Completo**: Diálogo modal centralizado e bloqueante com suporte a todos os metadados do host.
- **Barra Lateral com Ações Rápidas**:
  - **Cadastrar**: Abertura rápida do modal para inclusão de novos hosts com campos de IP, Nome, Método e Timeout.
  - **Importar**: Leitura em lote de arquivos `.json` estruturados (ex: sites/hosts com `name`, `url`/`ip`, `method`, `thresholdMs` e `id`).
  - **Varredura**: Disparo manual de ICMP Ping imediato com animação rotativa no ícone e feedback visual no status.
  - **Editar**: Edição rápida do host selecionado via botão na barra lateral ou **duplo clique** diretamente na linha da tabela.
  - **Remover**: Exclusão de host selecionado com confirmação imediata.
- **Varredura Automática em Background**: Verificação periódica a cada 1 minuto mantendo o painel de disponibilidade sempre atualizado.
- **Tipografia Confortável & Temas Claro e Escuro**: Fontes escalonadas para leitura ergonômica e alternância instantânea com contraste polido.

---

## 🛠️ Requisitos de Desenvolvimento

- **Go 1.25** ou superior
- **Go 100% Puro no Windows**: Não requer GCC, MinGW-w64, WinLibs ou CGo para compilar.
- **WebView2 Runtime** (já incluído nativamente no Windows 10 e Windows 11)

---

## 🚀 Compilação e Execução

O projeto conta com um **Makefile unificado** multiplataforma que detecta automaticamente o sistema operacional (Windows, Linux ou macOS):

```sh
make build   # Compila o executável único e portátil (CGO_ENABLED=0 no Windows)
make test    # Executa todos os testes unitários do serviço
make run     # Compila e inicia o aplicativo imediatamente
make clean   # Remove executáveis e resíduos temporários de build
make info    # Exibe dados do ambiente, SO detectado e flags de compilação
```

> 💡 **Dica no Windows**: Você pode compilar diretamente com `make build` ou rodar nativamente `go build -ldflags="-s -w -H=windowsgui" -o ipMonitorApp.exe .` em qualquer terminal (PowerShell, CMD ou Git Bash) **sem nenhuma ferramenta C instalada**!

---

## 📖 Como Usar

1. **Cadastrar Dispositivo**:
   - Clique em **Cadastrar** na barra lateral. O modal central bloqueante será aberto.
   - Preencha o **Endereço IP / Hostname** (obrigatório), **Nome / Identificação** (opcional), escolha o **Método de Teste** (PING ou TCP) e o **Timeout Limite (ms)**.
   - Pressione `Enter` ou clique em **Cadastrar**.
2. **Visualizar Detalhes Expandidos**:
   - Clique na **seta (chevron)** ao lado do endereço IP para abrir ou recolher a gaveta de detalhes do host (Nome, Método, Timeout e UUID).
3. **Editar Host**:
   - Selecione a linha do host desejado e clique no botão **Editar** na barra lateral (ou dê um **duplo clique rápido** diretamente no cartão).
   - O modal abrirá pré-preenchido com todos os dados atuais do dispositivo para alteração.
4. **Realizar Varredura Manual**:
   - Clique no botão **Varredura** na barra lateral. O ícone girará enquanto os pings ICMP são emitidos e a lista e os cards de disponibilidade serão atualizados.
5. **Importar Lista em Lote (JSON)**:
   - Clique em **Importar** na barra lateral e selecione o arquivo `.json`.
   - O aplicativo aceita tanto listas simples quanto formatos estruturados com objetos ricos (`name`, `url`, `method`, `thresholdMs`, `id`):
     ```json
     {
       "sites": [
         {
           "name": "RADAR-PA255 - KM3 - Monte Alegre",
           "url": "168.195.153.219",
           "method": "PING",
           "thresholdMs": 2000,
           "id": "B0CD4840-147C-4027-BAB0-87B74A6FE8F5"
         }
       ]
     }
     ```
6. **Remover Host**:
   - Selecione a linha do host desejado na lista e clique no botão **Remover** na barra lateral.
7. **Alternar Tema**:
   - Clique no botão **Tema** no canto superior direito para alternar dinamicamente entre os modos Escuro e Claro.

---

## 💾 Banco de Dados Local

O banco de dados SQLite (`ipmonitor.db`) é resolvido dinamicamente no diretório onde o executável se encontra. O aplicativo pode ser copiado para qualquer diretório ou pendrive sem perder suas configurações e histórico. Novas colunas de metadados (`name`, `method`, `threshold_ms`, `uuid`) são migradas automaticamente na inicialização.

---

Desenvolvido com **Go**, **Wails v3** e **SQLite3**.
