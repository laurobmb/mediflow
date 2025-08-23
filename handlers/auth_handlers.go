package handlers

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"mediflow/storage"
)

// AuthHandler gerencia a lógica de autenticação.
type AuthHandler struct {
	DB *sql.DB
}

// LoginData é o modelo para os dados de login.
type LoginData struct {
	Email    string `form:"email" binding:"required,email"`
	Password string `form:"password" binding:"required"`
}

// GetLogin renderiza a página de login ou redireciona se já estiver logado.
func (h *AuthHandler) GetLogin(c *gin.Context) {
	session := sessions.Default(c)
	if session.Get("user_id") != nil {
		userType := session.Get("user_type")
		switch userType {
		case "terapeuta":
			c.Redirect(http.StatusFound, "/terapeuta/dashboard")
		case "secretaria":
			c.Redirect(http.StatusFound, "/secretaria/dashboard")
		case "admin":
			c.Redirect(http.StatusFound, "/admin/dashboard")
		default:
			c.Redirect(http.StatusFound, "/login")
		}
		return
	}

	// NOVO: Verifica se há uma mensagem de erro na sessão (flash message)
	var errorMessage string
	flashes := session.Flashes("error")
	if len(flashes) > 0 {
		errorMessage = flashes[0].(string)
	}
	// Salva a sessão para limpar a flash message
	session.Save()

	c.HTML(http.StatusOK, "auth/login.html", gin.H{
		"Title": "Login",
		"Error": errorMessage,
	})
}

// PostLogin processa o formulário de login e cria uma sessão.
func (h *AuthHandler) PostLogin(c *gin.Context) {
	session := sessions.Default(c)
	var loginData LoginData

	if err := c.ShouldBind(&loginData); err != nil {
		session.AddFlash("Email e senha são obrigatórios.", "error")
		session.Save()
		c.Redirect(http.StatusFound, "/login")
		return
	}

	var user storage.User
	query := "SELECT id, name, email, password_hash, user_type FROM users WHERE email = $1"
	err := h.DB.QueryRow(query, loginData.Email).Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.UserType)

	if err != nil || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(loginData.Password)) != nil {
		// NOVO: Guarda o erro na sessão e redireciona
		session.AddFlash("Credenciais inválidas. Tente novamente.", "error")
		session.Save()
		c.Redirect(http.StatusFound, "/login")
		return
	}

	session.Set("user_id", user.ID)
	session.Set("user_type", user.UserType)
	session.Set("user_name", user.Name)
	if err := session.Save(); err != nil {
		log.Printf("Erro ao salvar sessão: %v", err)
		session.AddFlash("Ocorreu um erro interno. Tente novamente.", "error")
		session.Save()
		c.Redirect(http.StatusFound, "/login")
		return
	}

	logInfo := LogAction{DB: h.DB, Context: c, Action: "Login bem-sucedido"}
	AddAuditLog(logInfo)
	
	switch user.UserType {
	case "terapeuta":
		c.Redirect(http.StatusFound, "/terapeuta/dashboard")
	case "secretaria":
		c.Redirect(http.StatusFound, "/secretaria/dashboard")
	case "admin":
		c.Redirect(http.StatusFound, "/admin/dashboard")
	default:
		c.Redirect(http.StatusFound, "/login")
	}
}

// Logout limpa a sessão do usuário.
func (h *AuthHandler) Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Options(sessions.Options{
		MaxAge: -1,
	})
	session.Save()
	c.Redirect(http.StatusFound, "/login")
}
