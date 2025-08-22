package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
	"github.com/joho/godotenv"
	"mediflow/handlers"
	"mediflow/storage"
)

type multiTemplateRenderer struct {
	templates map[string]*template.Template
}

// newMultiTemplateRenderer (VERSÃO FINAL E ROBUSTA)
// Analisa todos os templates e os prepara para uso de forma isolada.
func newMultiTemplateRenderer(templatesDir string) multiTemplateRenderer {
	r := multiTemplateRenderer{
		templates: make(map[string]*template.Template),
	}

	layouts, err := filepath.Glob(filepath.Join(templatesDir, "layouts", "*.html"))
	if err != nil {
		panic(err.Error())
	}

	// Carrega todos os "partials" (componentes reutilizáveis que começam com _)
	partials, err := filepath.Glob(filepath.Join(templatesDir, "**", "_*.html"))
	if err != nil {
		panic(err.Error())
	}

	pages, err := filepath.Glob(filepath.Join(templatesDir, "**", "*.html"))
	if err != nil {
		panic(err.Error())
	}

	funcMap := template.FuncMap{
		"seq":        func(start, end int) []int { s := []int{}; for i := start; i <= end; i++ { s = append(s, i) }; return s },
		"plus":       func(a, b int) int { return a + b },
		"minus":      func(a, b int) int { return a - b },
		"percentage": func(part, total int) float64 { if total == 0 { return 0 }; return (float64(part) / float64(total)) * 100 },
	}

	// Combina layouts e partials numa base
	baseFiles := append(layouts, partials...)

	for _, page := range pages {
		// Ignora os layouts e os partials na lista de páginas principais
		if strings.Contains(page, "layouts") || strings.HasPrefix(filepath.Base(page), "_") {
			continue
		}

		// Para cada página, combina-a com os ficheiros base
		files := append(baseFiles, page)
		
		ts := template.Must(template.New(filepath.Base(page)).Funcs(funcMap).ParseFiles(files...))
		templateName := strings.TrimPrefix(page, templatesDir+string(filepath.Separator))
		r.templates[templateName] = ts
	}

	return r
}

func (r multiTemplateRenderer) Instance(name string, data interface{}) render.Render {
	return render.HTML{
		Template: r.templates[name],
		Name:     "main.html",
		Data:     data,
	}
}

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		if session.Get("user_id") == nil {
			log.Println("Acesso não autorizado, redirecionando para /login")
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}
		c.Next()
	}
}

func RoleRequired(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		userType := session.Get("user_type")

		// Verifica se o perfil existe na sessão e se é o perfil necessário
		if userType == nil || userType.(string) != requiredRole {
			// Se não for autorizado, envia para uma página de erro "Proibido"
			log.Printf("Acesso negado para o perfil '%v' na rota que requer '%s'", userType, requiredRole)
			c.HTML(http.StatusForbidden, "layouts/error.html", gin.H{
				"Title":   "Acesso Negado",
				"Message": "Você não tem permissão para acessar esta página.",
			})
			c.Abort() // Interrompe a requisição
			return
		}
		c.Next() // Permissão concedida, continua para a próxima função
	}
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Erro ao carregar o arquivo .env: %v", err)
	}
	db, err := storage.NewDBConnection()
	if err != nil {
		log.Fatalf("Falha ao conectar ao banco de dados: %v", err)
	}
	defer db.Close()

	// Inicialização de todos os handlers
	authHandler := &handlers.AuthHandler{DB: db}
	patientHandler := &handlers.PatientHandler{DB: db}
	adminHandler := &handlers.AdminHandler{DB: db}
	secretariaHandler := &handlers.SecretariaHandler{DB: db}
    portalHandler := &handlers.PortalHandler{DB: db} // Adicionar novo handler
    terapeutaHandler := &handlers.TerapeutaHandler{DB: db} // <-- ADICIONE ESTA LINHA
	
	router := gin.Default()
	router.HTMLRender = newMultiTemplateRenderer("templates")
	store := cookie.NewStore([]byte("nova-chave-secreta-agosto-2025"))
	store.Options(sessions.Options{Path: "/", HttpOnly: true, MaxAge: 86400 * 7})
	router.Use(sessions.Sessions("mediflow_session", store))
	router.Static("/static", "./static")

	// Rotas Públicas
	router.GET("/", handlers.DashboardHandler)
	router.GET("/login", authHandler.GetLogin)
	router.POST("/login", authHandler.PostLogin)
	router.GET("/logout", authHandler.Logout)

	portal := router.Group("/portal")
    {
        portal.GET("/login", portalHandler.ShowTokenLoginPage)
        portal.GET("/login/:token", portalHandler.ShowTokenLoginPage)
        portal.POST("/login", portalHandler.ProcessTokenLogin)
        portal.GET("/success", portalHandler.ShowSuccessPage)
    }

	
	// --- Rotas Protegidas do Portal do Paciente ---
	portalProtected := router.Group("/portal", handlers.AuthPatientRequired())
	{
		portalProtected.GET("/consent", portalHandler.ShowConsentForm)
		portalProtected.POST("/consent", portalHandler.ProcessConsentForm)
	}

	// Grupos de Rotas Protegidas
	secretariaGroup := router.Group("/secretaria", AuthRequired(), RoleRequired("secretaria"))
	{
		secretariaGroup.GET("/dashboard", secretariaHandler.ViewAgenda)

		secretariaGroup.GET("/pacientes/novo", patientHandler.GetNewPatientForm)
		secretariaGroup.POST("/pacientes/novo", patientHandler.CreatePatient)

		secretariaGroup.GET("/patients", secretariaHandler.ViewPatients)
		secretariaGroup.GET("/patients/profile/:id", secretariaHandler.GetPatientProfile)
		secretariaGroup.POST("/appointments/new", secretariaHandler.PostNewAppointment)
		secretariaGroup.GET("/appointments/cancel/:id", secretariaHandler.CancelAppointment)
		secretariaGroup.GET("/patients/search", secretariaHandler.SearchPatientsAPI)
		secretariaGroup.GET("/appointments/edit/:id", secretariaHandler.GetEditAppointmentForm)
		secretariaGroup.POST("/appointments/edit/:id", secretariaHandler.PostEditAppointment)
        secretariaGroup.GET("/pacientes/token/:id", secretariaHandler.ShowPatientToken)
		secretariaGroup.GET("/appointments/mark-as-paid/:id", secretariaHandler.MarkAppointmentAsPaid)		
	}

	terapeutaGroup := router.Group("/terapeuta", AuthRequired(), RoleRequired("terapeuta"))
	{
        terapeutaGroup.GET("/dashboard", terapeutaHandler.TerapeutaDashboard)
		terapeutaGroup.GET("/pacientes/prontuario/:id", terapeutaHandler.ShowPatientRecord)
		terapeutaGroup.POST("/pacientes/prontuario/:id", terapeutaHandler.ProcessPatientRecord)
		terapeutaGroup.GET("/pacientes/search", terapeutaHandler.SearchMyPatientsAPI)
	}

	adminGroup := router.Group("/admin", AuthRequired(), RoleRequired("admin"))
	{
		adminGroup.GET("/dashboard", handlers.AdminDashboard)
		adminGroup.GET("/agenda", adminHandler.ViewAgenda) // NOVA ROTA
		adminGroup.GET("/users", adminHandler.ViewUsers)
		adminGroup.GET("/users/new", adminHandler.GetNewUserForm)
		adminGroup.POST("/users/new", adminHandler.PostNewUser)
		adminGroup.GET("/users/edit/:id", adminHandler.GetEditUserForm)
		adminGroup.POST("/users/edit/:id", adminHandler.PostEditUser)
		adminGroup.GET("/users/delete/:id", adminHandler.DeleteUser)
		adminGroup.GET("/patients", adminHandler.ViewPatients)
		adminGroup.GET("/patients/new", adminHandler.GetNewPatientForm)
		adminGroup.POST("/patients/new", adminHandler.PostNewPatient)
		adminGroup.GET("/patients/edit/:id", adminHandler.GetEditPatientForm)
		adminGroup.POST("/patients/edit/:id", adminHandler.PostEditPatient)
		adminGroup.GET("/patients/delete/:id", adminHandler.DeletePatient)
		adminGroup.GET("/patients/search", adminHandler.SearchPatientsAPI)
		adminGroup.GET("/patients/profile/:id", adminHandler.GetPatientProfile)
		adminGroup.POST("/appointments/new", adminHandler.PostNewAppointment)
		adminGroup.GET("/monitoring", adminHandler.SystemMonitoring)

		adminGroup.GET("/appointments/edit/:id", adminHandler.GetEditAppointmentForm)
		adminGroup.POST("/appointments/edit/:id", adminHandler.PostEditAppointment)
		adminGroup.GET("/appointments/cancel/:id", adminHandler.CancelAppointment)
		adminGroup.GET("/appointments/mark-as-paid/:id", adminHandler.MarkAppointmentAsPaid)		
	}

	api := router.Group("/api/v1", AuthRequired())
	{
		api.POST("/patient", patientHandler.CreatePatient)
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Servidor iniciado em http://localhost:%s", port)
	if err := router.Run("0.0.0.0:" + port); err != nil {
		log.Fatalf("Erro ao iniciar o servidor: %v", err)
	}
}
