package main

import (
	"Obsonarium-backend/internal/handlers/auth"
	"Obsonarium-backend/internal/repositories"
	"Obsonarium-backend/internal/services"
	"Obsonarium-backend/internal/utils/jsonutils"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog"
)

type config struct {
	port int
	Env  string
	DB   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
}

type dependencies struct {
	logger                  zerolog.Logger
	JSONutils               jsonutils.JSONutils
	AuthService             services.AuthService
	RetailerProductsService services.RetailerProductsService
	RetailersService        services.RetailersService
	ProductService          services.ProductService
	CartService             services.CartService
	UserAddressesService    services.UserAddressesService
	UploadService           *services.UploadService
}

type application struct {
	config      config
	shared_deps dependencies
}

func main() {
	var cfg config
	flag.IntVar(&cfg.port, "port", 8000, "API server port")
	flag.StringVar(&cfg.Env, "env", "prod", "Environment (development|staging|production)")
	flag.StringVar(&cfg.DB.dsn, "db-dsn", os.Getenv("OBSONARIUM_DB_DSN"), "Database connection string")

	flag.IntVar(&cfg.DB.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.DB.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.DB.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time")
	flag.Parse()
	logger := zerolog.New(os.Stdout).Level(zerolog.InfoLevel).With().Timestamp().Logger()

	db, err := openDB(cfg)
	if err != nil {
		logger.Fatal().Msg(err.Error())
	}

	defer db.Close()

	app := &application{
		config: cfg,
		shared_deps: dependencies{
			logger:                  logger,
			JSONutils:               jsonutils.NewJSONutils(),
			AuthService:             *services.NewAuthService(repositories.NewUsersRepo(db), repositories.NewRetailersRepo(db)),
			RetailerProductsService: *services.NewRetailerProductsService(repositories.NewRetailerProductsRepo(db)),
			RetailersService:        *services.NewRetailersService(repositories.NewRetailersRepo(db)),
			ProductService:          *services.NewProductService(repositories.NewProductRepository(db)),
			CartService:             *services.NewCartService(repositories.NewCartRepo(db), repositories.NewUsersRepo(db)),
			UserAddressesService:    *services.NewUserAddressesService(repositories.NewUserAddressesRepo(db), repositories.NewUsersRepo(db)),
			UploadService:           services.NewUploadService(),
		},
	}

	auth.NewAuth(app.shared_deps.logger, app.config.Env)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.newRouter(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	logger.Info().Int("Port", cfg.port).Str("Environment", cfg.Env).Msg("Started Server")
	err = srv.ListenAndServe()

	logger.Fatal().Msg(err.Error())
}
