# IP Monitor App

Aplicativo moderno, ultra-leve e portátil em Go para monitoramento de IPs e conectividade com interface nativa [Wails v3](https://v3.wails.io) e persistência local em SQLite3.

## Funcionalidades
- **Executável Único e Portátil**: Frontend e backend 100% integrados em um binário autocontido (`ipMonitorApp.exe` de ~13 MB).
- **Zero DLLs ou Pastas Externas**: Funciona sobre o WebView2 nativo do Windows, sem necessidade de DLLs soltas na pasta.
- **SQLite Dinâmico**: Detecta automaticamente a pasta de execução e cria/utiliza o banco `ipmonitor.db` localmente.
- **Importação de JSON**: Importação em lote a partir de arquivos `.json` exportados, com validação e sanitização automática de hosts/URLs.
- **Varredura Ativa e Periódica**: Monitoramento com ICMP Ping em tempo real e atualização periódica em background.
- **Temas Modernos**: Alternância instantânea entre Tema Escuro e Tema Claro com alto contraste.
- **Ações Rápidas**: Adicionar, Atualizar Todos, Importar JSON e Remover hosts selecionados.

## Requisitos
- Go 1.25 ou superior
- Compilador C (GCC / MinGW-w64) no Windows (necessário para compilar o driver CGO do SQLite)
- WebView2 Runtime (nativo no Windows 10 e Windows 11)

## Compilação e Build

O projeto utiliza um **Makefile unificado** que detecta automaticamente o sistema operacional da máquina (Windows, Linux ou macOS) e ajusta as flags de compilação, o nome do executável e as ferramentas necessárias:

```sh
make build   # Detecta o SO e compila o binário único e portátil
make test    # Executa a suíte de testes unitários
make run     # Compila e executa o aplicativo
make clean   # Remove o executável gerado
make info    # Exibe o sistema operacional detectado e configurações de build
```

> **No Windows**: Se você usa o MinGW-w64 (WinLibs), você pode executar diretamente `mingw32-make` ou `make`. O Makefile já configura automaticamente os caminhos do GCC e as variáveis `CGO_ENABLED=1` e flags para ocultar o console (`-H=windowsgui`).
>
> **No Linux/macOS**: O executável é gerado como `ipMonitorApp` e sem terminal acoplado extra.

## Como usar

1. **Adicionar IP:**
   - Digite o endereço IP no campo e clique em "Adicionar" (ou pressione Enter).
2. **Atualizar Status:**
   - Clique em "Atualizar Todos" para disparar uma nova verificação via ICMP Ping.
3. **Importar JSON:**
   - Clique em "Importar JSON" e selecione o arquivo `.json` desejado.
4. **Remover IP:**
   - Clique em uma linha da tabela para selecionar o host e clique em "Remover".
5. **Alternar Tema:**
   - Clique no botão no canto superior direito para alternar entre os temas Claro e Escuro.

## Banco de Dados e Portabilidade
- O arquivo `ipmonitor.db` é resolvido dinamicamente no diretório onde o executável foi iniciado.
- Você pode mover o `ipMonitorApp.exe` para qualquer pasta ou pendrive; o banco de dados acompanhará a pasta de execução.

---

Desenvolvido com Go, Wails v3 e SQLite3.
