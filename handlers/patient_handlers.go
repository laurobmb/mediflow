package handlers

import (
	"crypto/rand" // Adicionar import
	"database/sql"
	"log"
	"net/http"
	"time"
	"encoding/hex" // Adicionar import
	"strconv" // <-- ADICIONE ESTA LINHA
	
	"github.com/gin-gonic/gin"
	"mediflow/storage"
)

// Nova função para gerar um token seguro
func generateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// PatientHandler gerencia a lógica de pacientes para a secretária.
type PatientHandler struct {
	DB *sql.DB
}

// GetNewPatientForm renderiza a página de cadastro de novo paciente para a secretária.
func (h *PatientHandler) GetNewPatientForm(c *gin.Context) {
	c.HTML(http.StatusOK, "secretaria/new_patient.html", gin.H{
		"Title":     "Cadastrar Novo Paciente",
		"Action":    "/api/v1/patient",
		"IsNew":     true,
		"ActiveNav": "new_patient",
	})
}


// CreatePatient processa a submissão do formulário e agora gera um token.
func (h *PatientHandler) CreatePatient(c *gin.Context) {
	var patient storage.Patient
	if err := c.ShouldBind(&patient); err != nil {
		log.Printf("Erro ao fazer bind do formulário do paciente (secretária): %v", err)
		c.Redirect(http.StatusFound, "/secretaria/dashboard")
		return
	}

	// Gera o token de acesso único
	token, err := generateSecureToken(32)
	if err != nil {
		log.Printf("Erro ao gerar token seguro: %v", err)
		c.Redirect(http.StatusFound, "/secretaria/dashboard")
		return
	}

	// Query de inserção agora inclui o access_token
	query := `
		INSERT INTO patients (
			name, email, phone, mobile, access_token, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`

	var patientID int
	err = h.DB.QueryRow(query,
		patient.Name, patient.Email, patient.Phone, patient.Mobile, 
		token, time.Now(), time.Now(),
	).Scan(&patientID)

	if err != nil {
		log.Printf("Erro ao inserir novo paciente (secretária): %v", err)
		c.Redirect(http.StatusFound, "/secretaria/dashboard")
		return
	}

	// Redireciona para a nova página que exibe o token e o link
	c.Redirect(http.StatusFound, "/secretaria/pacientes/token/"+strconv.Itoa(patientID))
}