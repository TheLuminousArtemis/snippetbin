package main

import (
	"html/template"
	"log/slog"
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/form/v4"
	"github.com/theluminousartemis/letsgo_snippetbox/internal/ratelimiter"
	"github.com/theluminousartemis/letsgo_snippetbox/internal/store"
	"github.com/theluminousartemis/letsgo_snippetbox/internal/store/cache"
	"github.com/theluminousartemis/letsgo_snippetbox/ui"
)

type config struct {
	addr      string
	staticDir string
	db        dbConfig
	redisCfg  redisConfig
	rlCfg     ratelimiterConfig
}

type ratelimiterConfig struct {
	RequestsPerTimeFrame int
	Timeframe            time.Duration
	Enabled              bool
}

type dbConfig struct {
	addr         string
	maxOpenConns int
	maxIdleConns int
	maxIdleTime  string
}

type redisConfig struct {
	addr     string
	password string
	db       int
}

type application struct {
	logger         *slog.Logger
	store          store.Storage
	cache          cache.Storage
	templateCache  map[string]*template.Template
	formDecoder    *form.Decoder
	sessionManager *scs.SessionManager
	rateLimiter    ratelimiter.Limiter
}

func (app *application) routes() http.Handler {
	r := chi.NewRouter()

	// === Global middleware ===
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(app.recoverPanic)
	r.Use(app.logRequest)
	r.Use(commonHeaders)

	//ratelimiter
	r.Use(app.RateLimiterMiddleware)

	// === Static files ===
	// r.Get("/static", http.FileServerFS(ui.Files))
	// fs := http.FileServer(&neuteredFileSystem{http.Dir("./ui/static/")})
	r.Handle("/static/*", http.StripPrefix("/static", http.FileServerFS(ui.Files)))

	// === Public routes ===
	r.Group(func(r chi.Router) {
		r.Use(app.sessionManager.LoadAndSave)
		r.Use(noSurf)
		r.Use(app.authenticate)

		r.Get("/health", app.health)

		r.Get("/", app.home)
		r.Get("/snippet/view/{id}", app.snippetView)

		// User auth routes
		r.Get("/user/signup", app.userSignup)
		r.Post("/user/signup", app.userSignupPost)
		r.Get("/user/login", app.userLogin)
		r.Post("/user/login", app.userLoginPost)

		// About page
		r.Get("/about", app.about)
	})

	// === Protected routes ===
	r.Group(func(r chi.Router) {
		r.Use(app.sessionManager.LoadAndSave)
		r.Use(noSurf)
		r.Use(app.authenticate)
		r.Use(app.requireAuthentication)

		r.Get("/snippet/create", app.snippetCreate)
		r.Post("/snippet/create", app.snippetCreatePost)
		r.Post("/user/logout", app.userLogoutPost)
	})

	return r
}
