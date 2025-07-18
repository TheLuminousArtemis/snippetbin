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
	"github.com/theluminousartemis/snippetbin/internal/ratelimiter"
	"github.com/theluminousartemis/snippetbin/internal/store"
	"github.com/theluminousartemis/snippetbin/internal/store/cache"
	"github.com/theluminousartemis/snippetbin/ui"
)

type config struct {
	addr     string
	db       dbConfig
	redisCfg redisConfig
	rlCfg    ratelimiterConfig
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
	r.Use(middleware.Logger)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(commonHeaders)

	// //ratelimiter
	r.Use(app.RateLimiterMiddleware)
	r.Get("/ping", ping)

	// === Static files ===
	fs := http.FileServer(http.FS(ui.Files))
	r.Handle("/static/*", fs)

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
		r.Get("/account/", app.userProfile)
		r.Get("/account/password_change", app.userPasswordUpdate)
		r.Post("/account/password_change", app.userPasswordUpdatePost)
	})

	return r
}
