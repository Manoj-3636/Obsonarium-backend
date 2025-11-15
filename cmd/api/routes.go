package main

import (
	"Obsonarium-backend/internal/handlers/auth"
	"Obsonarium-backend/internal/handlers/healthcheck"
	"Obsonarium-backend/internal/handlers/retailer_products"

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
	return r
}
