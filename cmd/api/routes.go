package main

import (
	"Obsonarium-backend/internal/handlers/healthcheck"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func (app *application) newRouter() *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/healthcheck",healthcheck.NewHealthCheckHandler(app.config.Env,app.shared_deps.JSONutils.Writer))

	return r;
}
