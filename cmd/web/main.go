package main

import (
	"crypto/tls"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/postgresstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	"github.com/go-playground/validator/v10"
	"github.com/theluminousartemis/snippetbin/internal/db"
	"github.com/theluminousartemis/snippetbin/internal/env"
	"github.com/theluminousartemis/snippetbin/internal/ratelimiter"
	"github.com/theluminousartemis/snippetbin/internal/store"
	"github.com/theluminousartemis/snippetbin/internal/store/cache"
)

var validate *validator.Validate

func init() {
	validate = validator.New(validator.WithRequiredStructEnabled())
}

func main() {
	cfg := config{
		addr: env.GetString("ADDR", ":4000"),
		// staticDir: env.GetString("STATIC_DIR", "./ui/static"),
		db: dbConfig{
			addr:         env.GetString("DB_ADDR", "postgres://admin:adminpassword@localhost:5432/snippetbin?sslmode=disable"),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 30),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 30),
			maxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
		},
		redisCfg: redisConfig{
			addr:     env.GetString("REDIS_ADDR", "localhost:6379"),
			password: env.GetString("REDIS_PASSWORD", ""),
			db:       env.GetInt("REDIS_DB", 0),
		},
		rlCfg: ratelimiterConfig{
			RequestsPerTimeFrame: env.GetInt("RATELIMITER_REQUESTS_PER_TIME_FRAME", 20),
			Timeframe:            2 * time.Minute,
			Enabled:              env.GetBool("RATELIMITER_ENABLED", true),
		},
	}

	//logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{AddSource: true}))

	//database
	db, err := db.New(cfg.db.addr, cfg.db.maxOpenConns, cfg.db.maxIdleConns, cfg.db.maxIdleTime)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	defer db.Close()
	logger.Info("database connection pool established")
	store := store.NewPostgresStore(db)

	//template cache
	templateCache, err := newTemplateCache()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	formDecoder := form.NewDecoder()
	sessionManager := scs.New()
	sessionManager.Store = postgresstore.New(db)
	sessionManager.Lifetime = 12 * time.Hour

	//cache
	redisClient := cache.NewRedisClient(cfg.redisCfg.addr, cfg.redisCfg.password, cfg.redisCfg.db)
	logger.Info("redis client established")
	cache := cache.NewRedisStore(redisClient)

	//ratelimiter
	ratelimiter := ratelimiter.NewRedisFixedWindowRateLimiter(
		cache, cfg.rlCfg.RequestsPerTimeFrame, cfg.rlCfg.Timeframe,
	)

	app := &application{
		logger:         logger,
		store:          store,
		cache:          cache,
		templateCache:  templateCache,
		formDecoder:    formDecoder,
		sessionManager: sessionManager,
		rateLimiter:    ratelimiter,
	}

	tlsConfig := &tls.Config{
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	srv := &http.Server{
		Addr:         cfg.addr,
		Handler:      app.routes(),
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
		TLSConfig:    tlsConfig,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	app.logger.Info("starting server", slog.Any("addr", srv.Addr))
	err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
	app.logger.Error(err.Error())
	os.Exit(1)
}
