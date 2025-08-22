package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

// Versão Final e Completa do Schema
var createTableSQL = `
DROP TABLE IF EXISTS consultation_summaries, appointments, patient_records, patients, users CASCADE;

CREATE TABLE IF NOT EXISTS users (
  id SERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  email VARCHAR(255) UNIQUE NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  user_type VARCHAR(50) NOT NULL CHECK (user_type IN ('terapeuta', 'secretaria', 'admin')),
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS patients (
    id SERIAL PRIMARY KEY,
    access_token VARCHAR(64) UNIQUE,
    consent_given_at TIMESTAMP WITH TIME ZONE,	
    consent_date DATE, consent_name VARCHAR(255), consent_cpf_rg VARCHAR(50),
    signature_date DATE, signature_location VARCHAR(255), name VARCHAR(255),
    address_street VARCHAR(255), address_number VARCHAR(50), address_neighborhood VARCHAR(255),
    address_city VARCHAR(255), address_state VARCHAR(50), phone VARCHAR(50), mobile VARCHAR(50),
    dob DATE, age INT, gender VARCHAR(50), marital_status VARCHAR(255), children VARCHAR(10),
    num_children INT, profession VARCHAR(255), email VARCHAR(255), emergency_contact VARCHAR(255),
    emergency_phone VARCHAR(50), emergency_other VARCHAR(255), repetitive_effort TEXT,
    physical_activity TEXT, smoker TEXT, alcohol TEXT, mental_disorder TEXT,
    mental_disorder_treatment TEXT, mental_disorder_details TEXT, religion TEXT,
    religion_details TEXT, medication TEXT, medication_details TEXT, surgery TEXT,
    surgery_details TEXT, allergies TEXT, allergies_details TEXT, anxiety_level INT,
    anger_level INT, fear_level INT, sadness_level INT, joy_level INT, energy_level INT,
    main_complaint TEXT, complaint_history TEXT, signs_symptoms TEXT, current_treatment TEXT,
    how_found VARCHAR(255), referral_name VARCHAR(255), other_source VARCHAR(255), notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS patient_records (
    id SERIAL PRIMARY KEY, patient_id INT NOT NULL REFERENCES patients(id) ON DELETE CASCADE,
    doctor_id INT NOT NULL REFERENCES users(id), record_date TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    anxiety_level INT, anger_level INT, fear_level INT, sadness_level INT, joy_level INT,
    energy_level INT, main_complaint TEXT, complaint_history TEXT, signs_symptoms TEXT,
    current_treatment TEXT, notes TEXT, created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS appointments (
  id SERIAL PRIMARY KEY, patient_id INT NOT NULL REFERENCES patients(id) ON DELETE CASCADE,
  doctor_id INT NOT NULL REFERENCES users(id), start_time TIMESTAMP WITH TIME ZONE NOT NULL,
  end_time TIMESTAMP WITH TIME ZONE NOT NULL, notes TEXT,
  status VARCHAR(50) NOT NULL CHECK (status IN ('agendado', 'concluido', 'cancelado')),
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
`

func main() {
	initDB := flag.Bool("init", false, "Inicializa o banco de dados e cria tabelas.")
	createUsers := flag.Bool("create-users", false, "Cria usuários de exemplo.")
	populateDB := flag.Bool("populate", false, "Popula o banco com dados de teste.")
	flag.Parse()

	if err := godotenv.Load(); err != nil {
		log.Fatalf("Erro ao carregar o arquivo .env: %v", err)
	}
	db, err := newDBConnection()
	if err != nil {
		log.Fatalf("Falha ao conectar ao banco de dados: %v", err)
	}
	defer db.Close()

	if *initDB {
		fmt.Println("Inicializando o banco de dados...")
		if err := initializeDB(db); err != nil {
			log.Fatalf("Erro ao inicializar o banco de dados: %v", err)
		}
		fmt.Println("Banco de dados inicializado com sucesso!")
	}
	if *createUsers {
		fmt.Println("Criando usuários de exemplo...")
		if err := createDefaultUsers(db); err != nil {
			log.Fatalf("Erro ao criar usuários: %v", err)
		}
		fmt.Println("Usuários de exemplo criados com sucesso!")
	}
	if *populateDB {
		fmt.Println("Populando o banco de dados com 45 pacientes...")
		if err := populateData(db, 45); err != nil {
			log.Fatalf("Erro ao popular o banco de dados: %v", err)
		}
		fmt.Println("Dados de teste inseridos com sucesso!")
	}
	if !*initDB && !*createUsers && !*populateDB {
		fmt.Println("Use uma flag para executar uma ação: -init, -create-users, ou -populate.")
	}
}

// data_manager.go

// data_manager.go

// FUNÇÃO PARA POPULAR O BANCO COM DADOS RICOS E COMPLETOS (VERSÃO FINAL)
func populateData(db *sql.DB, patientCount int) error {
	rand.Seed(time.Now().UnixNano())

	rows, err := db.Query("SELECT id FROM users WHERE user_type = 'terapeuta'")
	if err != nil {
		return fmt.Errorf("falha ao buscar médicos: %w", err)
	}
	defer rows.Close()

	var doctorIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return err
		}
		doctorIDs = append(doctorIDs, id)
	}
	if len(doctorIDs) == 0 {
		return fmt.Errorf("nenhum médico encontrado para associar aos pacientes")
	}

	// Listas para geração de dados aleatórios
	nomes := []string{"Alice", "Bernardo", "Clara", "Davi", "Elena", "Felipe", "Giovanna", "Heitor", "Isadora", "Júlio"}
	sobrenomes := []string{"Andrade", "Barbosa", "Cardoso", "Dias", "Esteves", "Fernandes", "Gusmão", "Henriques", "Jesus", "Lopes"}
	cidades := []string{"São Paulo", "Rio de Janeiro", "Belo Horizonte", "Salvador", "Fortaleza"}
	estados := []string{"SP", "RJ", "MG", "BA", "CE"}
	profissoes := []string{"Engenheiro(a)", "Advogado(a)", "Professor(a)", "Médico(a)", "Designer", "Desenvolvedor(a)"}
	howFoundOptions := []string{"Instagram", "Google", "Indicação de conhecido", "Outro"}
	simNao := []string{"Sim", "Não"}
	
	// DADOS ADICIONADOS PARA COMPLETAR O HISTÓRICO
	queixas := []string{"Ansiedade constante no trabalho.", "Episódios de tristeza profunda.", "Dificuldades para dormir.", "Ataques de pânico.", "Falta de motivação geral."}
	historicosQueixa := []string{"Sintomas começaram há cerca de 6 meses, após um período de grande estresse profissional.", "Paciente relata que se sente assim desde a adolescência, mas piorou nos últimos meses.", "Não consegue identificar um gatilho específico, mas a sensação é persistente."}
	sinaisSintomas := []string{"Insônia, falta de apetite e irritabilidade.", "Apatia, choro fácil e isolamento social.", "Coração acelerado, sudorese e sensação de perigo iminente."}
	tratamentosAtuais := []string{"Faz uso de Rivotril ocasionalmente.", "Nenhum tratamento medicamentoso no momento.", "Iniciou terapia cognitivo-comportamental há 2 semanas."}
	notasMedico := []string{"Paciente demonstra boa capacidade de insight, mas dificuldade em aplicar técnicas de relaxamento.", "Apresenta resistência em falar sobre o passado familiar. Focar em construir confiança.", "Sessão produtiva. Paciente conseguiu identificar padrões de pensamento negativos."}


	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	patientStmt, err := tx.Prepare(`
		INSERT INTO patients (
			name, email, age, dob, phone, mobile, profession, address_city, address_state,
			how_found, physical_activity, smoker, alcohol, repetitive_effort, mental_disorder, medication,
			anxiety_level, anger_level, fear_level, sadness_level, joy_level, energy_level, 
			main_complaint, complaint_history, signs_symptoms, current_treatment, notes
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27)
		RETURNING id`)
	if err != nil {
		return err
	}
	defer patientStmt.Close()

	appointmentStmt, err := tx.Prepare(`INSERT INTO appointments (patient_id, doctor_id, start_time, end_time, status) VALUES ($1, $2, $3, $4, $5)`)
	if err != nil {
		return err
	}
	defer appointmentStmt.Close()

	for i := 0; i < patientCount; i++ {
		nomeCompleto := fmt.Sprintf("%s %s", nomes[rand.Intn(len(nomes))], sobrenomes[rand.Intn(len(sobrenomes))])
		email := fmt.Sprintf("paciente.%d@example.com", i+1)
		idade := 20 + rand.Intn(40)
		dataNascimento := time.Now().AddDate(-idade, rand.Intn(12), rand.Intn(28))
		cidadeIdx := rand.Intn(len(cidades))
		telefone := fmt.Sprintf("(%d) 9%d-%d", 11+rand.Intn(88), 8000+rand.Intn(1999), 1000+rand.Intn(8999))

		var patientID int
		err := patientStmt.QueryRow(
			nomeCompleto, email, idade, dataNascimento, telefone, telefone,
			profissoes[rand.Intn(len(profissoes))],
			cidades[cidadeIdx],
			estados[cidadeIdx],
			howFoundOptions[rand.Intn(len(howFoundOptions))],
			simNao[rand.Intn(len(simNao))], // physical_activity
			simNao[rand.Intn(len(simNao))], // smoker
			simNao[rand.Intn(len(simNao))], // alcohol
			simNao[rand.Intn(len(simNao))], // repetitive_effort
			simNao[rand.Intn(len(simNao))], // mental_disorder
			simNao[rand.Intn(len(simNao))], // medication
			rand.Intn(11), rand.Intn(11), rand.Intn(11), rand.Intn(11), rand.Intn(11), rand.Intn(11), // Níveis emocionais
			queixas[rand.Intn(len(queixas))],
			historicosQueixa[rand.Intn(len(historicosQueixa))], // DADO NOVO
			sinaisSintomas[rand.Intn(len(sinaisSintomas))],     // DADO NOVO
			tratamentosAtuais[rand.Intn(len(tratamentosAtuais))], // DADO NOVO
			notasMedico[rand.Intn(len(notasMedico))],           // DADO NOVO
		).Scan(&patientID)
		if err != nil {
			return fmt.Errorf("erro ao inserir paciente #%d: %w", i+1, err)
		}

		// Adiciona um agendamento futuro para o paciente
		doctorID := doctorIDs[rand.Intn(len(doctorIDs))]
		consultaData := time.Now().AddDate(0, 0, 7+rand.Intn(60))
		_, err = appointmentStmt.Exec(patientID, doctorID, consultaData, consultaData.Add(1*time.Hour), "agendado")
		if err != nil {
			return fmt.Errorf("erro ao inserir consulta para o paciente #%d: %w", i+1, err)
		}
	}

	return tx.Commit()
}

// Funções auxiliares
func newDBConnection() (*sql.DB, error) { dbType := os.Getenv("DB_TYPE"); dbHost := os.Getenv("DB_HOST"); dbPort := os.Getenv("DB_PORT"); dbUser := os.Getenv("DB_USER"); dbPass := os.Getenv("DB_PASS"); dbName := os.Getenv("DB_NAME"); connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPass, dbName); db, err := sql.Open(dbType, connStr); if err != nil { return nil, err }; if err = db.Ping(); err != nil { return nil, err }; return db, nil }
func initializeDB(db *sql.DB) error { _, err := db.Exec(createTableSQL); return err }
func createDefaultUsers(db *sql.DB) error { users := []struct { name string; email string; password string; userType string }{ {"Admin User", "admin@mediflow.com", "senha123", "admin"}, {"Dr. Exemplo", "terapeuta@mediflow.com", "senha123", "terapeuta"}, {"Secretaria Exemplo", "secretaria@mediflow.com", "senha123", "secretaria"}, }; for _, u := range users { hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.password), bcrypt.DefaultCost); if err != nil { return fmt.Errorf("falha ao gerar hash para %s: %w", u.email, err) }; query := ` INSERT INTO users (name, email, password_hash, user_type) VALUES ($1, $2, $3, $4) ON CONFLICT (email) DO NOTHING; `; _, err = db.Exec(query, u.name, u.email, string(hashedPassword), u.userType); if err != nil { return fmt.Errorf("falha ao inserir usuário %s: %w", u.email, err) } }; return nil }