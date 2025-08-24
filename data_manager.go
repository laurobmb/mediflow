package main

import (
	"database/sql"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
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
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL
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
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL
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
  price NUMERIC(10, 2) DEFAULT 0.00,
  payment_status VARCHAR(50) NOT NULL DEFAULT 'pendente' CHECK (payment_status IN ('pendente', 'pago', 'isento')),
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS audit_logs (
  id SERIAL PRIMARY KEY,
  user_id INT,
  user_name VARCHAR(255),
  action TEXT NOT NULL,
  target_type VARCHAR(255),
  target_id INT,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
`

func main() {
	initDB := flag.Bool("init", false, "Inicializa o banco de dados e cria tabelas.")
	createUsers := flag.Bool("create-users", false, "Cria usuários de exemplo.")
	populateDB := flag.Bool("populate", false, "Popula o banco com dados de teste.")
	viewLogs := flag.Bool("audit", false, "Exibe os logs de auditoria do sistema.")
	viewDeletedPatients := flag.Bool("deleted-patients", false, "Exibe os pacientes que foram removidos (soft delete).")
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

	if *viewLogs {
		fmt.Println("Exibindo logs de auditoria...")
		if err := displayAuditLogs(db); err != nil {
			log.Fatalf("Erro ao exibir logs de auditoria: %v", err)
		}
	}

	if *viewDeletedPatients {
		fmt.Println("Exibindo pacientes removidos...")
		if err := displayDeletedPatients(db); err != nil {
			log.Fatalf("Erro ao exibir pacientes removidos: %v", err)
		}
	}

	if !*initDB && !*createUsers && !*populateDB && !*viewLogs && !*viewDeletedPatients {
		fmt.Println("Use uma flag para executar uma ação: -init, -create-users, -populate, -audit, ou -deleted-patients.")
	}
}

// FUNÇÃO PARA POPULAR O BANCO COM DADOS RICOS E COMPLETOS (VERSÃO ATUALIZADA)
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

	// --- DADOS CLÍNICOS CONSISTENTES E DETALHADOS ---
	// Perfil geral
	physicalActivity := "Não"
	smoker := "Não"
	alcohol := "Sim, socialmente nos fins de semana."
	repetitiveEffort := "Sim, trabalho com digitação intensa."
	mentalDisorder := "Sim"
	mentalDisorderTreatment := "Não"
	mentalDisorderDetails := "Paciente relata histórico familiar de ansiedade, mas nunca buscou um diagnóstico formal."
	medication := "Sim"
	medicationDetails := "Faz uso esporádico de Rivotril (clonazepam) 0.5mg, sem prescrição médica, quando se sente sobrecarregado."
	surgery := "Não"
	allergies := "Não"

	// Queixas e histórico
	queixaPrincipal := "Sente uma ansiedade generalizada e constante, com picos de pânico em situações de estresse no trabalho. Relata dificuldade para dormir e uma sensação de sobrecarga mental."
	historicoQueixa := "Os sintomas se intensificaram nos últimos 8 meses, após uma mudança de responsabilidades no emprego. Paciente menciona que sempre foi uma pessoa 'ansiosa', mas que agora a situação está insustentável e afetando seus relacionamentos."
	sinaisSintomas := "Aperto no peito, taquicardia, insônia, irritabilidade, dificuldade de concentração e pensamentos negativos recorrentes sobre o futuro e seu desempenho profissional."
	tratamentoAtual := "Nenhum tratamento medicamentoso no momento. Tentou meditação através de aplicativos, mas com pouca consistência."
	notasMedico := "Paciente demonstra bom insight sobre os gatilhos de sua ansiedade, mas apresenta dificuldade em estabelecer limites no ambiente de trabalho. Na sessão, exploramos a origem de sua necessidade de performance e a conexão com o medo de falhar. Demonstrou abertura para praticar técnicas de relaxamento e reestruturação cognitiva."
	// --- FIM DOS DADOS CLÍNICOS ---

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	patientStmt, err := tx.Prepare(`
        INSERT INTO patients (
            name, email, age, dob, phone, mobile, profession, address_city, address_state,
            how_found, physical_activity, smoker, alcohol, repetitive_effort, 
            mental_disorder, mental_disorder_treatment, mental_disorder_details, 
            medication, medication_details, surgery, allergies,
            anxiety_level, anger_level, fear_level, sadness_level, joy_level, energy_level,  
            main_complaint, complaint_history, signs_symptoms, current_treatment, notes,
            access_token  
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32, $33)
        RETURNING id`)
	if err != nil {
		return err
	}
	defer patientStmt.Close()

	// --- NOVA PREPARED STATEMENT PARA O HISTÓRICO ---
	recordStmt, err := tx.Prepare(`
		INSERT INTO patient_records (
			patient_id, doctor_id, anxiety_level, anger_level, fear_level, sadness_level, 
			joy_level, energy_level, main_complaint, complaint_history, signs_symptoms, 
			current_treatment, notes, record_date
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`)
	if err != nil {
		return err
	}
	defer recordStmt.Close()

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

		tokenBytes := make([]byte, 16)
		rand.Read(tokenBytes)
		token := hex.EncodeToString(tokenBytes)

		// Níveis emocionais para esta iteração
		ansiedade := rand.Intn(5) + 5 // Ansiedade (5-10)
		raiva := rand.Intn(6)      // Raiva (0-5)
		medo := rand.Intn(5) + 3       // Medo (3-8)
		tristeza := rand.Intn(5) + 5   // Tristeza (5-10)
		alegria := rand.Intn(4)      // Alegria (0-3)
		energia := rand.Intn(5)      // Energia (0-4)

		var patientID int
		err := patientStmt.QueryRow(
			// Informações Básicas (9)
			nomeCompleto, email, idade, dataNascimento, telefone, telefone,
			profissoes[rand.Intn(len(profissoes))],
			cidades[cidadeIdx], estados[cidadeIdx],
			// Como Encontrou (1)
			howFoundOptions[rand.Intn(len(howFoundOptions))],
			// Perfil Clínico (11)
			physicalActivity, smoker, alcohol, repetitiveEffort,
			mentalDisorder, mentalDisorderTreatment, mentalDisorderDetails,
			medication, medicationDetails, surgery, allergies,
			// Níveis Emocionais (6)
			ansiedade, raiva, medo, tristeza, alegria, energia,
			// Campos de Texto (5)
			queixaPrincipal,
			historicoQueixa,
			sinaisSintomas,
			tratamentoAtual,
			notasMedico,
			// Token (1)
			token,
		).Scan(&patientID)

		if err != nil {
			return fmt.Errorf("erro ao inserir paciente #%d: %w", i+1, err)
		}

		// --- INSERIR O REGISTRO INICIAL NO HISTÓRICO ---
		doctorID := doctorIDs[rand.Intn(len(doctorIDs))]
		_, err = recordStmt.Exec(
			patientID, doctorID, ansiedade, raiva, medo, tristeza,
			alegria, energia, queixaPrincipal, historicoQueixa, sinaisSintomas,
			tratamentoAtual, notasMedico, time.Now().AddDate(0, 0, -7), // Data do registro uma semana atrás
		)
		if err != nil {
			return fmt.Errorf("erro ao inserir registro histórico para o paciente #%d: %w", i+1, err)
		}
		// --- FIM DA INSERÇÃO DO HISTÓRICO ---

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

// NOVA FUNÇÃO PARA EXIBIR PACIENTES REMOVIDOS
func displayDeletedPatients(db *sql.DB) error {
	query := `SELECT id, name, email, phone, deleted_at 
			  FROM patients 
			  WHERE deleted_at IS NOT NULL 
			  ORDER BY deleted_at DESC`

	rows, err := db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	fmt.Println(strings.Repeat("-", 120))
	fmt.Printf("%-5s | %-40s | %-30s | %-20s | %-30s\n", "ID", "Nome", "Email", "Telefone", "Data da Remoção")
	fmt.Println(strings.Repeat("-", 120))

	for rows.Next() {
		var id int
		var deletedAt time.Time
		var name, email, phone sql.NullString

		if err := rows.Scan(&id, &name, &email, &phone, &deletedAt); err != nil {
			log.Printf("Erro ao escanear paciente removido: %v", err)
			continue
		}

		fmt.Printf("%-5d | %-40s | %-30s | %-20s | %-30s\n",
			id,
			name.String,
			email.String,
			phone.String,
			deletedAt.Format("02/01/2006 15:04:05"),
		)
	}
	fmt.Println(strings.Repeat("-", 120))
	return nil
}

// NOVA FUNÇÃO PARA EXIBIR OS LOGS DE AUDITORIA
func displayAuditLogs(db *sql.DB) error {
	query := `SELECT id, user_id, user_name, action, target_type, target_id, created_at 
              FROM audit_logs ORDER BY created_at DESC`

	rows, err := db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	fmt.Println(strings.Repeat("-", 120))
	fmt.Printf("%-5s | %-25s | %-50s | %-10s | %-10s | %-30s\n", "ID", "Usuário", "Ação", "Alvo", "ID Alvo", "Data")
	fmt.Println(strings.Repeat("-", 120))

	for rows.Next() {
		var id int
		var createdAt time.Time
		var userID sql.NullInt64
		var userName, targetType sql.NullString
		var targetID sql.NullInt64
		var action string

		if err := rows.Scan(&id, &userID, &userName, &action, &targetType, &targetID, &createdAt); err != nil {
			log.Printf("Erro ao escanear log: %v", err)
			continue
		}

		userNameStr := "N/A"
		if userName.Valid {
			userNameStr = fmt.Sprintf("%d - %s", userID.Int64, userName.String)
		}

		targetTypeStr := "N/A"
		if targetType.Valid {
			targetTypeStr = targetType.String
		}

		targetIDStr := "N/A"
		if targetID.Valid {
			targetIDStr = fmt.Sprintf("%d", targetID.Int64)
		}

		fmt.Printf("%-5d | %-25s | %-50s | %-10s | %-10s | %-30s\n",
			id,
			userNameStr,
			action,
			targetTypeStr,
			targetIDStr,
			createdAt.Format("02/01/2006 15:04:05"),
		)
	}
	fmt.Println(strings.Repeat("-", 120))
	return nil
}
