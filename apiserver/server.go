package apiserver

import (
	"fmt"
	"log"
	"net/http"

	riemann "github.com/riemann/riemann-go-client"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"

	"riemannhttp/domain/metric"
)

type Server struct {
	app *chi.Mux
	cfg ApiConfig
}

func NewServer(rc *riemann.TCPClient, cfg ApiConfig) *Server {
	creds := cfg.GetApiCredential()
	app := chi.NewRouter()
	app.Use(middleware.Logger)
	app.Use(middleware.BasicAuth("Realm", creds))
	app.Use(render.SetContentType(render.ContentTypeJSON))

	metricSvc := metric.NewService(rc)
	h := metric.NewHTTP(metricSvc)
	app.Post("/metric", h.Create)
	log.Print("Listen to metrics")

	return &Server{
		app: app,
		cfg: cfg,
	}
}

func (s Server) Run() error {
	httpPort := s.cfg.GetApiPort()
	return http.ListenAndServe(fmt.Sprintf(":%d", httpPort), s.app)
}
