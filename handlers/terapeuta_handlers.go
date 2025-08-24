package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"time"
	"strconv" // <-- ADICIONE ESTA LINHA
	
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"mediflow/storage"
)

// TerapeutaHandler gerencia a lógica do painel do terapeuta.
type TerapeutaHandler struct {
	DB *sql.DB
}

// Struct simples para passar os dados para o template
type TerapeutaDashboardData struct {
	UpcomingAppointments []AppointmentDetails
	MyPatients           []storage.Patient
}

// TerapeutaDashboard busca os dados e renderiza a página inicial do terapeuta.
func (h *TerapeutaHandler) TerapeutaDashboard(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id").(int) // Pega o ID do terapeuta logado
	searchTerm := c.Query("search")
	data := TerapeutaDashboardData{}

	// 1. Buscar as próximas 10 consultas do terapeuta
	queryAppointments := `
		SELECT a.id, a.start_time, p.name as patient_name, p.id as patient_id
		FROM appointments a
		JOIN patients p ON a.patient_id = p.id
		WHERE a.doctor_id = $1 AND a.start_time >= $2 AND a.status = 'agendado'
		ORDER BY a.start_time ASC
		LIMIT 10`

	rows, err := h.DB.Query(queryAppointments, userID, time.Now())
	if err != nil {
		log.Printf("Erro ao buscar agendamentos do terapeuta: %v", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var app AppointmentDetails
			rows.Scan(&app.ID, &app.StartTime, &app.PatientName, &app.PatientID)
			data.UpcomingAppointments = append(data.UpcomingAppointments, app)
		}
	}

	// 2. Buscar uma lista de todos os pacientes associados a este terapeuta
	queryPatients := `
		SELECT DISTINCT p.id, p.name, p.email, p.phone
		FROM patients p
		JOIN appointments a ON p.id = a.patient_id
		WHERE a.doctor_id = $1 AND p.name ILIKE $2 AND p.deleted_at IS NULL
		ORDER BY p.name ASC`

	rows, err = h.DB.Query(queryPatients, userID, "%"+searchTerm+"%")
	if err != nil {
		log.Printf("Erro ao buscar pacientes do terapeuta: %v", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var p storage.Patient
			var email, phone sql.NullString
			rows.Scan(&p.ID, &p.Name, &email, &phone)
			p.Email = email.String
			p.Phone = phone.String
			data.MyPatients = append(data.MyPatients, p)
		}
	}

	c.HTML(http.StatusOK, "terapeuta/terapeuta_dashboard.html", gin.H{
		"Title":     "Meu Dashboard",
		"Data":      data,
		"ActiveNav": "dashboard",
		"SearchTerm": searchTerm, // Passa o termo de busca de volta para o HTML		
	})
}

// handlers/terapeuta_handlers.go

// ShowPatientRecord exibe o prontuário completo, reutilizando a lógica e o template do admin.
func (h *TerapeutaHandler) ShowPatientRecord(c *gin.Context) {
	session := sessions.Default(c)
	therapistID := session.Get("user_id").(int)
	patientIDStr := c.Param("id")
	patientID, _ := strconv.Atoi(patientIDStr)

	// Verificação de Segurança: Este terapeuta tem acesso a este paciente?
	var count int
	err := h.DB.QueryRow("SELECT COUNT(*) FROM appointments WHERE patient_id = $1 AND doctor_id = $2", patientID, therapistID).Scan(&count)
	if err != nil || count == 0 {
		c.HTML(http.StatusForbidden, "layouts/error.html", gin.H{"Title": "Acesso Negado", "Message": "Você não tem permissão para ver o prontuário deste paciente."})
		return
	}

	// Reutiliza a função de busca de dados que o admin também usa.
	pageData, err := GetPatientDataForForm(h.DB, patientID)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "layouts/error.html", gin.H{"Title": "Erro", "Message": "Não foi possível carregar os dados do paciente."})
		return
	}

	// Prepara os dados específicos para a página do terapeuta
	pageData.Title = "Prontuário do Paciente"
	pageData.Action = "/terapeuta/pacientes/prontuario/" + patientIDStr
	pageData.ActiveNav = "dashboard" // Mantém o dashboard como ativo no menu
	pageData.UserType = "terapeuta" // <-- LINHA ADICIONADA AQUI
	
	// Renderiza o MESMO template que o admin usa, garantindo que sejam idênticos.
	c.HTML(http.StatusOK, "admin/patient_form.html", pageData)
}


// ProcessPatientRecord salva os dados completos do prontuário, assim como o admin.
func (h *TerapeutaHandler) ProcessPatientRecord(c *gin.Context) {
	session := sessions.Default(c)
	therapistID := session.Get("user_id").(int)
	patientIDStr := c.Param("id")
	patientID, _ := strconv.Atoi(patientIDStr)

	// Verificação de Segurança
	var count int
	// CORREÇÃO: Usar QueryRow e Scan corretamente
	err := h.DB.QueryRow("SELECT COUNT(*) FROM appointments WHERE patient_id = $1 AND doctor_id = $2", patientID, therapistID).Scan(&count)
	if err != nil || count == 0 {
		c.HTML(http.StatusForbidden, "layouts/error.html", gin.H{"Title": "Acesso Negado", "Message": "Você não tem permissão para alterar o prontuário deste paciente."})
		return
	}

	// 1. Atualiza a tabela 'patients' com o estado mais recente
	queryPatients := `
		UPDATE patients SET 
			name=$1, address_street=$2, address_number=$3, address_neighborhood=$4, address_city=$5, address_state=$6,
			phone=$7, mobile=$8, dob=$9, age=$10, email=$11, profession=$12,
			anxiety_level=$13, anger_level=$14, fear_level=$15, sadness_level=$16, joy_level=$17, energy_level=$18,
			main_complaint=$19, complaint_history=$20, signs_symptoms=$21, current_treatment=$22, notes=$23,
			updated_at=$24 
		WHERE id=$25`

	age, _ := strconv.Atoi(c.PostForm("age"))
	anxiety, _ := strconv.Atoi(c.PostForm("anxiety_level"))
	anger, _ := strconv.Atoi(c.PostForm("anger_level"))
	fear, _ := strconv.Atoi(c.PostForm("fear_level"))
	sadness, _ := strconv.Atoi(c.PostForm("sadness_level"))
	joy, _ := strconv.Atoi(c.PostForm("joy_level"))
	energy, _ := strconv.Atoi(c.PostForm("energy_level"))

	// CORREÇÃO: Usar a atribuição correta para h.DB.Exec
	_, err = h.DB.Exec(queryPatients,
		c.PostForm("client_name"), c.PostForm("address_street"), c.PostForm("address_number"), c.PostForm("address_neighborhood"),
		c.PostForm("address_city"), c.PostForm("address_state"), c.PostForm("phone"), c.PostForm("mobile"),
		c.PostForm("dob"), age, c.PostForm("email"), c.PostForm("profession"),
		anxiety, anger, fear, sadness, joy, energy,
		c.PostForm("main_complaint"), c.PostForm("complaint_history"), c.PostForm("signs_symptoms"),
		c.PostForm("current_treatment"), c.PostForm("notes"),
		time.Now(), patientID)

	if err != nil {
		log.Printf("Erro ao ATUALIZAR paciente pelo terapeuta: %v", err)
	}

	// 2. Insere um novo registro no histórico
	queryRecords := `
		INSERT INTO patient_records (
			patient_id, doctor_id, anxiety_level, anger_level, fear_level, sadness_level, 
			joy_level, energy_level, main_complaint, complaint_history, signs_symptoms, 
			current_treatment, notes, record_date
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`

	// CORREÇÃO: Usar a atribuição correta para h.DB.Exec
	_, err = h.DB.Exec(queryRecords,
		patientID, therapistID, anxiety, anger, fear, sadness, joy, energy,
		c.PostForm("main_complaint"), c.PostForm("complaint_history"), c.PostForm("signs_symptoms"),
		c.PostForm("current_treatment"), c.PostForm("notes"), time.Now())

	if err != nil {
		log.Printf("Erro ao INSERIR registro de prontuário pelo terapeuta: %v", err)
	} else {
		logInfo := LogAction{
			DB:         h.DB,
			Context:    c,
			Action:     "Adicionou nova entrada ao prontuário",
			TargetType: "Paciente",
			TargetID:   patientID,
		}
		AddAuditLog(logInfo)
		}

	c.Redirect(http.StatusFound, "/terapeuta/pacientes/prontuario/"+patientIDStr)
}


// SearchMyPatientsAPI busca e retorna apenas os pacientes associados ao terapeuta logado.
func (h *TerapeutaHandler) SearchMyPatientsAPI(c *gin.Context) {
    session := sessions.Default(c)
    therapistID := session.Get("user_id").(int)
    term := c.Query("term")

    if term == "" {
        c.JSON(http.StatusOK, []string{})
        return
    }

    // query := `
    //     SELECT DISTINCT p.name
    //     FROM patients p
    //     JOIN appointments a ON p.id = a.patient_id
    //     WHERE a.doctor_id = $1 AND p.name ILIKE $2
    //     ORDER BY p.name ASC
    //     LIMIT 10`

	query := `
		SELECT DISTINCT p.name
		FROM patients p
		JOIN appointments a ON p.id = a.patient_id
		WHERE a.doctor_id = $1 AND p.name ILIKE $2 AND p.deleted_at IS NULL
		ORDER BY p.name ASC
		LIMIT 10`		

    rows, err := h.DB.Query(query, therapistID, term+"%")
    if err != nil {
        log.Printf("Erro na busca de pacientes do terapeuta: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro no servidor"})
        return
    }
    defer rows.Close()

    var names []string
    for rows.Next() {
        var name string
        if err := rows.Scan(&name); err == nil {
            names = append(names, name)
        }
    }
    c.JSON(http.StatusOK, names)
}