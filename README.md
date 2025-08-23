# MediFlow - Sistema de Gestão de Clínicas

MediFlow é uma aplicação web completa, desenvolvida em **Go** com o framework **Gin**, para a gestão de clínicas de terapia. O sistema foi projetado com uma arquitetura robusta, focada em segurança, rastreabilidade e separação de responsabilidades entre diferentes perfis de usuários.

## ✨ Funcionalidades Principais

O sistema é dividido em painéis seguros e independentes para cada perfil de usuário, além de um portal exclusivo para pacientes.

### 🔐 Segurança e Acesso

  * **Controle de Acesso por Perfil:** O sistema possui 3 níveis de acesso (Administrador, Secretária, Terapeuta), cada um com suas permissões estritamente controladas por middleware.
  * **"Soft Deletes" (Exclusão Lógica):** Nenhum usuário é permanentemente apagado do banco de dados. Em vez disso, são marcados como "inativos", preservando 100% do histórico e das relações de dados.
  * **Logs de Auditoria:** Todas as ações críticas (logins, criação de prontuários, pagamentos, exclusão de usuários, etc.) são registradas em uma tabela de auditoria, garantindo total rastreabilidade.

### 👤 Portal do Paciente

  * **Acesso Seguro por Token:** Pacientes não precisam de senha. Eles recebem um link único e seguro para acessar um portal exclusivo.
  * **Consentimento Online:** O paciente pode ler e fornecer o Termo de Consentimento diretamente pelo portal, incluindo a validação completa de CPF para garantir a autenticidade.

### 👩‍💼 Painel da Secretária

  * **Foco na Agilidade:** Desenhado para as tarefas administrativas do dia a dia.
  * **Cadastro Rápido de Pacientes:** Registra apenas as informações de contato essenciais para gerar o link do portal.
  * **Gestão da Agenda:** Visualização da agenda da clínica, com capacidade de marcar, editar e cancelar consultas.
  * **Gestão Financeira:** Visualiza o status de pagamento das consultas e possui a funcionalidade de "Marcar como Pago".
  * **Recuperação de Link:** Pode facilmente recuperar o link de consentimento para reenviar a um paciente.

### 👨‍⚕️ Painel do Terapeuta

  * **Foco Clínico e Privacidade:** O terapeuta tem acesso apenas aos seus próprios dados.
  * **Dashboard Personalizado:** Visualiza sua agenda e uma lista de seus pacientes.
  * **Acesso Seguro ao Prontuário:** Pode acessar o prontuário completo de seus pacientes para visualizar o histórico e adicionar novas anotações.
  * **Segurança:** Totalmente isolado das áreas de administração e dos dados de outros terapeutas.

### 👑 Painel do Administrador

  * **Controle Total:** Visão e controle completos sobre todos os aspectos do sistema.
  * **Gestão de Usuários e Pacientes:** CRUD (Criar, Ler, Atualizar, Desativar) completo para todos os usuários e pacientes.
  * **Dashboard de Monitoramento:** Painel com KPIs (Indicadores-Chave de Desempenho) operacionais e financeiros, incluindo:
      * Faturamento no período.
      * Valor a receber (pagamentos pendentes).
      * Novos pacientes e fontes de aquisição.
      * Estatísticas sobre o perfil dos novos pacientes.
  * **Visualização de Logs:** Acesso à tela de auditoria para monitorar todas as ações realizadas no sistema.

## 🚀 Como Executar o Projeto

Siga os passos abaixo para configurar e executar a aplicação no seu ambiente.

### Pré-requisitos

  * **Go:** Versão 1.18 ou superior.
  * **PostgreSQL:** Uma instância do PostgreSQL ativa.

### 1\. Configuração do Ambiente

Clone o repositório e navegue para a pasta raiz do projeto. Crie um arquivo chamado `.env` e preencha com suas credenciais do banco de dados:

```env
DB_TYPE=postgres
DB_HOST=localhost
DB_PORT=5432
DB_USER=seu_usuario_aqui
DB_PASS=sua_senha_aqui
DB_NAME=mediflow
PORT=8080
```

### 2\. Instalação das Dependências

Execute o comando abaixo para baixar todas as dependências:

```sh
go mod tidy
```

### 3\. Configuração do Banco de Dados

O projeto inclui um gestor de banco de dados (`data_manager.go`) para facilitar a configuração inicial.

**a. Crie o banco de dados** no PostgreSQL com o nome que especificou no `.env`.

**b. Inicialize as tabelas:** O comando a seguir irá apagar tabelas antigas (se existirem) e criar a estrutura mais recente.

```sh
go run data_manager.go -init
```

**c. Crie os usuários padrão:**

```sh
go run data_manager.go -create-users
```

**d. (Opcional) Popule com dados de teste:** Para ter dados realistas para testar, execute:

```sh
go run data_manager.go -populate
```

**e. (Opcional) Verifique os logs de auditoria** via terminal:

```sh
go run data_manager.go -audit
```

### 4\. Executar a Aplicação

Inicie o servidor web:

```sh
go run main.go
```

A aplicação estará acessível em `http://localhost:8080`.

## 🔑 Credenciais de Acesso Padrão

  * **Admin:** `admin@mediflow.com` / `senha123`
  * **Secretária:** `secretaria@mediflow.com` / `senha123`
  * **Terapeuta:** `terapeuta@mediflow.com` / `senha123`