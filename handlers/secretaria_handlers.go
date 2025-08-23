package handlers

import (
	"database/sql"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"
	"fmt"
	
	"github.com/gin-gonic/gin"
	"mediflow/storage"
)

// SecretariaHandler gerencia a lógica do painel da secretária.
type SecretariaHandler struct {
	DB *sql.DB
}

// AppointmentDetails é uma struct para organizar os dados da agenda para o template.
type AppointmentDetails struct {
	ID          int
	StartTime   time.Time
	PatientName string
	DoctorName  string
	Status      string
	PatientID   int
}

// DaySchedule é uma struct para organizar os dados da agenda para o template.
type DaySchedule struct {
	Date         time.Time
	Appointments []AppointmentDetails
}

// ViewAgenda renderiza o novo dashboard de agenda da secretária.
func (h *SecretariaHandler) ViewAgenda(c *gin.Context) {
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
		log.Printf("Erro ao buscar agendamentos para a agenda: %v", err)
		c.HTML(http.StatusInternalServerError, "layouts/error.html", gin.H{"Title": "Erro", "Message": "Não foi possível carregar a agenda."})
		return
	}
	defer rows.Close()

	appointmentsByDay := make(map[string][]AppointmentDetails)
	for rows.Next() {
		var app AppointmentDetails
		if err := rows.Scan(&app.ID, &app.StartTime, &app.Status, &app.PatientName, &app.DoctorName, &app.PatientID); err != nil {
			continue
		}
		dayKey := app.StartTime.Format("2006-01-02")
		appointmentsByDay[dayKey] = append(appointmentsByDay[dayKey], app)
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

	c.HTML(http.StatusOK, "secretaria/secretaria_dashboard.html", gin.H{
		"Title":        "Agenda da Clínica",
		"WeekSchedule": weekSchedule,
		"PrevWeekLink": "/secretaria/dashboard?date=" + startOfWeek.AddDate(0, 0, -7).Format("2006-01-02"),
		"NextWeekLink": "/secretaria/dashboard?date=" + startOfWeek.AddDate(0, 0, 7).Format("2006-01-02"),
		"TodayLink":    "/secretaria/dashboard",
		"ActiveNav":    "agenda",
	})
}

// handlers/secretaria_handlers.go

func (h *SecretariaHandler) ViewPatients(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}
	searchTerm := c.Query("search")
	pageSize := 10
	offset := (page - 1) * pageSize

	query := `SELECT id, name, email, phone FROM patients WHERE name ILIKE $1 ORDER BY name ASC LIMIT $2 OFFSET $3`
	countQuery := `SELECT COUNT(*) FROM patients WHERE name ILIKE $1`

	var totalRecords int
	err = h.DB.QueryRow(countQuery, "%"+searchTerm+"%").Scan(&totalRecords)
	if err != nil {
		log.Printf("Erro ao contar pacientes (secretária): %v", err)
		c.HTML(http.StatusInternalServerError, "layouts/error.html", gin.H{"Title": "Erro", "Message": "Não foi possível carregar os dados dos pacientes."})
		return
	}
	totalPages := int(math.Ceil(float64(totalRecords) / float64(pageSize)))

	rows, err := h.DB.Query(query, "%"+searchTerm+"%", pageSize, offset)
	if err != nil {
		log.Printf("Erro ao buscar pacientes (secretária): %v", err)
		c.HTML(http.StatusInternalServerError, "layouts/error.html", gin.H{"Title": "Erro", "Message": "Não foi possível carregar os pacientes."})
		return
	}
	defer rows.Close()

	var patients []storage.Patient
	for rows.Next() {
		var p storage.Patient
		// CORREÇÃO: Usar sql.NullString para campos que podem ser nulos
		var email, phone sql.NullString

		if err := rows.Scan(&p.ID, &p.Name, &email, &phone); err != nil {
			log.Printf("Erro ao escanear paciente (secretária): %v", err)
			continue
		}
		
		// Se o valor do banco não for nulo, atribui à struct.
		if email.Valid {
			p.Email = email.String
		}
		if phone.Valid {
			p.Phone = phone.String
		}

		patients = append(patients, p)
	}

	c.HTML(http.StatusOK, "secretaria/view_patients.html", gin.H{
		"Title":       "Consultar Pacientes",
		"Patients":    patients,
		"TotalPages":  totalPages,
		"CurrentPage": page,
		"SearchTerm":  searchTerm,
		"ActiveNav":   "patients",
	})
}

// GetPatientProfile renderiza o perfil de agendamento de um paciente.
func (h *SecretariaHandler) GetPatientProfile(c *gin.Context) {
	idStr := c.Param("id")
	patientID, err := strconv.Atoi(idStr)
	if err != nil {
		c.HTML(http.StatusBadRequest, "layouts/error.html", gin.H{"Title": "Erro", "Message": "ID de paciente inválido."})
		return
	}

	var patient storage.Patient
	query := "SELECT id, name, consent_given_at, access_token FROM patients WHERE id = $1"
	
	// CORREÇÃO: Adicionado o 'patientID' que estava faltando para o parâmetro $1
	err = h.DB.QueryRow(query, patientID).Scan(&patient.ID, &patient.Name, &patient.ConsentGivenAt, &patient.AccessToken)
	
	if err != nil {
		if err == sql.ErrNoRows {
			c.HTML(http.StatusNotFound, "layouts/error.html", gin.H{"Title": "Erro", "Message": "Paciente não encontrado."})
		} else {
			log.Printf("Erro ao buscar perfil do paciente (secretária): %v", err)
			c.HTML(http.StatusInternalServerError, "layouts/error.html", gin.H{"Title": "Erro", "Message": "Erro ao buscar dados do paciente."})
		}
		return
	}

	futureAppointments, _ := getAppointmentsByTime(h.DB, patientID, ">=")
	pastAppointments, _ := getAppointmentsByTime(h.DB, patientID, "<")

	var doctors []storage.User
	rows, _ := h.DB.Query("SELECT id, name FROM users WHERE user_type = 'terapeuta' ORDER BY name ASC")
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var doc storage.User
			if err := rows.Scan(&doc.ID, &doc.Name); err == nil {
				doctors = append(doctors, doc)
			}
		}
	}

	c.HTML(http.StatusOK, "secretaria/patient_profile.html", gin.H{
		"Title":              "Agendamentos de " + patient.Name,
		"Patient":            patient,
		"FutureAppointments": futureAppointments,
		"PastAppointments":   pastAppointments,
		"Doctors":            doctors,
		"ActiveNav":          "patients",
	})
}

// PostNewAppointment cria uma nova consulta.
func (h *SecretariaHandler) PostNewAppointment(c *gin.Context) {
	patientIDStr := c.PostForm("patient_id")
	doctorIDStr := c.PostForm("doctor_id")
	dateStr := c.PostForm("appointment_date")
	timeStr := c.PostForm("start_time")

	patientID, _ := strconv.Atoi(patientIDStr)
	doctorID, _ := strconv.Atoi(doctorIDStr)

	startTime, err := time.Parse("2006-01-02 15:04", dateStr+" "+timeStr)
	if err != nil {
		log.Printf("Erro ao converter data/hora do agendamento (secretária): %v", err)
		c.Redirect(http.StatusFound, "/secretaria/patients/profile/"+patientIDStr)
		return
	}
	endTime := startTime.Add(1 * time.Hour)

	var status string
	if startTime.Before(time.Now()) {
		status = "concluido"
	} else {
		status = "agendado"
	}

	query := `INSERT INTO appointments (patient_id, doctor_id, start_time, end_time, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err = h.DB.Exec(query, patientID, doctorID, startTime, endTime, status, time.Now(), time.Now())
	if err != nil {
		log.Printf("Erro ao agendar nova consulta (secretária): %v", err)
	}

	c.Redirect(http.StatusFound, "/secretaria/patients/profile/"+patientIDStr)
}

// CancelAppointment (Desmarcar) atualiza o status de uma consulta para 'cancelado'.
func (h *SecretariaHandler) CancelAppointment(c *gin.Context) {
	appointmentID := c.Param("id")
	patientID := c.Query("patient_id")

	log.Printf("[DEBUG] Tentando cancelar consulta ID: %s para paciente ID: %s", appointmentID, patientID)

	query := `UPDATE appointments SET status = 'cancelado', updated_at = $1 WHERE id = $2`
	result, err := h.DB.Exec(query, time.Now(), appointmentID)
	if err != nil {
		log.Printf("ERRO DE BANCO DE DADOS ao cancelar consulta: %v", err)
		c.HTML(http.StatusInternalServerError, "layouts/error.html", gin.H{
			"Title":   "Erro Interno",
			"Message": "Ocorreu um erro ao tentar desmarcar a consulta.",
		})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		log.Printf("AVISO: Nenhuma consulta foi desmarcada. ID da consulta (%s) pode não existir.", appointmentID)
	} else {
		log.Printf("SUCESSO: Consulta ID %s desmarcada.", appointmentID)
	}

	c.Redirect(http.StatusFound, "/secretaria/patients/profile/"+patientID)
}

// SearchPatientsAPI é o endpoint para o autopreenchimento da busca da secretária.
func (h *SecretariaHandler) SearchPatientsAPI(c *gin.Context) {
	term := c.Query("term")
	if term == "" {
		c.JSON(http.StatusOK, []string{})
		return
	}

	query := `SELECT name FROM patients WHERE name ILIKE $1 ORDER BY name ASC LIMIT 10`
	rows, err := h.DB.Query(query, term+"%")
	if err != nil {
		log.Printf("Erro na busca por autocompletar (secretária): %v", err)
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

// GetEditAppointmentForm renderiza o formulário para editar uma consulta.
func (h *SecretariaHandler) GetEditAppointmentForm(c *gin.Context) {
	appointmentIDStr := c.Param("id")
	appointmentID, _ := strconv.Atoi(appointmentIDStr)

	var app storage.Appointment
	var patientName string
	query := `SELECT a.id, a.patient_id, a.doctor_id, a.start_time, p.name 
              FROM appointments a JOIN patients p ON a.patient_id = p.id 
              WHERE a.id = $1`
	err := h.DB.QueryRow(query, appointmentID).Scan(&app.ID, &app.PatientID, &app.DoctorID, &app.StartTime, &patientName)
	if err != nil {
		log.Printf("Erro ao buscar consulta para edição: %v", err)
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

	c.HTML(http.StatusOK, "secretaria/edit_appointment.html", gin.H{
		"Title":       "Editar Agendamento",
		"Appointment": app,
		"PatientID":   app.PatientID,
		"PatientName": patientName,
		"Doctors":     doctors,
		"ActiveNav":   "patients",
	})
}

// PostEditAppointment processa a atualização de uma consulta.
func (h *SecretariaHandler) PostEditAppointment(c *gin.Context) {
	appointmentIDStr := c.Param("id")
	patientIDStr := c.Query("patient_id")

	doctorIDStr := c.PostForm("doctor_id")
	dateStr := c.PostForm("appointment_date")
	timeStr := c.PostForm("start_time")

	doctorID, _ := strconv.Atoi(doctorIDStr)

	startTime, err := time.Parse("2006-01-02 15:04", dateStr+" "+timeStr)
	if err != nil {
		log.Printf("Erro ao converter data/hora na edição: %v", err)
		c.Redirect(http.StatusFound, "/secretaria/patients/profile/"+patientIDStr)
		return
	}
	endTime := startTime.Add(1 * time.Hour)

	query := `UPDATE appointments SET doctor_id = $1, start_time = $2, end_time = $3, updated_at = $4 WHERE id = $5`
	_, err = h.DB.Exec(query, doctorID, startTime, endTime, time.Now(), appointmentIDStr)
	if err != nil {
		log.Printf("Erro ao atualizar consulta: %v", err)
	}

	c.Redirect(http.StatusFound, "/secretaria/patients/profile/"+patientIDStr)
}

// Função de ajuda para buscar consultas (agora com preço e status de pagamento)
func getAppointmentsByTime(db *sql.DB, patientID int, comparison string) ([]map[string]interface{}, error) {
	query := `
		SELECT a.id, a.start_time, a.status, a.notes, u.name as doctor_name, a.price, a.payment_status
		FROM appointments a
		JOIN users u ON a.doctor_id = u.id
		WHERE a.patient_id = $1 AND a.start_time ` + comparison + ` $2
		ORDER BY a.start_time DESC
	`
	rows, err := db.Query(query, patientID, time.Now())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var appointments []map[string]interface{}
	for rows.Next() {
		var appID int
		var startTime time.Time
		var status string
		var doctorName sql.NullString
		var notes sql.NullString
		var price sql.NullFloat64
		var paymentStatus sql.NullString

		if err := rows.Scan(&appID, &startTime, &status, &notes, &doctorName, &price, &paymentStatus); err != nil {
			log.Printf("Erro ao escanear linha de consulta: %v", err)
			continue
		}

		appointmentData := map[string]interface{}{
			"ID":            appID,
			"StartTime":     startTime,
			"Status":        status,
			"DoctorName":    doctorName.String,
			"Notes":         notes.String,
			"Price":         price.Float64,
			"PaymentStatus": paymentStatus.String,
		}
		appointments = append(appointments, appointmentData)
	}
	return appointments, nil
}

// handlers/secretaria_handlers.go
// Adicione esta nova função

// ShowPatientToken exibe o link de acesso único para o paciente recém-criado.
func (h *SecretariaHandler) ShowPatientToken(c *gin.Context) {
	patientID := c.Param("id")

	var name, token string
	err := h.DB.QueryRow("SELECT name, access_token FROM patients WHERE id = $1", patientID).Scan(&name, &token)
	if err != nil {
		// Tratar erro, talvez redirecionar para a lista de pacientes
		c.Redirect(http.StatusFound, "/secretaria/patients")
		return
	}

	// Constrói a URL completa para o portal do paciente
	baseURL := "http://" + c.Request.Host // Ex: http://localhost:8080
	portalURL := fmt.Sprintf("%s/portal/login/%s", baseURL, token)

	c.HTML(http.StatusOK, "secretaria/show_token.html", gin.H{
		"Title":     "Link de Acesso do Paciente",
		"PatientName": name,
		"PortalURL": portalURL,
		"ActiveNav": "new_patient",
	})
}

// handlers/secretaria_handlers.go

// MarkAppointmentAsPaid atualiza o status de pagamento de uma consulta para 'pago'.
func (h *SecretariaHandler) MarkAppointmentAsPaid(c *gin.Context) {
    appointmentID := c.Param("id")
    patientID := c.Query("patient_id") // Precisamos saber para qual paciente voltar

    query := `UPDATE appointments SET payment_status = 'pago', updated_at = $1 WHERE id = $2`
    _, err := h.DB.Exec(query, time.Now(), appointmentID)
    if err != nil {
        log.Printf("ERRO ao marcar consulta como paga (secretária): %v", err)
    } else {
		// ======================================================
		// ADICIONANDO O REGISTRO DE AUDITORIA
		// ======================================================
		logInfo := LogAction{
			DB:         h.DB,
			Context:    c,
			Action:     fmt.Sprintf("Marcou a consulta #%s como PAGA para o paciente #%s", appointmentID, patientID),
			TargetType: "Agendamento",
			TargetID:   safeAtoi(appointmentID),
		}
		AddAuditLog(logInfo)
		// ======================================================
	}

    // Redireciona de volta para a página de perfil do paciente
    c.Redirect(http.StatusFound, "/secretaria/patients/profile/"+patientID)
}