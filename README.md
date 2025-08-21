# MediFlow - Sistema de Gestão de Consultas

MediFlow é uma aplicação web desenvolvida em Go com o framework Gin para a gestão de agendamentos e prontuários de pacientes para clínicas de psicologia. O sistema foi projetado para ser utilizado por diferentes perfis de utilizadores (Administrador, Secretária e Médico), cada um com os seus próprios níveis de acesso e funcionalidades.

## ✨ Funcionalidades Principais

- **Gestão de Acessos por Perfis:**
  - **Administrador:** Acesso total ao sistema, incluindo gestão de utilizadores, gestão de pacientes, agendamentos e painel de monitoramento com estatísticas.
  - **Secretária:** Focada na operação diária. Pode registar novos pacientes (apenas dados de identificação), consultar a lista de pacientes e gerir a agenda de consultas (marcar, desmarcar, editar). Não tem acesso a dados clínicos sensíveis.
  - **Médico:** (A ser implementado) Terá acesso ao perfil completo dos seus pacientes, incluindo histórico clínico, e poderá gerir as suas próprias consultas.

- **Gestão de Pacientes (CRUD Completo para Admin):**
  - Visualização de todos os pacientes com paginação e busca.
  - Busca com autopreenchimento para encontrar pacientes rapidamente.
  - Formulário completo para adicionar e editar todos os dados do paciente, incluindo informações de perfil, histórico de saúde e termo de consentimento.

- **Agenda Visual e Interativa (para Secretária):**
  - Dashboard principal com uma visualização de agenda semanal.
  - Navegação temporal para semanas anteriores e futuras.
  - Links diretos dos agendamentos para o perfil do paciente, facilitando a remarcação.
  - Funcionalidade para marcar, editar e desmarcar (cancelar) consultas.

- **Painel de Monitoramento (para Admin):**
  - Dashboard com estatísticas sobre a operação da clínica.
  - Filtros para visualizar dados dos últimos 7, 15 ou 30 dias.
  - Métricas sobre consultas concluídas, novos pacientes, fontes de aquisição e perfil dos pacientes.
  - Visualização das próximas consultas agendadas.

## 🚀 Como Executar o Projeto

Siga os passos abaixo para configurar e executar a aplicação no seu ambiente de desenvolvimento.

### Pré-requisitos

- **Go:** Versão 1.18 ou superior.
- **PostgreSQL:** Uma instância do PostgreSQL a correr localmente ou acessível pela rede.

### 1. Configuração do Ambiente

Clone o repositório (ou certifique-se de que tem todos os ficheiros na estrutura correta) e navegue para a pasta raiz do projeto.

Crie um ficheiro chamado `.env` na raiz do projeto e adicione as suas credenciais de acesso ao banco de dados. Pode usar o exemplo abaixo como base:

```env
DB_TYPE=postgres
DB_HOST=localhost
DB_PORT=5432
DB_USER=seu_usuario_aqui
DB_PASS=sua_senha_aqui
DB_NAME=mediflow
PORT=8080
````

### 2\. Instalação das Dependências

Abra um terminal na raiz do projeto e execute o comando abaixo para descarregar todas as dependências necessárias:

```sh
go mod tidy
```

### 3\. Configuração do Banco de Dados

O projeto inclui um gestor de banco de dados para facilitar a configuração inicial.

**a. Crie o banco de dados:** Certifique-se de que criou uma base de dados no PostgreSQL com o nome que especificou no seu ficheiro `.env` (ex: `mediflow`).

**b. Inicialize as tabelas:** Execute o comando abaixo para criar todas as tabelas necessárias.

```sh
go run ./data_manager.go -init
```

**c. Crie os utilizadores padrão:** Este comando irá criar os três perfis de utilizador para que possa aceder ao sistema.

```sh
go run ./data_manager.go -create-users
```

**d. (Opcional) Povoe com dados de teste:** Para testar a aplicação com dados realistas, execute o comando abaixo. Ele irá criar 45 pacientes e as suas respetivas consultas passadas e futuras.

```sh
go run ./data_manager.go -populate
```

### 4\. Executar a Aplicação

Depois de configurar o banco de dados, inicie o servidor web:

```sh
go run main.go
```

O servidor estará a correr e acessível no endereço `http://localhost:8080` (ou na porta que especificou no ficheiro `.env`).

### Credenciais de Acesso Padrão

  - **Admin:** `admin@mediflow.com` / `senha123`
  - **Secretária:** `secretaria@mediflow.com` / `senha123`
  - **Médico:** `medico@mediflow.com` / `senha123`

