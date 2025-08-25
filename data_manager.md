# Gerenciador de Banco de Dados (data_manager.go)

Este script é uma ferramenta de linha de comando (CLI) para gerenciar o banco de dados da aplicação MediFlow. Ele permite inicializar tabelas, popular dados de teste e realizar tarefas administrativas de usuários sem a necessidade de acessar a interface web ou um cliente de banco de dados.

## Uso Básico

Todos os comandos devem ser executados a partir da pasta raiz do projeto. O formato geral do comando é:

```

go run data_manager.go [ação] [parâmetros\_de\_conexão]

```

## Configuração via .env

A maneira mais fácil de configurar a conexão é criando um arquivo chamado `.env` na raiz do projeto. O script irá carregar essas variáveis automaticamente. Se as flags de linha de comando forem usadas, elas terão prioridade sobre o `.env`.

**Exemplo de arquivo `.env`:**
```

DB\_TYPE=postgres
DB\_HOST=localhost
DB\_PORT=5432
DB\_USER=seu\_usuario\_aqui
DB\_PASS=sua\_senha\_aqui
DB\_NAME=mediflow

```

## Flags de Ação

Você deve especificar pelo menos uma das seguintes flags para executar uma tarefa:

| Flag                   | Descrição                                                                        |
| ---------------------- | ---------------------------------------------------------------------------------- |
| `-init`                | Apaga todas as tabelas existentes e recria a estrutura do zero.                    |
| `-create-users`        | Insere os 3 usuários padrão (admin, secretaria, terapeuta).                        |
| `-populate`            | Preenche o banco com 45 pacientes de teste com dados clínicos detalhados.          |
| `-create-single-user`  | Cria um único usuário. Requer as flags `-user-name`, `-user-email`, etc.           |
| `-delete-user-by-email`| Deleta permanentemente um usuário pelo seu email. Ex: `-delete-user-by-email "user@email.com"` |
| `-list-users`          | Exibe uma lista de todos os usuários ativos no banco de dados.                     |
| `-create-test-db`      | Cria um banco de dados de teste separado com um nome único (ex: `mediflow_test_DD_MM_YY`). |
| `-audit`               | Exibe os últimos 100 logs de auditoria do sistema no terminal.                     |
| `-deleted-patients`    | Exibe uma lista de todos os pacientes que foram removidos (soft delete).           |

## Flags de Configuração

Use estas flags para especificar os detalhes do banco de dados ou de um novo usuário. Elas sobrescrevem qualquer valor existente no arquivo `.env`.

#### Conexão com o Banco de Dados

| Flag      | Descrição                           |
| --------- | ----------------------------------- |
| `-dbhost` | Endereço do servidor PostgreSQL.    |
| `-dbport` | Porta do servidor PostgreSQL.       |
| `-dbuser` | Nome de usuário do PostgreSQL.      |
| `-dbpass` | Senha do usuário do PostgreSQL.     |
| `-dbname` | Nome do banco de dados a ser usado. |

#### Detalhes para Criação de Usuário

| Flag            | Descrição                                         |
| --------------- | --------------------------------------------------- |
| `-user-name`    | Nome completo do novo usuário.                      |
| `-user-email`   | Email de login do novo usuário.                     |
| `-user-password`| Senha do novo usuário.                              |
| `-user-role`    | Perfil do usuário (`admin`, `secretaria`, `terapeuta`). |

## Exemplos de Comandos

### 1. Preparar um banco de dados do zero (usando `.env`):

```

go run data_manager.go -init -create-users -populate

```

### 2. Criar um novo usuário administrador (usando `.env`):

```

go run data_manager.go -create-single-user  
\-user-name "Novo Admin"  
\-user-email "novo.admin@mediflow.com"  
\-user-password "senhaForte123"  
\-user-role admin

```

### 3. Deletar um usuário pelo email (usando `.env`):

```

go run data_manager.go -delete-user-by-email "novo.admin@mediflow.com"

```

### 4. Listar todos os usuários existentes (usando `.env`):

```

go run data_manager.go -list-users

```

### 5. Criar um usuário **sem .env**, especificando tudo na linha de comando:

```

go run data_manager.go -create-single-user  
\-user-name "Lauro Gomes"  
\-user-email "lauro@exemplo.com"  
\-user-password "senha123"  
\-user-role admin  
\-dbhost localhost  
\-dbport 5432  
\-dbuser me  
\-dbpass 1q2w3e  
\-dbname mediflow

```
