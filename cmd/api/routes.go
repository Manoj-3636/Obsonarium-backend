package main

import (
	"Obsonarium-backend/internal/handlers/auth"
	"Obsonarium-backend/internal/handlers/cart"
	"Obsonarium-backend/internal/handlers/healthcheck"
	"Obsonarium-backend/internal/handlers/retailer_products"
	"Obsonarium-backend/internal/handlers/retailers"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func (app *application) newRouter() *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/api/healthcheck", healthcheck.NewHealthCheckHandler(app.config.Env, app.shared_deps.JSONutils.Writer))
	r.Get("/api/auth/{provider}/callback", auth.NewAuthCallback(app.shared_deps.logger, &app.shared_deps.AuthService))
	r.Get("/api/auth/{provider}", auth.AuthProvider)
	r.Get("/api/logout/{provider}", auth.AuthLogout)
	r.Get("/api/shop", retailer_products.GetProducts(&app.shared_deps.RetailerProductsService, app.shared_deps.JSONutils.Writer))
	r.Get("/api/shop/{id}", retailer_products.GetProduct(&app.shared_deps.RetailerProductsService, app.shared_deps.JSONutils.Writer))
	r.Get("/api/retailers/{id}", retailers.GetRetailer(&app.shared_deps.RetailersService, app.shared_deps.JSONutils.Writer))

	// Cart routes with authentication middleware
	r.Route("/api/cart", func(r chi.Router) {
		r.Use(auth.RequireAuth(&app.shared_deps.AuthService, app.shared_deps.logger))
		r.Get("/", cart.GetCart(&app.shared_deps.CartService, app.shared_deps.JSONutils.Writer))
		r.Post("/", cart.AddCartItem(&app.shared_deps.CartService, app.shared_deps.JSONutils.Writer, app.shared_deps.JSONutils.Reader))
		r.Delete("/{product_id}", cart.RemoveCartItem(&app.shared_deps.CartService, app.shared_deps.JSONutils.Writer))
	})

	return r
}
