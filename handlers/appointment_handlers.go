package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"mediflow/storage"
)

// AppointmentHandler gerencia a lógica de agendamentos.
type AppointmentHandler struct {
	DB *sql.DB
}

// GetAppointments busca e exibe a lista de agendamentos.
func (h *AppointmentHandler) GetAppointments(c *gin.Context) {
	// Exemplo de como obter o ID do usuário logado (substitua por sua lógica de sessão/JWT)
	userID := 1 // Exemplo: ID do médico logado

	// Lógica de busca no banco de dados
	rows, err := h.DB.Query("SELECT id, patient_id, doctor_id, start_time, end_time, notes FROM appointments WHERE doctor_id = $1 ORDER BY start_time ASC", userID)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"Error": "Erro ao buscar agendamentos."})
		return
	}
	defer rows.Close()

	var appointments []storage.Appointment
	for rows.Next() {
		var app storage.Appointment
		err := rows.Scan(&app.ID, &app.PatientID, &app.DoctorID, &app.StartTime, &app.EndTime, &app.Notes)
		if err != nil {
			continue // Ignora agendamentos com erro
		}
		appointments = append(appointments, app)
	}

	// Renderiza a página da agenda com os dados
	c.HTML(http.StatusOK, "agenda.html", gin.H{
		"Title":        "Agenda",
		"Appointments": appointments,
	})
}

// CreateAppointment cria um novo agendamento no banco de dados.
func (h *AppointmentHandler) CreateAppointment(c *gin.Context) {
	var newAppointment storage.Appointment

	if err := c.ShouldBind(&newAppointment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados de agendamento inválidos."})
		return
	}

	// Lógica de validação (ex: checar disponibilidade do horário)
	
	// Insere no banco de dados
	_, err := h.DB.Exec("INSERT INTO appointments (patient_id, doctor_id, start_time, end_time) VALUES ($1, $2, $3, $4)",
		newAppointment.PatientID, newAppointment.DoctorID, newAppointment.StartTime, newAppointment.EndTime)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao criar o agendamento."})
		return
	}

	// Uso do pacote fmt
	fmt.Printf("Agendamento criado para o paciente %d em %s\n", newAppointment.PatientID, newAppointment.StartTime.Format(time.RFC3339))

	c.JSON(http.StatusCreated, gin.H{"message": "Agendamento criado com sucesso!"})
}

// UpdateAppointment atualiza um agendamento existente.
func (h *AppointmentHandler) UpdateAppointment(c *gin.Context) {
	id := c.Param("id")
	var updatedData storage.Appointment
	if err := c.ShouldBindJSON(&updatedData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados de atualização inválidos."})
		return
	}

	_, err := h.DB.Exec("UPDATE appointments SET start_time = $1, end_time = $2, notes = $3 WHERE id = $4",
		updatedData.StartTime, updatedData.EndTime, updatedData.Notes, id)
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao atualizar o agendamento."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Agendamento atualizado com sucesso!"})
}

// DeleteAppointment exclui um agendamento.
func (h *AppointmentHandler) DeleteAppointment(c *gin.Context) {
	id := c.Param("id")

	_, err := h.DB.Exec("DELETE FROM appointments WHERE id = $1", id)
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao cancelar o agendamento."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Agendamento cancelado com sucesso!"})
}