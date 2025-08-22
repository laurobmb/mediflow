package handlers

import (
	"database/sql"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

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
}

// --- Funções de Gestão de Utilizadores ---
func (h *AdminHandler) ViewUsers(c *gin.Context) {
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
	c.HTML(http.StatusOK, "admin/view_users.html", gin.H{"Title": "Gerenciar Usuários", "Users": users, "ActiveNav": "users"})
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

func (h *AdminHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	_, err := h.DB.Exec("DELETE FROM users WHERE id = $1", id)
	if err != nil {
		log.Printf("Erro ao remover usuário: %v", err)
	}
	c.Redirect(http.StatusFound, "/admin/users")
}


// --- Funções de Gestão de Pacientes ---
func (h *AdminHandler) ViewPatients(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}
	searchTerm := c.Query("search")
	pageSize := 10
	offset := (page - 1) * pageSize
	query := `SELECT id, name, email, phone FROM patients`
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
		if err := rows.Scan(&patient.ID, &patient.Name, &patient.Email, &patient.Phone); err != nil {
			log.Printf("Erro ao escanear paciente: %v", err)
			continue
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
	
	query := `INSERT INTO patients (consent_date, consent_name, consent_cpf_rg, signature_date, signature_location, name, address_street, address_number, address_neighborhood, address_city, address_state, phone, mobile, dob, age, gender, marital_status, children, num_children, profession, email, emergency_contact, emergency_phone, emergency_other, repetitive_effort, physical_activity, smoker, alcohol, mental_disorder, religion, medication, surgery, allergies, anxiety_level, anger_level, fear_level, sadness_level, joy_level, energy_level, main_complaint, complaint_history, signs_symptoms, current_treatment, how_found, notes, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32, $33, $34, $35, $36, $37, $38, $39, $40, $41, $42, $43, $44, $45, $46, $47)`
	
	_, err := h.DB.Exec(query, toDate(patient.ConsentDate), patient.ConsentName, patient.ConsentCpfRg, toDate(patient.SignatureDate), patient.SignatureLocation, patient.Name, patient.AddressStreet, patient.AddressNumber, patient.AddressNeighborhood, patient.AddressCity, patient.AddressState, patient.Phone, patient.Mobile, toDate(patient.DOB), patient.Age, patient.Gender, patient.MaritalStatus, patient.Children, patient.NumChildren, patient.Profession, patient.Email, patient.EmergencyContact, patient.EmergencyPhone, patient.EmergencyOther, patient.RepetitiveEffort, patient.PhysicalActivity, patient.Smoker, patient.Alcohol, patient.MentalDisorder, patient.Religion, patient.Medication, patient.Surgery, patient.Allergies, patient.AnxietyLevel, patient.AngerLevel, patient.FearLevel, patient.SadnessLevel, patient.JoyLevel, patient.EnergyLevel, patient.MainComplaint, patient.ComplaintHistory, patient.SignsSymptoms, patient.CurrentTreatment, patient.HowFound, patient.Notes, time.Now(), time.Now())
	if err != nil {
		log.Printf("Erro ao inserir novo paciente pelo admin: %v", err)
	}
	c.Redirect(http.StatusFound, "/admin/patients")
}

func (h *AdminHandler) GetEditPatientForm(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	var patient storage.Patient
	query := `SELECT id, consent_date, consent_name, consent_cpf_rg, signature_date, signature_location, name, address_street, address_number, address_neighborhood, address_city, address_state, phone, mobile, dob, age, gender, marital_status, children, num_children, profession, email, emergency_contact, emergency_phone, emergency_other, repetitive_effort, physical_activity, smoker, alcohol, mental_disorder, religion, medication, surgery, allergies, anxiety_level, anger_level, fear_level, sadness_level, joy_level, energy_level, main_complaint, complaint_history, signs_symptoms, current_treatment, how_found, notes FROM patients WHERE id = $1`
	
	var consentDate, signatureDate, dob sql.NullTime
	var consentName, consentCpfRg, signatureLocation, addressStreet, addressNumber, addressNeighborhood, addressCity, addressState, phone, mobile, gender, maritalStatus, children, profession, email, emergencyContact, emergencyPhone, emergencyOther, repetitiveEffort, physicalActivity, smoker, alcohol, mentalDisorder, religion, medication, surgery, allergies, mainComplaint, complaintHistory, signsSymptoms, currentTreatment, howFound, notes sql.NullString
	var age, numChildren, anxietyLevel, angerLevel, fearLevel, sadnessLevel, joyLevel, energyLevel sql.NullInt64

	err := h.DB.QueryRow(query, id).Scan(
		&patient.ID, &consentDate, &consentName, &consentCpfRg, &signatureDate, &signatureLocation, 
		&patient.Name, &addressStreet, &addressNumber, &addressNeighborhood, &addressCity, &addressState, 
		&phone, &mobile, &dob, &age, &gender, &maritalStatus, &children, &numChildren, &profession, &email, 
		&emergencyContact, &emergencyPhone, &emergencyOther, 
		&repetitiveEffort, &physicalActivity, &smoker, &alcohol, &mentalDisorder, &religion, 
		&medication, &surgery, &allergies, 
		&anxietyLevel, &angerLevel, &fearLevel, &sadnessLevel, &joyLevel, &energyLevel, 
		&mainComplaint, &complaintHistory, &signsSymptoms, &currentTreatment, &howFound, &notes,
	)
	
	if err != nil {
		log.Printf("Erro ao buscar paciente para edição: %v", err)
		c.Redirect(http.StatusFound, "/admin/patients")
		return
	}

	if consentDate.Valid { patient.ConsentDate = consentDate.Time.Format("2006-01-02") }
	if consentName.Valid { patient.ConsentName = consentName.String }
	if consentCpfRg.Valid { patient.ConsentCpfRg = consentCpfRg.String }
	if signatureDate.Valid { patient.SignatureDate = signatureDate.Time.Format("2006-01-02") }
	if signatureLocation.Valid { patient.SignatureLocation = signatureLocation.String }
	if addressStreet.Valid { patient.AddressStreet = addressStreet.String }
	if addressNumber.Valid { patient.AddressNumber = addressNumber.String }
	if addressNeighborhood.Valid { patient.AddressNeighborhood = addressNeighborhood.String }
	if addressCity.Valid { patient.AddressCity = addressCity.String }
	if addressState.Valid { patient.AddressState = addressState.String }
	if phone.Valid { patient.Phone = phone.String }
	if mobile.Valid { patient.Mobile = mobile.String }
	if dob.Valid { patient.DOB = dob.Time.Format("2006-01-02") }
	if age.Valid { patient.Age = int(age.Int64) }
	if gender.Valid { patient.Gender = gender.String }
	if maritalStatus.Valid { patient.MaritalStatus = maritalStatus.String }
	if children.Valid { patient.Children = children.String }
	if numChildren.Valid { patient.NumChildren = int(numChildren.Int64) }
	if profession.Valid { patient.Profession = profession.String }
	if email.Valid { patient.Email = email.String }
	if emergencyContact.Valid { patient.EmergencyContact = emergencyContact.String }
	if emergencyPhone.Valid { patient.EmergencyPhone = emergencyPhone.String }
	if emergencyOther.Valid { patient.EmergencyOther = emergencyOther.String }
	if repetitiveEffort.Valid { patient.RepetitiveEffort = repetitiveEffort.String }
	if physicalActivity.Valid { patient.PhysicalActivity = physicalActivity.String }
	if smoker.Valid { patient.Smoker = smoker.String }
	if alcohol.Valid { patient.Alcohol = alcohol.String }
	if mentalDisorder.Valid { patient.MentalDisorder = mentalDisorder.String }
	if religion.Valid { patient.Religion = religion.String }
	if medication.Valid { patient.Medication = medication.String }
	if surgery.Valid { patient.Surgery = surgery.String }
	if allergies.Valid { patient.Allergies = allergies.String }
	if anxietyLevel.Valid { patient.AnxietyLevel = int(anxietyLevel.Int64) }
	if angerLevel.Valid { patient.AngerLevel = int(angerLevel.Int64) }
	if fearLevel.Valid { patient.FearLevel = int(fearLevel.Int64) }
	if sadnessLevel.Valid { patient.SadnessLevel = int(sadnessLevel.Int64) }
	if joyLevel.Valid { patient.JoyLevel = int(joyLevel.Int64) }
	if energyLevel.Valid { patient.EnergyLevel = int(energyLevel.Int64) }
	if mainComplaint.Valid { patient.MainComplaint = mainComplaint.String }
	if complaintHistory.Valid { patient.ComplaintHistory = complaintHistory.String }
	if signsSymptoms.Valid { patient.SignsSymptoms = signsSymptoms.String }
	if currentTreatment.Valid { patient.CurrentTreatment = currentTreatment.String }
	if howFound.Valid { patient.HowFound = howFound.String }
	if notes.Valid { patient.Notes = notes.String }

	c.HTML(http.StatusOK, "admin/patient_form.html", gin.H{
		"Title":     "Editar Paciente",
		"Action":    "/admin/patients/edit/" + idStr,
		"IsNew":     false,
		"Patient":   patient,
		"ActiveNav": "patients",
	})
}


func (h *AdminHandler) PostEditPatient(c *gin.Context) {
    idStr := c.Param("id")
    var patient storage.Patient
    if err := c.ShouldBind(&patient); err != nil {
        log.Printf("Erro ao fazer bind do formulário de edição do paciente: %v", err)
        c.Redirect(http.StatusFound, "/admin/patients")
        return
    }

    query := `UPDATE patients SET consent_date=$1, consent_name=$2, consent_cpf_rg=$3, signature_date=$4, signature_location=$5, name=$6, address_street=$7, address_number=$8, address_neighborhood=$9, address_city=$10, address_state=$11, phone=$12, mobile=$13, dob=$14, age=$15, gender=$16, marital_status=$17, children=$18, num_children=$19, profession=$20, email=$21, emergency_contact=$22, emergency_phone=$23, emergency_other=$24, repetitive_effort=$25, physical_activity=$26, smoker=$27, alcohol=$28, mental_disorder=$29, religion=$30, medication=$31, surgery=$32, allergies=$33, anxiety_level=$34, anger_level=$35, fear_level=$36, sadness_level=$37, joy_level=$38, energy_level=$39, main_complaint=$40, complaint_history=$41, signs_symptoms=$42, current_treatment=$43, how_found=$44, notes=$45, updated_at=$46 WHERE id=$47`
    
    _, err := h.DB.Exec(query, toDate(patient.ConsentDate), patient.ConsentName, patient.ConsentCpfRg, toDate(patient.SignatureDate), patient.SignatureLocation, patient.Name, patient.AddressStreet, patient.AddressNumber, patient.AddressNeighborhood, patient.AddressCity, patient.AddressState, patient.Phone, patient.Mobile, toDate(patient.DOB), patient.Age, patient.Gender, patient.MaritalStatus, patient.Children, patient.NumChildren, patient.Profession, patient.Email, patient.EmergencyContact, patient.EmergencyPhone, patient.EmergencyOther, patient.RepetitiveEffort, patient.PhysicalActivity, patient.Smoker, patient.Alcohol, patient.MentalDisorder, patient.Religion, patient.Medication, patient.Surgery, patient.Allergies, patient.AnxietyLevel, patient.AngerLevel, patient.FearLevel, patient.SadnessLevel, patient.JoyLevel, patient.EnergyLevel, patient.MainComplaint, patient.ComplaintHistory, patient.SignsSymptoms, patient.CurrentTreatment, patient.HowFound, patient.Notes, time.Now(), idStr)
    if err != nil {
        log.Printf("Erro ao atualizar paciente: %v", err)
    }
    c.Redirect(http.StatusFound, "/admin/patients")
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

// --- NOVAS Funções para Perfil e Agendamentos ---

// GetPatientProfile busca e exibe o perfil completo de um paciente, incluindo consultas.
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

    futureAppointments, err := h.getAppointmentsByTime(patientID, ">=")
    if err != nil {
        log.Printf("Erro ao buscar consultas futuras: %v", err)
    }

    pastAppointments, err := h.getAppointmentsByTime(patientID, "<")
    if err != nil {
        log.Printf("Erro ao buscar histórico de consultas: %v", err)
    }

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
        "Title":            "Perfil de " + patient.Name,
        "Patient":          patient,
        "FutureAppointments": futureAppointments,
        "PastAppointments": pastAppointments,
        "Doctors":          doctors,
		"ActiveNav": "patients",
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

    patientID, _ := strconv.Atoi(patientIDStr)
    doctorID, _ := strconv.Atoi(doctorIDStr)

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

    query := `INSERT INTO appointments (patient_id, doctor_id, start_time, end_time, status, notes, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
    _, err = h.DB.Exec(query, patientID, doctorID, startTime, endTime, status, notes, time.Now(), time.Now())
    if err != nil {
        log.Printf("Erro ao agendar nova consulta: %v", err)
    }

    c.Redirect(http.StatusFound, "/admin/patients/profile/"+patientIDStr)
}

// getAppointmentsByTime é uma função de ajuda para buscar consultas.
func (h *AdminHandler) getAppointmentsByTime(patientID int, comparison string) ([]map[string]interface{}, error) {
    query := `
        SELECT a.id, a.start_time, a.status, a.notes, u.name as doctor_name
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
        if err := rows.Scan(&app.ID, &app.StartTime, &app.Status, &notes, &doctorName); err != nil {
            continue
        }
        app.Notes = notes.String
        appointments = append(appointments, map[string]interface{}{
            "ID":         app.ID,
            "StartTime":  app.StartTime,
            "Status":     app.Status,
            "Notes":      app.Notes,
            "DoctorName": doctorName,
        })
    }
    return appointments, nil
}

// SystemMonitoring renderiza a página de monitoramento do sistema.
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

	// 1. Próximas Consultas
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

	// 2. Consultas Concluídas no período
	err = h.DB.QueryRow("SELECT COUNT(*) FROM appointments WHERE status = 'concluido' AND start_time >= $1", startDate).Scan(&data.CompletedCount)
	if err != nil {
		log.Printf("Erro ao contar consultas concluídas: %v", err)
	}

	// 3. Estatísticas de "Como nos encontrou"
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

	// 4. Estatísticas do Perfil
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

	// 5. Médias Emocionais
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
