package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"regexp"

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
	err := h.DB.QueryRow("SELECT id FROM patients WHERE access_token = $1 AND consent_given_at IS NULL", token).Scan(&patientID)

	if err != nil {
		log.Printf("Tentativa de login com token inválido ou já utilizado: %s", token)
		c.HTML(http.StatusUnauthorized, "portal/token_login.html", gin.H{
			"Title": "Acesso ao Portal do Paciente",
			"Error": "Token inválido, expirado ou já utilizado. Por favor, solicite um novo link.",
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

// ProcessConsentForm salva os dados do formulário de consentimento e REGISTRA A AÇÃO NA AUDITORIA.
func (h *PortalHandler) ProcessConsentForm(c *gin.Context) {
	session := sessions.Default(c)
	patientID := session.Get("patient_id")

	patientIDInt, ok := patientID.(int)
	if !ok {
		log.Printf("Erro: patient_id na sessão não é um inteiro válido.")
		return
	}

	consentName := c.PostForm("consent_name_inline")
	consentCpfRg := c.PostForm("consent_cpf_rg_inline")
	howFound := c.PostForm("how_found")

	errors := make(map[string]string)

	nameRegex := regexp.MustCompile(`^[\p{L}´]+\s[\p{L}´\s]+$`)
	if !nameRegex.MatchString(consentName) {
		errors["Name"] = "Por favor, insira o nome completo, sem números ou caracteres especiais."
	}

	if !IsCPFValid(consentCpfRg) {
		errors["CPF"] = "O CPF informado é inválido. Por favor, verifique."
	}

	if len(errors) > 0 {
		var patient storage.Patient
		h.DB.QueryRow("SELECT name FROM patients WHERE id = $1", patientID).Scan(&patient.Name)

		c.HTML(http.StatusBadRequest, "portal/consent_form.html", gin.H{
			"Title":     "Termo de Consentimento",
			"Patient":   patient,
			"Errors":    errors,
			"NameValue": consentName,
			"CpfValue":  consentCpfRg,
		})
		return
	}

	re := regexp.MustCompile(`[^0-9]`)
	cpfClean := re.ReplaceAllString(consentCpfRg, "")

	query := `UPDATE patients SET 
		consent_name = $1, consent_cpf_rg = $2, how_found = $3, 
		consent_given_at = NOW(), consent_date = NOW(), signature_date = NOW()
		WHERE id = $4`

	_, err := h.DB.Exec(query, consentName, cpfClean, howFound, patientID)
	if err != nil {
		log.Printf("Erro ao salvar consentimento do paciente %d: %v", patientIDInt, err)
		return
	}

	logInfo := LogAction{
		DB:         h.DB,
		Context:    c,
		Action:     "Paciente forneceu o termo de consentimento através do portal.",
		TargetType: "Paciente",
		TargetID:   patientIDInt,
	}
	AddAuditLog(logInfo)

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