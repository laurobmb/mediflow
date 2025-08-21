package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"mediflow/storage"
)

// PatientHandler gerencia a lógica de pacientes para a secretária.
type PatientHandler struct {
	DB *sql.DB
}

// GetNewPatientForm renderiza a página de cadastro de novo paciente para a secretária.
func (h *PatientHandler) GetNewPatientForm(c *gin.Context) {
	// CORREÇÃO: Renderiza o novo formulário simplificado da secretária.
	c.HTML(http.StatusOK, "secretaria/new_patient.html", gin.H{
		"Title":  "Cadastrar Novo Paciente",
		"Action": "/api/v1/patient",
		"IsNew":  true,
	})
}

// CreatePatient processa a submissão do formulário de novo paciente pela secretária.
func (h *PatientHandler) CreatePatient(c *gin.Context) {
	var patient storage.Patient
	if err := c.ShouldBind(&patient); err != nil {
		log.Printf("Erro ao fazer bind do formulário do paciente (secretária): %v", err)
		c.Redirect(http.StatusFound, "/secretaria/dashboard")
		return
	}

	// CORREÇÃO: A query de inserção agora é mais simples e só inclui os campos da secretária.
	query := `
        INSERT INTO patients (
            name, address_street, address_number, address_neighborhood, address_city, address_state, 
            phone, mobile, dob, age, gender, marital_status, children, num_children, profession, email, 
            emergency_contact, emergency_phone, emergency_other, 
            created_at, updated_at
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21
        )`

	var dobValue interface{}
	if patient.DOB != "" {
		dobValue, _ = time.Parse("2006-01-02", patient.DOB)
	} else {
		dobValue = nil
	}

	_, err := h.DB.Exec(query,
		patient.Name, patient.AddressStreet, patient.AddressNumber, patient.AddressNeighborhood,
		patient.AddressCity, patient.AddressState, patient.Phone, patient.Mobile, dobValue,
		patient.Age, patient.Gender, patient.MaritalStatus, patient.Children, patient.NumChildren,
		patient.Profession, patient.Email, patient.EmergencyContact, patient.EmergencyPhone,
		patient.EmergencyOther, time.Now(), time.Now())

	if err != nil {
		log.Printf("Erro ao inserir novo paciente (secretária): %v", err)
	}

	c.Redirect(http.StatusFound, "/secretaria/dashboard")
}
