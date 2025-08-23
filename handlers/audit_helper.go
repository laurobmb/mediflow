package handlers

import (
	"database/sql"
	"log"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// LogAction representa uma ação a ser registrada.
type LogAction struct {
	DB         *sql.DB
	Context    *gin.Context
	Action     string
	TargetType string
	TargetID   int
}

// AddAuditLog registra uma ação no banco de dados.
func AddAuditLog(logInfo LogAction) {
	session := sessions.Default(logInfo.Context)
	userID, okUserID := session.Get("user_id").(int)
	userName, okUserName := session.Get("user_name").(string)

	if !okUserID || !okUserName {
		log.Printf("AVISO: Tentativa de log sem usuário autenticado na sessão.")
		// Usar um ID/Nome genérico para ações sem sessão (como login falho).
		userID = 0
		userName = "Sistema"
	}

	query := `INSERT INTO audit_logs (user_id, user_name, action, target_type, target_id) 
			  VALUES ($1, $2, $3, $4, $5)`

	_, err := logInfo.DB.Exec(query, userID, userName, logInfo.Action, logInfo.TargetType, logInfo.TargetID)
	if err != nil {
		log.Printf("ERRO CRÍTICO: Falha ao registrar log de auditoria: %v", err)
	}
}

// safeAtoi é um helper para converter string para int de forma segura para o log.
func safeAtoi(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}