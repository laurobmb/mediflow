package handlers

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"mediflow/storage"
)

type PortalHandler struct {
	DB *sql.DB
}

// ShowTokenLoginPage exibe a página de login por token para o paciente.
func (h *PortalHandler) ShowTokenLoginPage(c *gin.Context) {
	token := c.Param("token")
	c.HTML(http.StatusOK, "portal/token_login.html", gin.H{
		"Title": "Acesso ao Portal do Paciente",
		"Token": token,
	})
}

// ProcessTokenLogin valida o token e cria uma sessão para o paciente.
func (h *PortalHandler) ProcessTokenLogin(c *gin.Context) {
	token := c.PostForm("token")
	session := sessions.Default(c)

	var patientID int
	err := h.DB.QueryRow("SELECT id FROM patients WHERE access_token = $1", token).Scan(&patientID)

	if err != nil {
		log.Printf("Tentativa de login com token inválido: %s", token)
		c.HTML(http.StatusUnauthorized, "portal/token_login.html", gin.H{
			"Title": "Acesso ao Portal do Paciente",
			"Error": "Token inválido ou não encontrado. Verifique o link recebido.",
		})
		return
	}

	session.Set("patient_id", patientID)
	session.Save()
	c.Redirect(http.StatusFound, "/portal/consent")
}

// ShowConsentForm exibe o formulário de consentimento.
func (h *PortalHandler) ShowConsentForm(c *gin.Context) {
	session := sessions.Default(c)
	patientID := session.Get("patient_id")

	var patient storage.Patient
	err := h.DB.QueryRow("SELECT id, name, consent_given_at FROM patients WHERE id = $1", patientID).Scan(&patient.ID, &patient.Name, &patient.ConsentGivenAt)
	if err != nil {
		c.String(http.StatusInternalServerError, "Erro ao buscar dados do paciente.")
		return
	}

	c.HTML(http.StatusOK, "portal/consent_form.html", gin.H{
		"Title":   "Termo de Consentimento",
		"Patient": patient,
	})
}

// ProcessConsentForm salva os dados do formulário de consentimento.
func (h *PortalHandler) ProcessConsentForm(c *gin.Context) {
	session := sessions.Default(c)
	patientID := session.Get("patient_id")

	consentName := c.PostForm("consent_name_inline")
	consentCpfRg := c.PostForm("consent_cpf_rg_inline")
	howFound := c.PostForm("how_found")
	
	query := `UPDATE patients SET 
		consent_name = $1, consent_cpf_rg = $2, how_found = $3, 
		consent_given_at = NOW(), consent_date = NOW(), signature_date = NOW()
		WHERE id = $4`

	_, err := h.DB.Exec(query, consentName, consentCpfRg, howFound, patientID)
	if err != nil {
		log.Printf("Erro ao salvar consentimento do paciente %d: %v", patientID, err)
		// Adicionar página de erro para o paciente
		return
	}

	c.Redirect(http.StatusFound, "/portal/success")
}

// ShowSuccessPage exibe a página de agradecimento.
func (h *PortalHandler) ShowSuccessPage(c *gin.Context) {
	c.HTML(http.StatusOK, "portal/success.html", gin.H{
		"Title": "Sucesso!",
	})
}

// AuthPatientRequired é um middleware para proteger as rotas do portal.
func AuthPatientRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		if session.Get("patient_id") == nil {
			c.Redirect(http.StatusFound, "/portal/login")
			c.Abort()
			return
		}
		c.Next()
	}
}