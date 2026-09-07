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

### Windows (`make.bat`)
Para compilar o aplicativo no Windows em modo de produção (sem janela de terminal aberta e com símbolos otimizados):

```bat
make.bat
```

Isso gerará o executável único `ipMonitorApp.exe` na raiz do projeto.

### Makefile
Alternativamente, se tiver o `make` instalado:
```sh
make build   # Compila o executável ipMonitorApp.exe
make test    # Executa a suíte de testes unitários
make run     # Compila e executa o app
make clean   # Remove o executável gerado
```

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
