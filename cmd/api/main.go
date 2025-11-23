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

	"github.com/joho/godotenv"
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
	logger                         zerolog.Logger
	JSONutils                      jsonutils.JSONutils
	AuthService                    services.AuthService
	RetailerProductsService        services.RetailerProductsService
	RetailersService               services.RetailersService
	WholesalersService             services.WholesalersService
	WholesalerProductService       services.WholesalerProductService
	WholesalerProductsService      services.WholesalerProductsService
	ProductService                 services.ProductService
	CartService                    services.CartService
	RetailerCartService            services.RetailerCartService
	UserAddressesService           services.UserAddressesService
	ProductReviewsService          services.ProductReviewsService
	ProductQueriesService          services.ProductQueriesService
	UploadService                  *services.UploadService
	CheckoutService                *services.CheckoutService
	ConsumerOrdersService          *services.ConsumerOrdersService
	RetailerCheckoutService        *services.RetailerCheckoutService
	RetailerWholesaleOrdersService *services.RetailerWholesaleOrdersService
	StripeService                  *services.StripeService
	ConsumerOTPService            *services.ConsumerOTPService
	ShopsService                   *services.ShopsService
	UsersRepo                      repositories.IUsersRepo
}

type application struct {
	config      config
	shared_deps dependencies
}

func main() {
	err := godotenv.Load()
	if err != nil {
		// It's okay if .env doesn't exist in production if env vars are set otherwise
		fmt.Println("Error loading .env file")
	}

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
			logger:                         logger,
			JSONutils:                      jsonutils.NewJSONutils(),
			AuthService:                    *services.NewAuthService(repositories.NewUsersRepo(db), repositories.NewRetailersRepo(db), repositories.NewWholesalersRepo(db)),
			RetailerProductsService:        *services.NewRetailerProductsService(repositories.NewRetailerProductsRepo(db)),
			RetailersService:               *services.NewRetailersService(repositories.NewRetailersRepo(db)),
			WholesalersService:             *services.NewWholesalersService(repositories.NewWholesalersRepo(db)),
			WholesalerProductService:       *services.NewWholesalerProductService(repositories.NewWholesalerProductRepository(db)),
			WholesalerProductsService:      *services.NewWholesalerProductsService(repositories.NewWholesalerProductRepository(db)),
			ProductService:                 *services.NewProductService(repositories.NewProductRepository(db)),
			CartService:                    *services.NewCartService(repositories.NewCartRepo(db), repositories.NewUsersRepo(db), repositories.NewRetailerProductsRepo(db)),
			RetailerCartService:            *services.NewRetailerCartService(repositories.NewRetailerCartRepo(db), repositories.NewRetailersRepo(db), repositories.NewWholesalerProductRepository(db)),
			UserAddressesService:           *services.NewUserAddressesService(repositories.NewUserAddressesRepo(db), repositories.NewUsersRepo(db)),
			ProductReviewsService:          *services.NewProductReviewsService(repositories.NewProductReviewsRepo(db)),
			ProductQueriesService:          *services.NewProductQueriesService(repositories.NewProductQueriesRepo(db), repositories.NewUsersRepo(db), services.NewEmailService(os.Getenv("MAILTRAP_API_TOKEN"))),
			UploadService:                  services.NewUploadService(),
			StripeService:                  services.NewStripeService(),
			CheckoutService:                services.NewCheckoutService(repositories.NewConsumerOrdersRepository(db), repositories.NewRetailerProductsRepo(db), repositories.NewCartRepo(db), services.NewStripeService(), repositories.NewUsersRepo(db)),
			ConsumerOrdersService:          services.NewConsumerOrdersService(repositories.NewConsumerOrdersRepository(db), repositories.NewUsersRepo(db), services.NewEmailService(os.Getenv("MAILTRAP_API_TOKEN"))),
			RetailerWholesaleOrdersService: services.NewRetailerWholesaleOrdersService(repositories.NewRetailerWholesaleOrdersRepository(db), repositories.NewRetailersRepo(db), services.NewEmailService(os.Getenv("MAILTRAP_API_TOKEN"))),
			RetailerCheckoutService:        services.NewRetailerCheckoutService(repositories.NewRetailerWholesaleOrdersRepository(db), repositories.NewWholesalerProductRepository(db), repositories.NewRetailerCartRepo(db), services.NewStripeService(), repositories.NewRetailersRepo(db)),
			ConsumerOTPService:            services.NewConsumerOTPService(repositories.NewConsumerOTPRepo(db), services.NewEmailService(os.Getenv("MAILTRAP_API_TOKEN")), repositories.NewUsersRepo(db)),
			ShopsService:                  services.NewShopsService(repositories.NewRetailersRepo(db)),
			UsersRepo:                      repositories.NewUsersRepo(db),
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
