package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/dhinogz/lenslocked/controllers"
	"github.com/dhinogz/lenslocked/migrations"
	"github.com/dhinogz/lenslocked/models"
	"github.com/dhinogz/lenslocked/templates"
	"github.com/dhinogz/lenslocked/views"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/csrf"
	"github.com/joho/godotenv"
)

type config struct {
	PSQL models.PostgresConfig
	SMTP models.SMTPConfig
	CSRF struct {
		Key    string
		Secure bool
	}
	Server struct {
		Address string
	}
}

func loadEnvConfig() (config, error) {
	var cfg config
	err := godotenv.Load()
	if err != nil {
		return cfg, err
	}
	// TODO: Read PSQL values from an ENV variable
	cfg.PSQL = models.DefaultPostgresConfig()

	// TODO: SMTP

	cfg.SMTP.Host = os.Getenv("SMTP_HOST")
	portStr := os.Getenv("SMTP_PORT")
	cfg.SMTP.Port, err = strconv.Atoi(portStr)
	if err != nil {
		return cfg, err
	}
	cfg.SMTP.Username = os.Getenv("SMTP_USERNAME")
	cfg.SMTP.Password = os.Getenv("SMTP_PASSWORD")

	// TODO: Read the csrf values from an env variable
	cfg.CSRF.Key = "gFvi45R4fy5xNBlnEeZtQbfAVCYEIAUX"
	cfg.CSRF.Secure = false

	// TODO: Read the server values from an ENV variable
	cfg.Server.Address = ":3000"

	return cfg, nil
}

func main() {
	// Set up config
	cfg, err := loadEnvConfig()
	if err != nil {
		panic(err)
	}

	// Set up database
	db, err := models.Open(cfg.PSQL)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = models.MigrateFS(db, migrations.FS, ".")
	if err != nil {
		panic(err)
	}

	// Set up services
	userService := &models.UserService{
		DB: db,
	}
	sessionService := &models.SessionService{
		DB: db,
	}
	pwResetService := &models.PasswordResetService{
		DB: db,
	}
	emailService, err := models.NewEmailService(cfg.SMTP)
	if err != nil {
		panic(err)
	}

	// Set up middleware
	csrfMw := csrf.Protect(
		[]byte(cfg.CSRF.Key),
		csrf.Secure(cfg.CSRF.Secure),
	)
	umw := controllers.UserMiddleware{
		SessionService: sessionService,
	}

	// Set up controllers
	usersC := controllers.Users{
		UserService:          userService,
		SessionService:       sessionService,
		PasswordResetService: pwResetService,
		EmailService:         emailService,
	}
	usersC.Templates.New = views.Must(views.ParseFS(
		templates.FS,
		"signup.html", "tailwind.html",
	))
	usersC.Templates.SignIn = views.Must(views.ParseFS(
		templates.FS,
		"signin.html", "tailwind.html",
	))
	usersC.Templates.ForgotPassword = views.Must(views.ParseFS(
		templates.FS,
		"forgot-pw.html", "tailwind.html",
	))
	usersC.Templates.CheckYourEmail = views.Must(views.ParseFS(
		templates.FS,
		"check-your-email.html", "tailwind.html",
	))
	usersC.Templates.ResetPassword = views.Must(views.ParseFS(
		templates.FS,
		"reset-pw.html", "tailwind.html",
	))

	// Set up router
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(csrfMw)
	r.Use(umw.SetUser)

	r.Get("/", controllers.StaticHandler(views.Must(views.ParseFS(
		templates.FS,
		"home.html", "tailwind.html",
	))))
	r.Get("/contact", controllers.StaticHandler(views.Must(views.ParseFS(
		templates.FS,
		"contact.html", "tailwind.html",
	))))
	r.Get("/faq", controllers.FAQ(views.Must(views.ParseFS(
		templates.FS,
		"faq.html", "tailwind.html",
	))))
	r.Get("/check-email", controllers.StaticHandler(views.Must(views.ParseFS(
		templates.FS,
		"check-your-email.html", "tailwind.html",
	))))
	r.Get("/forgot-pw", usersC.ForgotPassword)
	r.Post("/forgot-pw", usersC.ProcessForgotPassword)
	r.Get("/reset-pw", usersC.ResetPassword)
	r.Post("/reset-pw", usersC.ProcessResetPassword)

	r.Route("/signup", func(r chi.Router) {
		r.Get("/", usersC.New)
		r.Post("/", usersC.Create)
	})

	r.Route("/signin", func(r chi.Router) {
		r.Get("/", usersC.SignIn)
		r.Post("/", usersC.ProcessSignIn)
	})

	r.Post("/signout", usersC.ProcessSignOut)

	r.Route("/users/me", func(r chi.Router) {
		r.Use(umw.RequireUser)
		r.Get("/", usersC.CurrentUser)
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Page not found", http.StatusNotFound)
	})

	fmt.Printf("Starting server on %s...\n", cfg.Server.Address)
	err = http.ListenAndServe(":3000", r)
	if err != nil {
		panic(err)
	}
}
