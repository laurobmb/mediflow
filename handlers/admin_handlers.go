package handlers

import (
	"database/sql"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"
	"fmt"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"mediflow/storage"
)

// AdminHandler gerencia a lógica do painel de administração.
type AdminHandler struct {
	DB *sql.DB
}

// MonitoringData é a struct para os dados do dashboard.
type MonitoringData struct {
	UpcomingAppointments []map[string]interface{}
	CompletedCount       int
	HowFoundStats        map[string]int
	ProfileStats         map[string]int
	EmotionalAverages    map[string]float64
	ActiveDaysFilter     int
	TotalNewPatients     int
	PendingConsentPatients []PendingConsentPatient
}

// --- Funções de Gestão de Utilizadores ---

// handlers/admin_handlers.go

func (h *AdminHandler) ViewUsers(c *gin.Context) {
	// Pega a sessão atual
	session := sessions.Default(c)
	// Pega as mensagens flash de erro e sucesso da sessão
	errorFlashes := session.Flashes("error")
	successFlashes := session.Flashes("success")
	// Salva a sessão para limpar as mensagens após a leitura
	session.Save()

	rows, err := h.DB.Query("SELECT id, name, email, user_type FROM users ORDER BY name ASC")
	if err != nil {
		log.Printf("Erro ao buscar usuários: %v", err)
		c.HTML(http.StatusInternalServerError, "layouts/error.html", gin.H{"Title": "Erro", "Message": "Não foi possível carregar a lista de usuários."})
		return
	}
	defer rows.Close()

	var users []storage.User
	for rows.Next() {
		var user storage.User
		if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.UserType); err != nil {
			log.Printf("Erro ao escanear usuário: %v", err)
			continue
		}
		users = append(users, user)
	}

	// Envia os dados para o template, incluindo as mensagens
	c.HTML(http.StatusOK, "admin/view_users.html", gin.H{
		"Title":          "Gerenciar Usuários",
		"Users":          users,
		"ActiveNav":      "users",
		"ErrorFlashes":   errorFlashes,   // Passa as mensagens de erro
		"SuccessFlashes": successFlashes, // Passa as mensagens de sucesso
	})
}

func (h *AdminHandler) GetNewUserForm(c *gin.Context) {
	c.HTML(http.StatusOK, "admin/user_form.html", gin.H{"Title": "Adicionar Novo Usuário", "Action": "/admin/users/new", "IsNew": true, "UserTypes": []string{"admin", "terapeuta", "secretaria"}, "ActiveNav": "users"})
}

func (h *AdminHandler) PostNewUser(c *gin.Context) {
	name := c.PostForm("name")
	email := c.PostForm("email")
	password := c.PostForm("password")
	userType := c.PostForm("user_type")
	if name == "" || email == "" || password == "" || userType == "" {
		c.Redirect(http.StatusFound, "/admin/users/new")
		return
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Erro ao gerar hash da senha: %v", err)
		c.Redirect(http.StatusFound, "/admin/users")
		return
	}
	_, err = h.DB.Exec("INSERT INTO users (name, email, password_hash, user_type) VALUES ($1, $2, $3, $4)", name, email, string(hashedPassword), userType)
	if err != nil {
		log.Printf("Erro ao inserir novo usuário: %v", err)
	}
	c.Redirect(http.StatusFound, "/admin/users")
}

func (h *AdminHandler) GetEditUserForm(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.Redirect(http.StatusFound, "/admin/users")
		return
	}
	var user storage.User
	err = h.DB.QueryRow("SELECT id, name, email, user_type FROM users WHERE id = $1", id).Scan(&user.ID, &user.Name, &user.Email, &user.UserType)
	if err != nil {
		log.Printf("Erro ao buscar usuário para edição: %v", err)
		c.Redirect(http.StatusFound, "/admin/users")
		return
	}
	c.HTML(http.StatusOK, "admin/user_form.html", gin.H{"Title": "Editar Usuário", "Action": "/admin/users/edit/" + idStr, "IsNew": false, "User": user, "UserTypes": []string{"admin", "terapeuta", "secretaria"}, "ActiveNav": "users"})
}

func (h *AdminHandler) PostEditUser(c *gin.Context) {
	idStr := c.Param("id")
	name := c.PostForm("name")
	email := c.PostForm("email")
	password := c.PostForm("password")
	userType := c.PostForm("user_type")
	if password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("Erro ao gerar hash da senha na edição: %v", err)
			c.Redirect(http.StatusFound, "/admin/users")
			return
		}
		_, err = h.DB.Exec("UPDATE users SET name = $1, email = $2, user_type = $3, password_hash = $4 WHERE id = $5", name, email, userType, string(hashedPassword), idStr)
		if err != nil {
			log.Printf("Erro ao atualizar usuário com senha: %v", err)
		}
	} else {
		_, err := h.DB.Exec("UPDATE users SET name = $1, email = $2, user_type = $3 WHERE id = $4", name, email, userType, idStr)
		if err != nil {
			log.Printf("Erro ao atualizar usuário sem senha: %v", err)
		}
	}
	c.Redirect(http.StatusFound, "/admin/users")
}

// handlers/admin_handlers.go

func (h *AdminHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	session := sessions.Default(c)

	// --- LÓGICA ATUALIZADA ---
	// Antes de apagar, verificamos duas dependências: prontuários e agendamentos.

	// Verificação 1: Checar se o usuário tem registros de prontuário.
	var recordCount int
	recordQuery := `SELECT COUNT(*) FROM patient_records WHERE doctor_id = $1`
	err := h.DB.QueryRow(recordQuery, id).Scan(&recordCount)

	// Se houver registros, bloqueia a exclusão e informa o usuário.
	if err == nil && recordCount > 0 {
		message := fmt.Sprintf("Não é possível excluir este usuário, pois ele é responsável por %d registro(s) de prontuário.", recordCount)
		session.AddFlash(message, "error")
		session.Save()
		c.Redirect(http.StatusFound, "/admin/users")
		return // Interrompe a função aqui
	}

	// Verificação 2: Checar se o usuário tem agendamentos (não cancelados).
	var appointmentCount int
	appointmentQuery := `SELECT COUNT(*) FROM appointments WHERE doctor_id = $1 AND status != 'cancelado'`
	err = h.DB.QueryRow(appointmentQuery, id).Scan(&appointmentCount)

	// Se houver agendamentos, bloqueia a exclusão e informa o usuário.
	if err == nil && appointmentCount > 0 {
		message := fmt.Sprintf("Não é possível excluir este usuário, pois ele possui %d agendamento(s) na agenda.", appointmentCount)
		session.AddFlash(message, "error")
		session.Save()
		c.Redirect(http.StatusFound, "/admin/users")
		return // Interrompe a função aqui
	}

	// Se passou por ambas as verificações, pode excluir.
	_, err = h.DB.Exec("DELETE FROM users WHERE id = $1", id)
	if err != nil {
		log.Printf("Erro ao remover usuário: %v", err)
		session.AddFlash("Ocorreu um erro ao tentar remover o usuário.", "error")
	} else {
		session.AddFlash("Usuário removido com sucesso!", "success")
	}
	session.Save()
	c.Redirect(http.StatusFound, "/admin/users")
}

func (h *AdminHandler) ViewPatients(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}
	searchTerm := c.Query("search")
	pageSize := 10
	offset := (page - 1) * pageSize

	// CORRETO: A query busca 6 colunas.
	query := `SELECT id, name, email, phone, consent_given_at, access_token FROM patients`
	countQuery := `SELECT COUNT(*) FROM patients`
	var args []interface{}
	var countArgs []interface{}
	if searchTerm != "" {
		condition := ` WHERE name ILIKE $1 OR email ILIKE $1`
		query += condition
		countQuery += condition
		args = append(args, "%"+searchTerm+"%")
		countArgs = append(countArgs, "%"+searchTerm+"%")
	}
	query += ` ORDER BY name ASC LIMIT $` + strconv.Itoa(len(args)+1) + ` OFFSET $` + strconv.Itoa(len(args)+2)
	args = append(args, pageSize, offset)

	var totalRecords int
	err = h.DB.QueryRow(countQuery, countArgs...).Scan(&totalRecords)
	if err != nil {
		log.Printf("Erro ao contar pacientes: %v", err)
		c.HTML(http.StatusInternalServerError, "layouts/error.html", gin.H{"Title": "Erro", "Message": "Erro ao buscar dados."})
		return
	}
	totalPages := int(math.Ceil(float64(totalRecords) / float64(pageSize)))

	rows, err := h.DB.Query(query, args...)
	if err != nil {
		log.Printf("Erro ao buscar pacientes: %v", err)
		c.HTML(http.StatusInternalServerError, "layouts/error.html", gin.H{"Title": "Erro", "Message": "Erro ao buscar dados."})
		return
	}
	defer rows.Close()

	var patients []storage.Patient
	for rows.Next() {
		var patient storage.Patient
		var email, phone sql.NullString
		var consentGivenAt sql.NullTime
		var accessToken sql.NullString

		// CORREÇÃO: O Scan agora tem 6 variáveis de destino, correspondendo às 6 colunas da query.
		if err := rows.Scan(&patient.ID, &patient.Name, &email, &phone, &consentGivenAt, &accessToken); err != nil {
			log.Printf("Erro ao escanear paciente: %v", err)
			continue
		}

		if email.Valid {
			patient.Email = email.String
		}
		if phone.Valid {
			patient.Phone = phone.String
		}
		if consentGivenAt.Valid {
			patient.ConsentGivenAt = consentGivenAt
		}
		if accessToken.Valid {
			patient.AccessToken = accessToken
		}

		patients = append(patients, patient)
	}

	c.HTML(http.StatusOK, "admin/view_patients.html", gin.H{
		"Title":       "Gerenciar Pacientes",
		"Patients":    patients,
		"TotalPages":  totalPages,
		"CurrentPage": page,
		"SearchTerm":  searchTerm,
		"ActiveNav":   "patients",
	})
}

func (h *AdminHandler) SearchPatientsAPI(c *gin.Context) {
	term := c.Query("term")
	if term == "" {
		c.JSON(http.StatusOK, []string{})
		return
	}
	query := `SELECT name FROM patients WHERE name ILIKE $1 ORDER BY name ASC LIMIT 10`
	rows, err := h.DB.Query(query, term+"%")
	if err != nil {
		log.Printf("Erro na busca por autocompletar: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro no servidor"})
		return
	}
	defer rows.Close()
	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			continue
		}
		names = append(names, name)
	}
	c.JSON(http.StatusOK, names)
}

func (h *AdminHandler) GetNewPatientForm(c *gin.Context) {
	c.HTML(http.StatusOK, "admin/patient_form.html", gin.H{
		"Title":     "Adicionar Novo Paciente",
		"Action":    "/admin/patients/new",
		"IsNew":     true,
		"Patient":   storage.Patient{}, // Envia um paciente vazio para o template
		"ActiveNav": "patients",
	})
}

func (h *AdminHandler) PostNewPatient(c *gin.Context) {
	var patient storage.Patient
	if err := c.ShouldBind(&patient); err != nil {
		log.Printf("Erro ao fazer bind do formulário do paciente: %v", err)
		c.Redirect(http.StatusFound, "/admin/patients")
		return
	}
	
	query := `
        INSERT INTO patients (
            consent_date, consent_name, consent_cpf_rg, signature_date, signature_location, 
            name, address_street, address_number, address_neighborhood, address_city, address_state, 
            phone, mobile, dob, age, gender, marital_status, children, num_children, profession, email, 
            emergency_contact, emergency_phone, emergency_other, 
            created_at, updated_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26)
    `
	
	_, err := h.DB.Exec(query, 
		toDate(patient.ConsentDate), patient.ConsentName, patient.ConsentCpfRg, toDate(patient.SignatureDate), patient.SignatureLocation, 
		patient.Name, patient.AddressStreet, patient.AddressNumber, patient.AddressNeighborhood, patient.AddressCity, patient.AddressState, 
		patient.Phone, patient.Mobile, toDate(patient.DOB), patient.Age, patient.Gender, patient.MaritalStatus, patient.Children, patient.NumChildren, 
		patient.Profession, patient.Email, patient.EmergencyContact, patient.EmergencyPhone, patient.EmergencyOther, 
		time.Now(), time.Now())
	if err != nil {
		log.Printf("Erro ao inserir novo paciente pelo admin: %v", err)
	}
	c.Redirect(http.StatusFound, "/admin/patients")
}


type PendingConsentPatient struct {
    ID   int
    Name string
	AccessToken sql.NullString // Adicione este campo	
}
// Estrutura para passar todos os dados necessários para o template
type PatientEditPageData struct {
	Title        string
	Action       string
	IsNew        bool
	ActiveNav    string
	Patient      storage.Patient
	LatestRecord storage.PatientRecord   // O registro mais recente para preencher o formulário
	History      []storage.PatientRecord // Todos os registros para exibir na lista de histórico
	UserType     string // <-- CAMPO ADICIONADO	
}

// handlers/admin_handlers.go

// GetPatientDataForForm é uma função reutilizável para buscar todos os dados de um paciente.
func GetPatientDataForForm(db *sql.DB, patientID int) (PatientEditPageData, error) {
	pageData := PatientEditPageData{}

	// 1. Buscar os dados cadastrais do paciente da tabela 'patients'
	patientQuery := `SELECT id, name, consent_date, consent_name, consent_cpf_rg, signature_date, signature_location, 
		address_street, address_number, address_neighborhood, address_city, address_state, 
		phone, mobile, dob, age, gender, marital_status, children, num_children, profession, email, 
		emergency_contact, emergency_phone, emergency_other, consent_given_at
		FROM patients WHERE id = $1`

	var consentDate, signatureDate, dob, consentGivenAt sql.NullTime
	var consentName, consentCpfRg, signatureLocation, addressStreet, addressNumber, addressNeighborhood, addressCity, addressState, phone, mobile, gender, maritalStatus, children, profession, email, emergencyContact, emergencyPhone, emergencyOther sql.NullString
	var age, numChildren sql.NullInt64

	err := db.QueryRow(patientQuery, patientID).Scan(
		&pageData.Patient.ID, &pageData.Patient.Name, &consentDate, &consentName, &consentCpfRg, &signatureDate, &signatureLocation,
		&addressStreet, &addressNumber, &addressNeighborhood, &addressCity, &addressState,
		&phone, &mobile, &dob, &age, &gender, &maritalStatus, &children, &numChildren, &profession, &email,
		&emergencyContact, &emergencyPhone, &emergencyOther, &consentGivenAt,
	)
	if err != nil {
		log.Printf("Erro ao buscar paciente para edição: %v", err)
		return pageData, err
	}

	// ===================================================================
	// ESTE É O BLOCO ESSENCIAL QUE DEVE ESTAR AQUI
	// ===================================================================
	if consentDate.Valid { pageData.Patient.ConsentDate = consentDate.Time.Format("2006-01-02") }
	if consentName.Valid { pageData.Patient.ConsentName = consentName.String }
	if consentCpfRg.Valid { pageData.Patient.ConsentCpfRg = consentCpfRg.String }
	if signatureDate.Valid { pageData.Patient.SignatureDate = signatureDate.Time.Format("2006-01-02") }
	if signatureLocation.Valid { pageData.Patient.SignatureLocation = signatureLocation.String }
	if addressStreet.Valid { pageData.Patient.AddressStreet = addressStreet.String }
	if addressNumber.Valid { pageData.Patient.AddressNumber = addressNumber.String }
	if addressNeighborhood.Valid { pageData.Patient.AddressNeighborhood = addressNeighborhood.String }
	if addressCity.Valid { pageData.Patient.AddressCity = addressCity.String }
	if addressState.Valid { pageData.Patient.AddressState = addressState.String }
	if phone.Valid { pageData.Patient.Phone = phone.String }
	if mobile.Valid { pageData.Patient.Mobile = mobile.String }
	if dob.Valid { pageData.Patient.DOB = dob.Time.Format("2006-01-02") }
	if age.Valid { pageData.Patient.Age = int(age.Int64) }
	if gender.Valid { pageData.Patient.Gender = gender.String }
	if maritalStatus.Valid { pageData.Patient.MaritalStatus = maritalStatus.String }
	if children.Valid { pageData.Patient.Children = children.String }
	if numChildren.Valid { pageData.Patient.NumChildren = int(numChildren.Int64) }
	if profession.Valid { pageData.Patient.Profession = profession.String }
	if email.Valid { pageData.Patient.Email = email.String }
	if emergencyContact.Valid { pageData.Patient.EmergencyContact = emergencyContact.String }
	if emergencyPhone.Valid { pageData.Patient.EmergencyPhone = emergencyPhone.String }
	if emergencyOther.Valid { pageData.Patient.EmergencyOther = emergencyOther.String }
	if consentGivenAt.Valid { pageData.Patient.ConsentGivenAt = consentGivenAt }


	// 2. Buscar o registro clínico MAIS RECENTE
	latestRecordQuery := `SELECT id, patient_id, doctor_id, record_date, anxiety_level, anger_level, fear_level, sadness_level, 
		joy_level, energy_level, main_complaint, complaint_history, signs_symptoms, current_treatment, notes
		FROM patient_records WHERE patient_id = $1 ORDER BY record_date DESC LIMIT 1`
	
	err = db.QueryRow(latestRecordQuery, patientID).Scan(
		&pageData.LatestRecord.ID, &pageData.LatestRecord.PatientID, &pageData.LatestRecord.DoctorID, &pageData.LatestRecord.RecordDate,
		&pageData.LatestRecord.AnxietyLevel, &pageData.LatestRecord.AngerLevel, &pageData.LatestRecord.FearLevel, &pageData.LatestRecord.SadnessLevel,
		&pageData.LatestRecord.JoyLevel, &pageData.LatestRecord.EnergyLevel, &pageData.LatestRecord.MainComplaint, &pageData.LatestRecord.ComplaintHistory,
		&pageData.LatestRecord.SignsSymptoms, &pageData.LatestRecord.CurrentTreatment, &pageData.LatestRecord.Notes,
	)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Erro ao buscar último registro do paciente: %v", err)
	}

	// 3. Buscar TODO o histórico de registros
	historyQuery := `SELECT r.id, r.record_date, u.name as doctor_name,
		r.main_complaint, r.complaint_history, r.signs_symptoms, r.current_treatment, r.notes,
		r.anxiety_level, r.anger_level, r.fear_level, r.sadness_level, r.joy_level, r.energy_level
		FROM patient_records r JOIN users u ON r.doctor_id = u.id
		WHERE r.patient_id = $1 ORDER BY r.record_date DESC`

	rows, err := db.Query(historyQuery, patientID)
	if err != nil {
		log.Printf("Erro ao buscar histórico do paciente: %v", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var rec storage.PatientRecord
			if err := rows.Scan(&rec.ID, &rec.RecordDate, &rec.DoctorName, &rec.MainComplaint, &rec.ComplaintHistory, &rec.SignsSymptoms, &rec.CurrentTreatment, &rec.Notes, &rec.AnxietyLevel, &rec.AngerLevel, &rec.FearLevel, &rec.SadnessLevel, &rec.JoyLevel, &rec.EnergyLevel); err == nil {
				pageData.History = append(pageData.History, rec)
			}
		}
	}

	return pageData, nil
}

func (h *AdminHandler) GetEditPatientForm(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	pageData, err := GetPatientDataForForm(h.DB, id)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "layouts/error.html", gin.H{"Title": "Erro", "Message": "Não foi possível carregar os dados do paciente."})
		return
	}

	// Define os dados específicos para esta página/rota
	pageData.Title = "Editar Paciente e Prontuário"
	pageData.Action = "/admin/patients/edit/" + idStr
	pageData.ActiveNav = "patients"
	pageData.UserType = "admin" // <-- LINHA ADICIONADA AQUI

	c.HTML(http.StatusOK, "admin/patient_form.html", pageData)
}

// PostEditPatient (VERSÃO CORRIGIDA)
// Agora atualiza a tabela 'patients' com os dados clínicos mais recentes E cria o registro histórico.
func (h *AdminHandler) PostEditPatient(c *gin.Context) {
	idStr := c.Param("id")
	patientID, _ := strconv.Atoi(idStr)

	// --- LÓGICA ATUALIZADA ---

	// 1. ATUALIZAR dados cadastrais E CLÍNICOS na tabela 'patients'
	// A query agora inclui todos os campos do formulário para garantir que o estado mais recente seja salvo.
	queryPatients := `
		UPDATE patients SET 
			consent_date=$1, consent_name=$2, consent_cpf_rg=$3, signature_date=$4, signature_location=$5, 
			name=$6, address_street=$7, address_number=$8, address_neighborhood=$9, address_city=$10, address_state=$11, 
			phone=$12, mobile=$13, dob=$14, age=$15, gender=$16, marital_status=$17, children=$18, num_children=$19, 
			profession=$20, email=$21, emergency_contact=$22, emergency_phone=$23, emergency_other=$24, 
			-- Campos clínicos adicionados à atualização
			anxiety_level=$25, anger_level=$26, fear_level=$27, sadness_level=$28, joy_level=$29, energy_level=$30,
			main_complaint=$31, complaint_history=$32, signs_symptoms=$33, current_treatment=$34, notes=$35,
			updated_at=$36 
		WHERE id=$37`

	// Pega e converte todos os valores do formulário
	age, _ := strconv.Atoi(c.PostForm("age"))
	numChildren, _ := strconv.Atoi(c.PostForm("num_children"))
	anxiety, _ := strconv.Atoi(c.PostForm("anxiety_level"))
	anger, _ := strconv.Atoi(c.PostForm("anger_level"))
	fear, _ := strconv.Atoi(c.PostForm("fear_level"))
	sadness, _ := strconv.Atoi(c.PostForm("sadness_level"))
	joy, _ := strconv.Atoi(c.PostForm("joy_level"))
	energy, _ := strconv.Atoi(c.PostForm("energy_level"))

	_, err := h.DB.Exec(queryPatients,
		// Dados cadastrais
		toDate(c.PostForm("consent_date")), c.PostForm("consent_name_inline"), c.PostForm("consent_cpf_rg_inline"), toDate(c.PostForm("signature_date")), c.PostForm("signature_location"),
		c.PostForm("client_name"), c.PostForm("address_street"), c.PostForm("address_number"), c.PostForm("address_neighborhood"), c.PostForm("address_city"), c.PostForm("address_state"),
		c.PostForm("phone"), c.PostForm("mobile"), toDate(c.PostForm("dob")), age, c.PostForm("gender"), c.PostForm("marital_status"), c.PostForm("children"), numChildren,
		c.PostForm("profession"), c.PostForm("email"), c.PostForm("emergency_contact_name"), c.PostForm("emergency_contact_phone"), c.PostForm("emergency_contact_other"),
		// Dados clínicos
		anxiety, anger, fear, sadness, joy, energy,
		c.PostForm("main_complaint"), c.PostForm("complaint_history"), c.PostForm("signs_symptoms"), c.PostForm("current_treatment"), c.PostForm("notes"),
		// Timestamp e ID
		time.Now(), patientID)

	if err != nil {
		log.Printf("Erro ao atualizar dados do paciente na tabela 'patients': %v", err)
	}

	// 2. INSERIR o mesmo registro na tabela de histórico 'patient_records'
	session := sessions.Default(c)
	userID := session.Get("user_id").(int)

	queryRecords := `
		INSERT INTO patient_records (
			patient_id, doctor_id, anxiety_level, anger_level, fear_level, sadness_level, 
			joy_level, energy_level, main_complaint, complaint_history, signs_symptoms, 
			current_treatment, notes, record_date
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`

	_, err = h.DB.Exec(queryRecords,
		patientID, userID, anxiety, anger, fear, sadness, joy, energy,
		c.PostForm("main_complaint"), c.PostForm("complaint_history"), c.PostForm("signs_symptoms"),
		c.PostForm("current_treatment"), c.PostForm("notes"), time.Now())

	if err != nil {
		log.Printf("Erro ao inserir novo registro de prontuário: %v", err)
	}

	c.Redirect(http.StatusFound, "/admin/patients/edit/"+idStr)
}

func (h *AdminHandler) DeletePatient(c *gin.Context) {
    id := c.Param("id")
    _, err := h.DB.Exec("DELETE FROM patients WHERE id = $1", id)
    if err != nil {
        log.Printf("Erro ao remover paciente: %v", err)
    }
    c.Redirect(http.StatusFound, "/admin/patients")
}

// toDate é uma função de ajuda para converter strings de data para o formato do DB ou nil.
func toDate(dateStr string) interface{} {
    if dateStr == "" {
        return nil
    }
    t, err := time.Parse("2006-01-02", dateStr)
    if err != nil {
        return nil
    }
    return t
}

// --- Funções para Perfil e Prontuário ---

// GetPatientRecord exibe o prontuário do paciente e o formulário para adicionar uma nova entrada.
func (h *AdminHandler) GetPatientRecord(c *gin.Context) {
	idStr := c.Param("id")
	patientID, _ := strconv.Atoi(idStr)

	var patient storage.Patient
	h.DB.QueryRow("SELECT id, name FROM patients WHERE id = $1", patientID).Scan(&patient.ID, &patient.Name)

	query := `
        SELECT r.id, r.record_date, r.main_complaint, r.notes, u.name as doctor_name
        FROM patient_records r
        JOIN users u ON r.doctor_id = u.id
        WHERE r.patient_id = $1
        ORDER BY r.record_date DESC
    `
	rows, err := h.DB.Query(query, patientID)
	if err != nil {
		log.Printf("Erro ao buscar histórico do paciente: %v", err)
	}
	defer rows.Close()

	var records []storage.PatientRecord
	for rows.Next() {
		var rec storage.PatientRecord
		rows.Scan(&rec.ID, &rec.RecordDate, &rec.MainComplaint, &rec.Notes, &rec.DoctorName)
		records = append(records, rec)
	}

	c.HTML(http.StatusOK, "admin/patient_record.html", gin.H{
		"Title":     "Prontuário de " + patient.Name,
		"Patient":   patient,
		"Records":   records,
		"ActiveNav": "patients",
	})
}

// PostNewPatientRecord guarda uma nova entrada no prontuário do paciente.
func (h *AdminHandler) PostNewPatientRecord(c *gin.Context) {
	var record storage.PatientRecord
	if err := c.ShouldBind(&record); err != nil {
		log.Printf("Erro ao fazer bind do registo do paciente: %v", err)
		c.Redirect(http.StatusFound, "/admin/patients")
		return
	}

	session := sessions.Default(c)
	userID := session.Get("user_id").(int)
	record.DoctorID = userID

	query := `
        INSERT INTO patient_records (
            patient_id, doctor_id, anxiety_level, anger_level, fear_level, 
            sadness_level, joy_level, energy_level, main_complaint, 
            complaint_history, signs_symptoms, current_treatment, notes
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
    `
	_, err := h.DB.Exec(query,
		record.PatientID, record.DoctorID, record.AnxietyLevel, record.AngerLevel, record.FearLevel,
		record.SadnessLevel, record.JoyLevel, record.EnergyLevel, record.MainComplaint,
		record.ComplaintHistory, record.SignsSymptoms, record.CurrentTreatment, record.Notes,
	)
	if err != nil {
		log.Printf("Erro ao inserir novo registo de paciente: %v", err)
	}

	c.Redirect(http.StatusFound, "/admin/patients/record/"+strconv.Itoa(record.PatientID))
}


func (h *AdminHandler) GetPatientProfile(c *gin.Context) {
	idStr := c.Param("id")
	patientID, _ := strconv.Atoi(idStr)

	var patient storage.Patient
	err := h.DB.QueryRow("SELECT id, name FROM patients WHERE id = $1", patientID).Scan(&patient.ID, &patient.Name)
	if err != nil {
		log.Printf("Erro ao buscar perfil do paciente: %v", err)
		c.Redirect(http.StatusFound, "/admin/patients")
		return
	}

	// Busca agendamentos futuros (que têm preço)
	futureAppointments, err := h.getAppointmentsByTime(patientID, ">=")
	if err != nil {
		log.Printf("Erro ao buscar consultas futuras: %v", err)
	}

	// Busca agendamentos passados (que também têm preço)
	pastAppointments, err := h.getAppointmentsByTime(patientID, "<")
	if err != nil {
		log.Printf("Erro ao buscar histórico de consultas: %v", err)
	}

	// Busca a lista de terapeutas para o formulário de agendamento
	var doctors []storage.User
	rows, err := h.DB.Query("SELECT id, name FROM users WHERE user_type = 'terapeuta' ORDER BY name ASC")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var doc storage.User
			if err := rows.Scan(&doc.ID, &doc.Name); err == nil {
				doctors = append(doctors, doc)
			}
		}
	}

	c.HTML(http.StatusOK, "admin/patient_profile.html", gin.H{
		"Title":              "Perfil de " + patient.Name,
		"Patient":            patient,
		"FutureAppointments": futureAppointments,
		"PastAppointments":   pastAppointments,
		"Doctors":            doctors,
		"ActiveNav":          "patients",
	})
}

// PostNewAppointment processa o agendamento de uma nova consulta.
func (h *AdminHandler) PostNewAppointment(c *gin.Context) {
    patientIDStr := c.PostForm("patient_id")
    doctorIDStr := c.PostForm("doctor_id")
    dateStr := c.PostForm("appointment_date")
    timeStr := c.PostForm("start_time")
    status := c.PostForm("status")
    notes := c.PostForm("notes")
	// NOVO: Pega o preço do formulário
	priceStr := c.PostForm("price")

    patientID, _ := strconv.Atoi(patientIDStr)
    doctorID, _ := strconv.Atoi(doctorIDStr)
	price, _ := strconv.ParseFloat(priceStr, 64)
		
    startTime, err := time.Parse("2006-01-02 15:04", dateStr+" "+timeStr)
    if err != nil {
        log.Printf("Erro ao converter data/hora do agendamento: %v", err)
        c.Redirect(http.StatusFound, "/admin/patients/profile/"+patientIDStr)
        return
    }
    endTime := startTime.Add(1 * time.Hour)

    if status == "agendado" && startTime.Before(time.Now()) {
        status = "concluido"
    }

	// NOVO: Query agora inclui a coluna 'price'
	query := `INSERT INTO appointments (patient_id, doctor_id, start_time, end_time, status, notes, price, created_at, updated_at) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err = h.DB.Exec(query, patientID, doctorID, startTime, endTime, status, notes, price, time.Now(), time.Now())
	if err != nil {
		log.Printf("Erro ao agendar nova consulta: %v", err)
	}

    c.Redirect(http.StatusFound, "/admin/patients/profile/"+patientIDStr)
}

// getAppointmentsByTime é uma função de ajuda para buscar consultas.
func (h *AdminHandler) getAppointmentsByTime(patientID int, comparison string) ([]map[string]interface{}, error) {
	query := `
		SELECT a.id, a.start_time, a.status, a.notes, u.name as doctor_name, a.price, a.payment_status
		FROM appointments a
		JOIN users u ON a.doctor_id = u.id
		WHERE a.patient_id = $1 AND a.start_time ` + comparison + ` $2
		ORDER BY a.start_time DESC
	`
    rows, err := h.DB.Query(query, patientID, time.Now())
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var appointments []map[string]interface{}
    for rows.Next() {
        var app storage.Appointment
        var doctorName string
		var notes sql.NullString
		var price sql.NullFloat64
		var paymentStatus sql.NullString
		if err := rows.Scan(&app.ID, &app.StartTime, &app.Status, &notes, &doctorName, &price, &paymentStatus); err != nil {
			continue
		}
        app.Notes = notes.String
        appointments = append(appointments, map[string]interface{}{
            "ID":         app.ID,
            "StartTime":  app.StartTime,
            "Status":     app.Status,
            "Notes":      app.Notes,
            "DoctorName": doctorName,
			"Price":         price.Float64,
			"PaymentStatus": paymentStatus.String,			
        })
    }
    return appointments, nil
}

// --- FUNÇÃO DE MONITORAMENTO ATUALIZADA ---
func (h *AdminHandler) SystemMonitoring(c *gin.Context) {
    daysStr := c.DefaultQuery("days", "7")
    days, _ := strconv.Atoi(daysStr)
    if days == 0 {
        days = 7
    }
    startDate := time.Now().AddDate(0, 0, -days)

    data := MonitoringData{
        ActiveDaysFilter: days,
    }

	// --- LÓGICA ATUALIZADA PARA BUSCAR PACIENTES PENDENTES ---
	var pendingPatients []PendingConsentPatient
	// Query agora busca também o access_token
	query := `
		SELECT id, name, access_token FROM patients
		WHERE consent_given_at IS NULL
		ORDER BY created_at ASC
		LIMIT 10`
	rowsPending, err := h.DB.Query(query)
	if err != nil {
		log.Printf("Erro ao buscar pacientes com consentimento pendente: %v", err)
	} else {
		defer rowsPending.Close()
		for rowsPending.Next() {
			var p PendingConsentPatient
			// Scan agora inclui o accessToken
			if err := rowsPending.Scan(&p.ID, &p.Name, &p.AccessToken); err == nil {
				pendingPatients = append(pendingPatients, p)
			}
		}
	}
	data.PendingConsentPatients = pendingPatients
	// --- FIM DA LÓGICA ATUALIZADA ---
			
    rows, err := h.DB.Query(`
        SELECT p.name, u.name, a.start_time 
        FROM appointments a
        JOIN patients p ON a.patient_id = p.id
        JOIN users u ON a.doctor_id = u.id
        WHERE a.start_time >= NOW() ORDER BY a.start_time ASC LIMIT 10`)
    if err != nil {
		log.Printf("Erro ao buscar próximas consultas: %v", err)
        c.HTML(http.StatusInternalServerError, "layouts/error.html", gin.H{"Title": "Erro", "Message": "Não foi possível carregar os dados de monitoramento."})
		return
	}
    for rows.Next() {
        var patientName, doctorName string
        var startTime time.Time
        rows.Scan(&patientName, &doctorName, &startTime)
        data.UpcomingAppointments = append(data.UpcomingAppointments, map[string]interface{}{
            "PatientName": patientName,
            "DoctorName":  doctorName,
            "StartTime":   startTime,
        })
    }
    rows.Close()

    err = h.DB.QueryRow("SELECT COUNT(*) FROM appointments WHERE status = 'concluido' AND start_time >= $1", startDate).Scan(&data.CompletedCount)
    if err != nil {
        log.Printf("Erro ao contar consultas concluídas: %v", err)
    }

    data.HowFoundStats = make(map[string]int)
    rows, err = h.DB.Query("SELECT how_found, COUNT(*) FROM patients WHERE created_at >= $1 AND how_found IS NOT NULL AND how_found != '' GROUP BY how_found", startDate)
    if err != nil {
        log.Printf("Erro ao buscar fontes de aquisição: %v", err)
    } else {
        totalNew := 0
        for rows.Next() {
            var source string
            var count int
            rows.Scan(&source, &count)
            data.HowFoundStats[source] = count
            totalNew += count
        }
        data.TotalNewPatients = totalNew
        rows.Close()
    }

    data.ProfileStats = make(map[string]int)
    var physicalActivity, smoker, alcohol, repetitiveEffort, mentalDisorder, medication int
    h.DB.QueryRow("SELECT COUNT(*) FROM patients WHERE physical_activity = 'Sim' AND created_at >= $1", startDate).Scan(&physicalActivity)
    h.DB.QueryRow("SELECT COUNT(*) FROM patients WHERE smoker != 'Não' AND created_at >= $1", startDate).Scan(&smoker)
    h.DB.QueryRow("SELECT COUNT(*) FROM patients WHERE alcohol = 'Sim' AND created_at >= $1", startDate).Scan(&alcohol)
    h.DB.QueryRow("SELECT COUNT(*) FROM patients WHERE repetitive_effort = 'Sim' AND created_at >= $1", startDate).Scan(&repetitiveEffort)
    h.DB.QueryRow("SELECT COUNT(*) FROM patients WHERE mental_disorder = 'Sim' AND created_at >= $1", startDate).Scan(&mentalDisorder)
    h.DB.QueryRow("SELECT COUNT(*) FROM patients WHERE medication = 'Sim' AND created_at >= $1", startDate).Scan(&medication)
    data.ProfileStats["PhysicalActivity"] = physicalActivity
    data.ProfileStats["Smoker"] = smoker
    data.ProfileStats["Alcohol"] = alcohol
    data.ProfileStats["RepetitiveEffort"] = repetitiveEffort
    data.ProfileStats["MentalDisorder"] = mentalDisorder
    data.ProfileStats["Medication"] = medication

    data.EmotionalAverages = make(map[string]float64)
    var anxiety, anger, fear, sadness, joy, energy float64
    h.DB.QueryRow("SELECT COALESCE(AVG(anxiety_level), 0), COALESCE(AVG(anger_level), 0), COALESCE(AVG(fear_level), 0), COALESCE(AVG(sadness_level), 0), COALESCE(AVG(joy_level), 0), COALESCE(AVG(energy_level), 0) FROM patients WHERE created_at >= $1", startDate).Scan(
        &anxiety, &anger, &fear, &sadness, &joy, &energy,
    )
    data.EmotionalAverages["Anxiety"] = anxiety
    data.EmotionalAverages["Anger"] = anger
    data.EmotionalAverages["Fear"] = fear
    data.EmotionalAverages["Sadness"] = sadness
    data.EmotionalAverages["Joy"] = joy
    data.EmotionalAverages["Energy"] = energy

    c.HTML(http.StatusOK, "admin/monitoring.html", gin.H{
        "Title":     "Monitoramento do Sistema",
        "Data":      data,
        "ActiveNav": "monitoring",
    })
}

// AdminDashboard renderiza a dashboard principal do admin.
func (h *AdminHandler) AdminDashboard(c *gin.Context) {
	c.HTML(http.StatusOK, "admin/admin_dashboard.html", gin.H{
		"Title":     "Dashboard do Admin",
		"ActiveNav": "dashboard",
	})
}

// ViewAgenda renderiza a agenda para o admin.
func (h *AdminHandler) ViewAgenda(c *gin.Context) {
    dateParam := c.Query("date")
    var referenceDate time.Time
    if dateParam == "" {
        referenceDate = time.Now()
    } else {
        referenceDate, _ = time.Parse("2006-01-02", dateParam)
    }

    weekday := int(referenceDate.Weekday())
    startOfWeek := referenceDate.AddDate(0, 0, -weekday)
    endOfWeek := startOfWeek.AddDate(0, 0, 7)

    query := `
        SELECT a.id, a.start_time, a.status, p.name as patient_name, u.name as doctor_name, p.id as patient_id
        FROM appointments a
        JOIN patients p ON a.patient_id = p.id
        JOIN users u ON a.doctor_id = u.id
        WHERE a.start_time >= $1 AND a.start_time < $2
        ORDER BY a.start_time ASC
    `
    rows, err := h.DB.Query(query, startOfWeek, endOfWeek)
    if err != nil {
        log.Printf("Erro ao buscar agendamentos para a agenda (admin): %v", err)
        c.HTML(http.StatusInternalServerError, "layouts/error.html", gin.H{"Title": "Erro", "Message": "Não foi possível carregar a agenda."})
        return
    }
    defer rows.Close()

    type AppointmentDetails struct {
        ID          int
        StartTime   time.Time
        PatientName string
        DoctorName  string
        Status      string
        PatientID   int
    }
    appointmentsByDay := make(map[string][]AppointmentDetails)
    for rows.Next() {
        var app AppointmentDetails
        if err := rows.Scan(&app.ID, &app.StartTime, &app.Status, &app.PatientName, &app.DoctorName, &app.PatientID); err != nil {
            continue
        }
        dayKey := app.StartTime.Format("2006-01-02")
        appointmentsByDay[dayKey] = append(appointmentsByDay[dayKey], app)
    }

    type DaySchedule struct {
        Date         time.Time
        Appointments []AppointmentDetails
    }
    var weekSchedule []DaySchedule
    for i := 0; i < 7; i++ {
        day := startOfWeek.AddDate(0, 0, i)
        dayKey := day.Format("2006-01-02")
        weekSchedule = append(weekSchedule, DaySchedule{
            Date:         day,
            Appointments: appointmentsByDay[dayKey],
        })
    }

    c.HTML(http.StatusOK, "admin/agenda.html", gin.H{
        "Title":        "Agenda da Clínica",
        "WeekSchedule": weekSchedule,
        "PrevWeekLink": "/admin/agenda?date=" + startOfWeek.AddDate(0, 0, -7).Format("2006-01-02"),
        "NextWeekLink": "/admin/agenda?date=" + startOfWeek.AddDate(0, 0, 7).Format("2006-01-02"),
        "TodayLink":    "/admin/agenda",
        "ActiveNav":    "agenda",
    })
}

// handlers/admin_handlers.go

// (Cole estas 3 funções no final do arquivo)

// GetEditAppointmentForm renderiza o formulário para editar uma consulta.
func (h *AdminHandler) GetEditAppointmentForm(c *gin.Context) {
	appointmentIDStr := c.Param("id")
	appointmentID, _ := strconv.Atoi(appointmentIDStr)

	var app storage.Appointment
	var patientName string
	query := `SELECT a.id, a.patient_id, a.doctor_id, a.start_time, p.name 
			  FROM appointments a JOIN patients p ON a.patient_id = p.id 
			  WHERE a.id = $1`
	err := h.DB.QueryRow(query, appointmentID).Scan(&app.ID, &app.PatientID, &app.DoctorID, &app.StartTime, &patientName)
	if err != nil {
		log.Printf("Erro ao buscar consulta para edição (admin): %v", err)
		c.HTML(http.StatusNotFound, "layouts/error.html", gin.H{"Title": "Erro", "Message": "Consulta não encontrada."})
		return
	}

	var doctors []storage.User
	rows, _ := h.DB.Query("SELECT id, name FROM users WHERE user_type = 'terapeuta'")
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var doc storage.User
			if err := rows.Scan(&doc.ID, &doc.Name); err == nil {
				doctors = append(doctors, doc)
			}
		}
	}

	// Reutiliza o template da secretaria que já está pronto
	c.HTML(http.StatusOK, "secretaria/edit_appointment.html", gin.H{
		"Title":       "Editar Agendamento",
		"Appointment": app,
		"PatientID":   app.PatientID,
		"PatientName": patientName,
		"Doctors":     doctors,
		"ActiveNav":   "patients",
		"AdminPath":   true, // Indica que a rota de volta deve ser a do admin
	})
}

// PostEditAppointment processa a atualização de uma consulta.
func (h *AdminHandler) PostEditAppointment(c *gin.Context) {
	appointmentIDStr := c.Param("id")
	patientIDStr := c.Query("patient_id") // Pega o ID do paciente da URL

	doctorIDStr := c.PostForm("doctor_id")
	dateStr := c.PostForm("appointment_date")
	timeStr := c.PostForm("start_time")
	doctorID, _ := strconv.Atoi(doctorIDStr)

	startTime, err := time.Parse("2006-01-02 15:04", dateStr+" "+timeStr)
	if err != nil {
		log.Printf("Erro ao converter data/hora na edição (admin): %v", err)
		c.Redirect(http.StatusFound, "/admin/patients/profile/"+patientIDStr)
		return
	}
	endTime := startTime.Add(1 * time.Hour)

	query := `UPDATE appointments SET doctor_id = $1, start_time = $2, end_time = $3, updated_at = $4 WHERE id = $5`
	_, err = h.DB.Exec(query, doctorID, startTime, endTime, time.Now(), appointmentIDStr)
	if err != nil {
		log.Printf("Erro ao atualizar consulta (admin): %v", err)
	}

	c.Redirect(http.StatusFound, "/admin/patients/profile/"+patientIDStr)
}

// CancelAppointment atualiza o status de uma consulta para 'cancelado'.
func (h *AdminHandler) CancelAppointment(c *gin.Context) {
	appointmentID := c.Param("id")
	patientID := c.Query("patient_id")

	query := `UPDATE appointments SET status = 'cancelado', updated_at = $1 WHERE id = $2`
	_, err := h.DB.Exec(query, time.Now(), appointmentID)
	if err != nil {
		log.Printf("ERRO DE BANCO DE DADOS ao cancelar consulta (admin): %v", err)
	}

	c.Redirect(http.StatusFound, "/admin/patients/profile/"+patientID)
}