package main

import (
	"Obsonarium-backend/internal/handlers/auth"
	"Obsonarium-backend/internal/handlers/cart"
	"Obsonarium-backend/internal/handlers/healthcheck"
	"Obsonarium-backend/internal/handlers/retailer_products"
	"Obsonarium-backend/internal/handlers/retailers"
	"Obsonarium-backend/internal/handlers/upload_handler"
	"Obsonarium-backend/internal/handlers/user_addresses"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func (app *application) newRouter() *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// File server for uploaded files
	r.Handle("/uploads/*", http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))

	r.Get("/api/healthcheck", healthcheck.NewHealthCheckHandler(app.config.Env, app.shared_deps.JSONutils.Writer))
	r.Get("/api/auth/{provider}/callback", auth.NewAuthCallback(app.shared_deps.logger, &app.shared_deps.AuthService, &app.shared_deps.RetailersService))
	r.Get("/api/auth/{provider}", auth.AuthProvider)
	r.Get("/api/logout/{provider}", auth.AuthLogout)
	r.Get("/api/shop", retailer_products.GetProducts(&app.shared_deps.RetailerProductsService, app.shared_deps.JSONutils.Writer))
	r.Get("/api/shop/{id}", retailer_products.GetProduct(&app.shared_deps.RetailerProductsService, app.shared_deps.JSONutils.Writer))

	// Upload routes with retailer authentication middleware
	r.Route("/api/upload", func(r chi.Router) {
		r.Use(auth.RequireRetailer(&app.shared_deps.AuthService, app.shared_deps.logger, app.shared_deps.JSONutils.Writer))
		r.Post("/product-image", upload_handler.UploadProductImage(app.shared_deps.UploadService, app.shared_deps.JSONutils.Writer))
	})

	// Cart routes with consumer authentication middleware
	r.Route("/api/cart", func(r chi.Router) {
		r.Use(auth.RequireConsumer(&app.shared_deps.AuthService, app.shared_deps.logger, app.shared_deps.JSONutils.Writer))
		r.Get("/", cart.GetCart(&app.shared_deps.CartService, app.shared_deps.JSONutils.Writer))
		r.Get("/number", cart.GetCartNumber(&app.shared_deps.CartService, app.shared_deps.JSONutils.Writer))
		r.Post("/", cart.AddCartItem(&app.shared_deps.CartService, app.shared_deps.JSONutils.Writer, app.shared_deps.JSONutils.Reader))
		r.Delete("/{product_id}", cart.RemoveCartItem(&app.shared_deps.CartService, app.shared_deps.JSONutils.Writer))
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
		// specific routes like /me must come before /{id} to avoid being captured
		r.Route("/me", func(r chi.Router) {
			r.Use(auth.RequireRetailer(&app.shared_deps.AuthService, app.shared_deps.logger, app.shared_deps.JSONutils.Writer))
			r.Get("/", retailers.GetCurrentRetailer(&app.shared_deps.RetailersService, app.shared_deps.JSONutils.Writer))
			r.Post("/", retailers.UpdateCurrentRetailer(&app.shared_deps.RetailersService, app.shared_deps.JSONutils.Writer, app.shared_deps.JSONutils.Reader))
		})

		// Get retailer by ID
		r.Get("/{id}", retailers.GetRetailer(&app.shared_deps.RetailersService, app.shared_deps.JSONutils.Writer))
	})

	return r
}
