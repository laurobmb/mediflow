package handlers

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

// DashboardHandler redireciona o usuário para a página de login.
func DashboardHandler(c *gin.Context) {
	c.Redirect(http.StatusFound, "/login")
}

// TerapeutaDashboard renderiza a dashboard do médico.
func TerapeutaDashboard(c *gin.Context) {
	// CORREÇÃO: Usar o caminho completo do template
	c.HTML(http.StatusOK, "terapeuta/terapeuta_dashboard.html", gin.H{"Title": "Dashboard do Médico"})
}

// SecretariaDashboard foi movida para secretaria_handlers.go como ViewAgenda.
// Esta função agora pode ser removida ou mantida como um redirecionamento, se necessário.
func SecretariaDashboard(c *gin.Context) {
    // Redireciona para a nova rota de agenda para garantir compatibilidade
    c.Redirect(http.StatusFound, "/secretaria/agenda")
}

// AdminDashboard renderiza a dashboard do administrador.
func AdminDashboard(c *gin.Context) {
	// CORREÇÃO: Usar o caminho completo do template
	c.HTML(http.StatusOK, "admin/admin_dashboard.html", gin.H{"Title": "Dashboard do Admin"})
}
