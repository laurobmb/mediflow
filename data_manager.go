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

// ... (a variável createTableSQL permanece a mesma) ...
var createTableSQL = `
DROP TABLE IF EXISTS consultation_summaries, appointments, patients, users CASCADE;

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
    consent_date DATE,
    consent_name VARCHAR(255),
    consent_cpf_rg VARCHAR(50),
    signature_date DATE,
    signature_location VARCHAR(255),
    name VARCHAR(255),
    address_street VARCHAR(255),
    address_number VARCHAR(255),
    address_neighborhood VARCHAR(255),
    address_city VARCHAR(255),
    address_state VARCHAR(255),
    phone VARCHAR(255),
    mobile VARCHAR(255),
    dob DATE,
    age INT,
    gender VARCHAR(255),
    marital_status VARCHAR(255),
    children VARCHAR(255),
    num_children INT,
    profession VARCHAR(255),
    email VARCHAR(255),
    emergency_contact VARCHAR(255),
    emergency_phone VARCHAR(255),
    emergency_other VARCHAR(255),
    repetitive_effort VARCHAR(255),
    physical_activity VARCHAR(255),
    smoker VARCHAR(255),
    alcohol VARCHAR(255),
    mental_disorder VARCHAR(255),
    religion VARCHAR(255),
    medication VARCHAR(255),
    surgery VARCHAR(255),
    allergies VARCHAR(255),
    anxiety_level INT,
    anger_level INT,
    fear_level INT,
    sadness_level INT,
    joy_level INT,
    energy_level INT,
    main_complaint TEXT,
    complaint_history TEXT,
    signs_symptoms TEXT,
    current_treatment TEXT,
    how_found VARCHAR(255),
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS appointments (
   id SERIAL PRIMARY KEY,
   patient_id INT NOT NULL REFERENCES patients(id) ON DELETE CASCADE,
   doctor_id INT NOT NULL REFERENCES users(id),
   start_time TIMESTAMP WITH TIME ZONE NOT NULL,
   end_time TIMESTAMP WITH TIME ZONE NOT NULL,
   notes TEXT,
   status VARCHAR(50) NOT NULL CHECK (status IN ('agendado', 'concluido', 'cancelado')),
   created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
   updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS consultation_summaries (
   id SERIAL PRIMARY KEY,
   appointment_id INT UNIQUE NOT NULL REFERENCES appointments(id) ON DELETE CASCADE,
   doctor_id INT NOT NULL REFERENCES users(id),
   patient_id INT NOT NULL REFERENCES patients(id),
   main_complaint TEXT,
   complaint_history TEXT,
   signs_symptoms TEXT,
   current_treatment TEXT,
   anxiety_level INT,
   anger_level INT,
   fear_level INT,
   sadness_level INT,
   joy_level INT,
   energy_level INT,
   created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
   updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
`

func main() {
	initDB := flag.Bool("init", false, "Inicializa o banco de dados e cria tabelas.")
	createUsers := flag.Bool("create-users", false, "Cria usuários de exemplo (admin, terapeuta, secretaria).")
	populateDB := flag.Bool("populate", false, "Popula o banco de dados com 45 pacientes de teste e suas consultas.")
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
		fmt.Println("Populando o banco de dados com 45 pacientes e suas consultas. Isso pode levar um momento...")
		if err := populateData(db, 45); err != nil {
			log.Fatalf("Erro ao popular o banco de dados: %v", err)
		}
		fmt.Println("Dados de teste inseridos com sucesso!")
	}

	if !*initDB && !*createUsers && !*populateDB {
		fmt.Println("Use uma flag para executar uma ação: -init, -create-users, ou -populate.")
	}
}

// FUNÇÃO PARA POPULAR PACIENTES E CONSULTAS (CORRIGIDA E COMPLETA)
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
		return fmt.Errorf("nenhum médico encontrado no banco de dados. Crie usuários com a flag -create-users primeiro")
	}

	nomes := []string{"Alice", "Bernardo", "Clara", "Davi", "Elena", "Felipe", "Giovanna", "Heitor", "Isadora", "Júlio", "Larissa", "Miguel", "Natália", "Otávio", "Pietra", "Rafael", "Sofia", "Thiago", "Valentina", "William"}
	sobrenomes := []string{"Andrade", "Barbosa", "Cardoso", "Dias", "Esteves", "Fernandes", "Gusmão", "Henriques", "Ibrahim", "Jesus", "Lopes", "Mendes", "Nogueira", "Pinto", "Queiroz", "Ribeiro", "Teixeira", "Vieira"}
	cidades := []string{"São Paulo", "Rio de Janeiro", "Belo Horizonte", "Salvador", "Fortaleza"}
	estados := []string{"SP", "RJ", "MG", "BA", "CE"}
	profissoes := []string{"Engenheiro(a)", "Advogado(a)", "Professor(a)", "Médico(a)", "Designer", "Desenvolvedor(a)"}
	queixas := []string{"Sente-se constantemente ansioso(a) devido à pressão no trabalho.", "Relata episódios de tristeza profunda.", "Tem dificuldades para dormir.", "Lida com ataques de pânico.", "Sente uma falta de motivação geral."}
	notasConsulta := []string{"Paciente demonstrou progresso na identificação de gatilhos de ansiedade.", "Discutimos estratégias de mindfulness para lidar com o stress.", "Sessão focada em técnicas de comunicação assertiva.", "Exploramos as causas da insónia e estabelecemos uma rotina de sono."}
	howFoundOptions := []string{"Instagram", "Google", "Indicação de conhecido", "Outro"} // CORREÇÃO: Nova lista de opções

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	patientStmt, err := tx.Prepare(`
        INSERT INTO patients (
            consent_date, consent_name, consent_cpf_rg, signature_date, signature_location, 
            name, address_street, address_number, address_neighborhood, address_city, address_state, 
            phone, mobile, dob, age, gender, marital_status, children, num_children, profession, email, 
            emergency_contact, emergency_phone, emergency_other, 
            repetitive_effort, physical_activity, smoker, alcohol, mental_disorder, religion, 
            medication, surgery, allergies, 
            anxiety_level, anger_level, fear_level, sadness_level, joy_level, energy_level, 
            main_complaint, complaint_history, signs_symptoms, current_treatment, how_found, notes, 
            created_at, updated_at
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, 
            $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32, $33, $34, 
            $35, $36, $37, $38, $39, $40, $41, $42, $43, $44, $45, $46, $47
        ) RETURNING id`)
	if err != nil {
		return err
	}
	defer patientStmt.Close()

	appointmentStmt, err := tx.Prepare(`INSERT INTO appointments (patient_id, doctor_id, start_time, end_time, status, notes, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`)
	if err != nil {
		return err
	}
	defer appointmentStmt.Close()

	for i := 0; i < patientCount; i++ {
		nomeCompleto := fmt.Sprintf("%s %s", nomes[rand.Intn(len(nomes))], sobrenomes[rand.Intn(len(sobrenomes))])
		email := fmt.Sprintf("paciente.%d@example.com", i)
		telefone := fmt.Sprintf("(%d) 9%d-%d", 11+rand.Intn(88), 8000+rand.Intn(1999), 1000+rand.Intn(8999))
		idade := 20 + rand.Intn(40)
		dataNascimento := time.Now().AddDate(-idade, rand.Intn(12), rand.Intn(28))
		cidadeIdx := rand.Intn(len(cidades))

		var patientID int
		err := patientStmt.QueryRow(
			time.Now().AddDate(0, 0, -rand.Intn(30)), // consent_date
			nomeCompleto,                             // consent_name
			fmt.Sprintf("%d%d%d.%d%d%d.%d%d%d-%d%d", rand.Intn(10), rand.Intn(10), rand.Intn(10), rand.Intn(10), rand.Intn(10), rand.Intn(10), rand.Intn(10), rand.Intn(10), rand.Intn(10), rand.Intn(10), rand.Intn(10)), // consent_cpf_rg
			time.Now().AddDate(0, 0, -rand.Intn(30)), // signature_date
			cidades[cidadeIdx],                       // signature_location
			nomeCompleto,                             // name
			"Rua das Flores",                         // address_street
			fmt.Sprintf("%d", 1+rand.Intn(1000)),      // address_number
			"Centro",                                 // address_neighborhood
			cidades[cidadeIdx],                       // address_city
			estados[cidadeIdx],                       // address_state
			telefone,                                 // phone
			telefone,                                 // mobile
			dataNascimento,                           // dob
			idade,                                    // age
			[]string{"M", "F"}[rand.Intn(2)],          // gender
			[]string{"Solteiro(a)", "Casado(a)", "Divorciado(a)"}[rand.Intn(3)], // marital_status
			"Não",                                    // children
			0,                                        // num_children
			profissoes[rand.Intn(len(profissoes))],    // profession
			email,                                    // email
			"Contato de Emergência",                  // emergency_contact
			telefone,                                 // emergency_phone
			"Amigo(a)",                               // emergency_other
			"Não",                                    // repetitive_effort
			"Sim",                                    // physical_activity
			"Não",                                    // smoker
			"Sim",                                    // alcohol
			"Sim",                                    // mental_disorder
			"Não",                                    // religion
			"Sim",                                    // medication
			"Não",                                    // surgery
			"Não",                                    // allergies
			rand.Intn(11),                            // anxiety_level
			rand.Intn(11),                            // anger_level
			rand.Intn(11),                            // fear_level
			rand.Intn(11),                            // sadness_level
			rand.Intn(11),                            // joy_level
			rand.Intn(11),                            // energy_level
			queixas[rand.Intn(len(queixas))],         // main_complaint
			"O paciente relata que os sintomas começaram há aproximadamente 6 meses.", // complaint_history
			"Insônia e falta de apetite.",            // signs_symptoms
			"Terapia cognitivo-comportamental.",      // current_treatment
			howFoundOptions[rand.Intn(len(howFoundOptions))], // how_found (CORRIGIDO)
			"Paciente responde bem à terapia. Recomenda-se continuar com as sessões.", // notes
			time.Now(),                               // created_at
			time.Now(),                               // updated_at
		).Scan(&patientID)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("erro ao inserir paciente #%d: %w", i+1, err)
		}

		// Criar consultas passadas
		for j := 0; j < 2+rand.Intn(4); j++ {
			doctorID := doctorIDs[rand.Intn(len(doctorIDs))]
			consultaData := time.Now().AddDate(0, -rand.Intn(6), -rand.Intn(28))
			_, err := appointmentStmt.Exec(patientID, doctorID, consultaData, consultaData.Add(1*time.Hour), "concluido", notasConsulta[rand.Intn(len(notasConsulta))], time.Now(), time.Now())
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("erro ao inserir consulta passada para o paciente #%d: %w", i+1, err)
			}
		}

		// Criar consultas futuras
		for j := 0; j < 1+rand.Intn(3); j++ {
			doctorID := doctorIDs[rand.Intn(len(doctorIDs))]
			consultaData := time.Now().AddDate(0, 0, 7+rand.Intn(60))
			_, err := appointmentStmt.Exec(patientID, doctorID, consultaData, consultaData.Add(1*time.Hour), "agendado", "Consulta de acompanhamento.", time.Now(), time.Now())
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("erro ao inserir consulta futura para o paciente #%d: %w", i+1, err)
			}
		}
	}

	return tx.Commit()
}


// ... (as funções newDBConnection, initializeDB, e createDefaultUsers permanecem as mesmas) ...
func newDBConnection() (*sql.DB, error) { dbType := os.Getenv("DB_TYPE"); dbHost := os.Getenv("DB_HOST"); dbPort := os.Getenv("DB_PORT"); dbUser := os.Getenv("DB_USER"); dbPass := os.Getenv("DB_PASS"); dbName := os.Getenv("DB_NAME"); connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPass, dbName); db, err := sql.Open(dbType, connStr); if err != nil { return nil, err }; if err = db.Ping(); err != nil { return nil, err }; return db, nil }
func initializeDB(db *sql.DB) error { _, err := db.Exec(createTableSQL); return err }
func createDefaultUsers(db *sql.DB) error { users := []struct { name string; email string; password string; userType string }{ {"Admin User", "admin@mediflow.com", "senha123", "admin"}, {"Dr. Exemplo", "terapeuta@mediflow.com", "senha123", "terapeuta"}, {"Secretaria Exemplo", "secretaria@mediflow.com", "senha123", "secretaria"}, }; for _, u := range users { hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.password), bcrypt.DefaultCost); if err != nil { return fmt.Errorf("falha ao gerar hash para %s: %w", u.email, err) }; query := ` INSERT INTO users (name, email, password_hash, user_type) VALUES ($1, $2, $3, $4) ON CONFLICT (email) DO NOTHING; `; _, err = db.Exec(query, u.name, u.email, string(hashedPassword), u.userType); if err != nil { return fmt.Errorf("falha ao inserir usuário %s: %w", u.email, err) } }; return nil }
