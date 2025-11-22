package main

import (
	"Obsonarium-backend/internal/handlers/auth"
	"Obsonarium-backend/internal/handlers/cart"
	"Obsonarium-backend/internal/handlers/checkout"
	consumer_orders "Obsonarium-backend/internal/handlers/consumer_orders"
	"Obsonarium-backend/internal/handlers/healthcheck"
	"Obsonarium-backend/internal/handlers/product_handler"
	"Obsonarium-backend/internal/handlers/product_queries"
	"Obsonarium-backend/internal/handlers/product_reviews"
	"Obsonarium-backend/internal/handlers/retailer_cart"
	retailer_checkout "Obsonarium-backend/internal/handlers/retailer_checkout"
	"Obsonarium-backend/internal/handlers/retailer_products"
	retailer_wholesale_orders "Obsonarium-backend/internal/handlers/retailer_wholesale_orders"
	"Obsonarium-backend/internal/handlers/retailers"
	"Obsonarium-backend/internal/handlers/upload_handler"
	"Obsonarium-backend/internal/handlers/user_addresses"
	wholesaler_orders "Obsonarium-backend/internal/handlers/wholesaler_orders"
	"Obsonarium-backend/internal/handlers/wholesaler_product_handler"
	"Obsonarium-backend/internal/handlers/wholesaler_products"
	"Obsonarium-backend/internal/handlers/wholesalers"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
)

func (app *application) newRouter() *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// CORS middleware to allow credentials (cookies)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://localhost:5174", "http://localhost:5175"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// File server for uploaded files
	r.Handle("/api/uploads/*", http.StripPrefix("/api/uploads/", http.FileServer(http.Dir("./uploads"))))

	r.Get("/api/healthcheck", healthcheck.NewHealthCheckHandler(app.config.Env, app.shared_deps.JSONutils.Writer))
	r.Get("/api/auth/{provider}/callback", auth.NewAuthCallback(app.shared_deps.logger, &app.shared_deps.AuthService, &app.shared_deps.RetailersService, &app.shared_deps.WholesalersService))
	r.Get("/api/auth/{provider}", auth.AuthProvider)
	r.Get("/api/logout/{provider}", auth.AuthLogout)
	r.Get("/api/shop", retailer_products.GetProducts(&app.shared_deps.RetailerProductsService, app.shared_deps.JSONutils.Writer))
	r.Get("/api/shop/{id}", retailer_products.GetProduct(&app.shared_deps.RetailerProductsService, app.shared_deps.JSONutils.Writer))
	r.Get("/api/wholesale", wholesaler_products.GetProducts(&app.shared_deps.WholesalerProductsService, app.shared_deps.JSONutils.Writer))
	r.Get("/api/wholesale/{id}", wholesaler_products.GetProduct(&app.shared_deps.WholesalerProductsService, app.shared_deps.JSONutils.Writer))

	// Product reviews routes
	r.Route("/api/products/{product_id}/reviews", func(r chi.Router) {
		// Public GET endpoint (no auth required)
		r.Get("/", product_reviews.GetReviews(&app.shared_deps.ProductReviewsService, app.shared_deps.JSONutils.Writer))
		// Protected POST endpoint (requires consumer auth)
		r.With(auth.RequireConsumer(&app.shared_deps.AuthService, app.shared_deps.logger, app.shared_deps.JSONutils.Writer)).Post("/", product_reviews.CreateReview(&app.shared_deps.ProductReviewsService, app.shared_deps.UsersRepo, app.shared_deps.JSONutils.Writer, app.shared_deps.JSONutils.Reader))
	})

	// Product queries routes
	r.Route("/api/products/{product_id}/queries", func(r chi.Router) {
		// Protected POST endpoint (requires consumer auth)
		r.With(auth.RequireConsumer(&app.shared_deps.AuthService, app.shared_deps.logger, app.shared_deps.JSONutils.Writer)).Post("/", product_queries.PostQuery(&app.shared_deps.ProductQueriesService, app.shared_deps.UsersRepo, app.shared_deps.JSONutils.Writer, app.shared_deps.JSONutils.Reader))
	})

	// Retailer queries routes
	r.Route("/api/retailer/queries", func(r chi.Router) {
		r.Use(auth.RequireRetailer(&app.shared_deps.AuthService, app.shared_deps.logger, app.shared_deps.JSONutils.Writer))
		r.Get("/", product_queries.GetQueries(&app.shared_deps.ProductQueriesService, &app.shared_deps.RetailersService, app.shared_deps.JSONutils.Writer))
		r.Post("/{query_id}/resolve", product_queries.ResolveQuery(&app.shared_deps.ProductQueriesService, &app.shared_deps.RetailersService, app.shared_deps.JSONutils.Writer, app.shared_deps.JSONutils.Reader))
	})

	// Retailer consumer orders routes
	r.Route("/api/retailer/orders", func(r chi.Router) {
		r.Use(auth.RequireRetailer(&app.shared_deps.AuthService, app.shared_deps.logger, app.shared_deps.JSONutils.Writer))
		r.Get("/", consumer_orders.GetActiveOrders(app.shared_deps.ConsumerOrdersService, &app.shared_deps.RetailersService, app.shared_deps.JSONutils.Writer))
		r.Get("/history", consumer_orders.GetHistoryOrders(app.shared_deps.ConsumerOrdersService, &app.shared_deps.RetailersService, app.shared_deps.JSONutils.Writer))
		r.Patch("/items/{item_id}/status", consumer_orders.UpdateOrderItemStatus(app.shared_deps.ConsumerOrdersService, &app.shared_deps.RetailersService, app.shared_deps.JSONutils.Writer, app.shared_deps.JSONutils.Reader))
	})

	// Upload routes with retailer authentication middleware
	r.Route("/api/upload", func(r chi.Router) {
		r.Use(auth.RequireRetailer(&app.shared_deps.AuthService, app.shared_deps.logger, app.shared_deps.JSONutils.Writer))
		r.Post("/product-image", upload_handler.UploadProductImage(app.shared_deps.UploadService, app.shared_deps.JSONutils.Writer))
	})

	// Upload routes with wholesaler authentication middleware
	r.Route("/api/upload/wholesaler", func(r chi.Router) {
		r.Use(auth.RequireWholesaler(&app.shared_deps.AuthService, app.shared_deps.logger, app.shared_deps.JSONutils.Writer))
		r.Post("/product-image", upload_handler.UploadProductImage(app.shared_deps.UploadService, app.shared_deps.JSONutils.Writer))
	})

	// Retailer product management routes
	r.Route("/api/retailer/products", func(r chi.Router) {
		r.Use(auth.RequireRetailer(&app.shared_deps.AuthService, app.shared_deps.logger, app.shared_deps.JSONutils.Writer))
		r.Get("/", product_handler.ListRetailerProducts(&app.shared_deps.ProductService, &app.shared_deps.RetailersService, app.shared_deps.JSONutils.Writer))
		r.Post("/", product_handler.CreateProduct(&app.shared_deps.ProductService, &app.shared_deps.RetailersService, app.shared_deps.JSONutils.Writer, app.shared_deps.JSONutils.Reader))
		r.Get("/{id}", product_handler.GetRetailerProduct(&app.shared_deps.ProductService, &app.shared_deps.RetailersService, app.shared_deps.JSONutils.Writer))
		r.Put("/{id}", product_handler.UpdateProduct(&app.shared_deps.ProductService, &app.shared_deps.RetailersService, app.shared_deps.JSONutils.Writer, app.shared_deps.JSONutils.Reader))
		r.Delete("/{id}", product_handler.DeleteProduct(&app.shared_deps.ProductService, &app.shared_deps.RetailersService, app.shared_deps.JSONutils.Writer))
	})

	// Cart routes with consumer authentication middleware
	r.Route("/api/cart", func(r chi.Router) {
		r.Use(auth.RequireConsumer(&app.shared_deps.AuthService, app.shared_deps.logger, app.shared_deps.JSONutils.Writer))
		r.Get("/", cart.GetCart(&app.shared_deps.CartService, app.shared_deps.JSONutils.Writer))
		r.Get("/number", cart.GetCartNumber(&app.shared_deps.CartService, app.shared_deps.JSONutils.Writer))
		r.Get("/validate", cart.ValidateCartStock(&app.shared_deps.CartService, app.shared_deps.JSONutils.Writer))
		r.Post("/", cart.AddCartItem(&app.shared_deps.CartService, app.shared_deps.JSONutils.Writer, app.shared_deps.JSONutils.Reader))
		r.Delete("/{product_id}", cart.RemoveCartItem(&app.shared_deps.CartService, app.shared_deps.JSONutils.Writer))
	})

	// Checkout routes
	r.Route("/api/checkout", func(r chi.Router) {
		r.Use(auth.RequireConsumer(&app.shared_deps.AuthService, app.shared_deps.logger, app.shared_deps.JSONutils.Writer))
		r.Post("/", checkout.NewCheckoutHandler(app.shared_deps.CheckoutService).HandleCheckout)
	})

	// Consumer orders routes
	r.Route("/api/consumer/orders", func(r chi.Router) {
		r.Use(auth.RequireConsumer(&app.shared_deps.AuthService, app.shared_deps.logger, app.shared_deps.JSONutils.Writer))
		r.Get("/", consumer_orders.GetConsumerOrders(app.shared_deps.ConsumerOrdersService, app.shared_deps.JSONutils.Writer))
	})

	// Retailer cart routes with retailer authentication middleware
	r.Route("/api/retailer/cart", func(r chi.Router) {
		r.Use(auth.RequireRetailer(&app.shared_deps.AuthService, app.shared_deps.logger, app.shared_deps.JSONutils.Writer))
		r.Get("/", retailer_cart.GetCart(&app.shared_deps.RetailerCartService, app.shared_deps.JSONutils.Writer))
		r.Get("/number", retailer_cart.GetCartNumber(&app.shared_deps.RetailerCartService, app.shared_deps.JSONutils.Writer))
		r.Get("/validate", retailer_cart.ValidateCartStock(&app.shared_deps.RetailerCartService, app.shared_deps.JSONutils.Writer))
		r.Post("/", retailer_cart.AddCartItem(&app.shared_deps.RetailerCartService, app.shared_deps.JSONutils.Writer, app.shared_deps.JSONutils.Reader))
		r.Delete("/{product_id}", retailer_cart.RemoveCartItem(&app.shared_deps.RetailerCartService, app.shared_deps.JSONutils.Writer))
	})

	// Retailer checkout routes
	r.Route("/api/retailer/checkout", func(r chi.Router) {
		r.Use(auth.RequireRetailer(&app.shared_deps.AuthService, app.shared_deps.logger, app.shared_deps.JSONutils.Writer))
		r.Post("/", retailer_checkout.NewRetailerCheckoutHandler(app.shared_deps.RetailerCheckoutService, &app.shared_deps.RetailersService).HandleCheckout)
	})

	// Retailer wholesale orders routes (for retailers to view their orders)
	r.Route("/api/retailer/wholesale-orders", func(r chi.Router) {
		r.Use(auth.RequireRetailer(&app.shared_deps.AuthService, app.shared_deps.logger, app.shared_deps.JSONutils.Writer))
		r.Get("/", retailer_wholesale_orders.GetOrders(app.shared_deps.RetailerWholesaleOrdersService, &app.shared_deps.RetailersService, app.shared_deps.JSONutils.Writer))
	})

	// User addresses routes with consumer authentication middleware
	r.Route("/api/addresses", func(r chi.Router) {
		r.Use(auth.RequireConsumer(&app.shared_deps.AuthService, app.shared_deps.logger, app.shared_deps.JSONutils.Writer))
		r.Get("/", user_addresses.GetAddresses(&app.shared_deps.UserAddressesService, app.shared_deps.JSONutils.Writer))
		r.Post("/", user_addresses.AddAddress(&app.shared_deps.UserAddressesService, app.shared_deps.JSONutils.Writer, app.shared_deps.JSONutils.Reader))
		r.Delete("/{id}", user_addresses.RemoveAddress(&app.shared_deps.UserAddressesService, app.shared_deps.JSONutils.Writer))
	})

	// Retailer routes
	r.Route("/api/retailers", func(r chi.Router) {
		// Get current retailer profile and onboarding status (no onboarding required)
		// specific routes like /me and /products must come before /{id} to avoid being captured
		r.Route("/me", func(r chi.Router) {
			r.Use(auth.RequireRetailer(&app.shared_deps.AuthService, app.shared_deps.logger, app.shared_deps.JSONutils.Writer))
			r.Get("/", retailers.GetCurrentRetailer(&app.shared_deps.RetailersService, app.shared_deps.JSONutils.Writer))
			r.Post("/", retailers.UpdateCurrentRetailer(&app.shared_deps.RetailersService, app.shared_deps.JSONutils.Writer, app.shared_deps.JSONutils.Reader))
		})

		// Get retailer by ID
		r.Get("/{id}", retailers.GetRetailer(&app.shared_deps.RetailersService, app.shared_deps.JSONutils.Writer))
	})

	// Wholesaler routes
	r.Route("/api/wholesalers", func(r chi.Router) {
		// Get current wholesaler profile and onboarding status (no onboarding required)
		// specific routes like /me must come before /{id} to avoid being captured
		r.Route("/me", func(r chi.Router) {
			r.Use(auth.RequireWholesaler(&app.shared_deps.AuthService, app.shared_deps.logger, app.shared_deps.JSONutils.Writer))
			r.Get("/", wholesalers.GetCurrentWholesaler(&app.shared_deps.WholesalersService, app.shared_deps.JSONutils.Writer))
			r.Post("/", wholesalers.UpdateCurrentWholesaler(&app.shared_deps.WholesalersService, app.shared_deps.JSONutils.Writer, app.shared_deps.JSONutils.Reader))
		})

		// Get wholesaler by ID
		r.Get("/{id}", wholesalers.GetWholesaler(&app.shared_deps.WholesalersService, app.shared_deps.JSONutils.Writer))
	})

	// Wholesaler product management routes
	r.Route("/api/wholesaler/products", func(r chi.Router) {
		r.Use(auth.RequireWholesaler(&app.shared_deps.AuthService, app.shared_deps.logger, app.shared_deps.JSONutils.Writer))
		r.Get("/", wholesaler_product_handler.ListWholesalerProducts(&app.shared_deps.WholesalerProductService, &app.shared_deps.WholesalersService, app.shared_deps.JSONutils.Writer))
		r.Post("/", wholesaler_product_handler.CreateProduct(&app.shared_deps.WholesalerProductService, &app.shared_deps.WholesalersService, app.shared_deps.JSONutils.Writer, app.shared_deps.JSONutils.Reader))
		r.Get("/{id}", wholesaler_product_handler.GetWholesalerProduct(&app.shared_deps.WholesalerProductService, &app.shared_deps.WholesalersService, app.shared_deps.JSONutils.Writer))
		r.Put("/{id}", wholesaler_product_handler.UpdateProduct(&app.shared_deps.WholesalerProductService, &app.shared_deps.WholesalersService, app.shared_deps.JSONutils.Writer, app.shared_deps.JSONutils.Reader))
		r.Delete("/{id}", wholesaler_product_handler.DeleteProduct(&app.shared_deps.WholesalerProductService, &app.shared_deps.WholesalersService, app.shared_deps.JSONutils.Writer))
	})

	// Wholesaler orders routes (for wholesalers to manage retailer orders)
	r.Route("/api/wholesaler/orders", func(r chi.Router) {
		r.Use(auth.RequireWholesaler(&app.shared_deps.AuthService, app.shared_deps.logger, app.shared_deps.JSONutils.Writer))
		r.Get("/", wholesaler_orders.GetActiveOrders(app.shared_deps.RetailerWholesaleOrdersService, &app.shared_deps.WholesalersService, app.shared_deps.JSONutils.Writer))
		r.Get("/history", wholesaler_orders.GetHistoryOrders(app.shared_deps.RetailerWholesaleOrdersService, &app.shared_deps.WholesalersService, app.shared_deps.JSONutils.Writer))
		r.Patch("/items/{item_id}/status", wholesaler_orders.UpdateOrderItemStatus(app.shared_deps.RetailerWholesaleOrdersService, &app.shared_deps.WholesalersService, app.shared_deps.JSONutils.Writer, app.shared_deps.JSONutils.Reader))
	})

	return r
}
