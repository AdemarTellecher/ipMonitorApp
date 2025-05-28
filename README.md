# IP Monitor App

Aplicativo multiplataforma em Go para monitoramento de IPs (online/offline) com interface gráfica (Fyne) e persistência local em SQLite3.

## Funcionalidades
- Cadastro de IPs para monitoramento
- Listagem dinâmica dos IPs monitorados
- Remoção de IPs
- Atualização do status (online/offline) manual
- Interface gráfica moderna baseada em cartões/seções
- Persistência automática em banco SQLite3 local
- Validação de IP na inclusão

## Requisitos
- Go 1.18 ou superior (recomendado Go 1.20+)
- Git (opcional, para clonar o repositório)
- Sistema operacional: Windows, Linux ou macOS

## Instalação

1. **Clone o repositório (opcional):**
   ```sh
   git clone <url-do-repositorio>
   cd ipMonitorApp
   ```

2. **Baixe as dependências:**
   ```sh
   go mod tidy
   ```

3. **Compile e execute:**
   ```sh
   go run main.go
   ```
   Ou para gerar o executável:
   ```sh
   go build -o ipmonitorapp.exe main.go
   ./ipmonitorapp.exe
   ```
   (No Linux/macOS, use `./ipmonitorapp`)

## Como usar

1. **Adicionar IP:**
   - Digite o endereço IP no campo "Digite o IP a ser monitorado".
   - Clique em "Adicionar IP".
   - Apenas IPs válidos são aceitos.

2. **Atualizar status:**
   - Clique em "Atualizar Status" para verificar se os IPs estão online ou offline.

3. **Remover IP:**
   - Selecione um IP na lista.
   - Clique em "Remover IP".

4. **Resumo:**
   - O topo da interface mostra o total de IPs cadastrados e o status geral (online/offline).

## Banco de Dados
- O arquivo `ipmonitor.db` é criado automaticamente na primeira execução, no mesmo diretório do aplicativo.
- Não é necessário configurar nada manualmente.

## Observações
- O status "Online" é determinado por uma tentativa de conexão TCP na porta 80 do IP.
- O app não faz atualização automática em background (apenas manual).
- O app aceita tanto IPv4 quanto IPv6 válidos.

## Personalização e Extensão
- O código segue o padrão MVC (Model-View-Controller) para facilitar manutenção e evolução.
- Para internacionalização, melhorias visuais ou automação, contribua ou solicite via issues.

## Suporte
- Dúvidas, sugestões ou bugs: abra uma issue no repositório ou entre em contato com o desenvolvedor.

---

Desenvolvido com Go, Fyne e SQLite3.
