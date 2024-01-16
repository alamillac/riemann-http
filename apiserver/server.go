package apiserver

import (
  "fmt"
  "log"
  "net/http"

  riemann "github.com/riemann/riemann-go-client"
  "github.com/go-redis/redis/v8"

  "github.com/go-chi/chi/v5"
  "github.com/go-chi/chi/v5/middleware"
  "github.com/go-chi/render"

  "riemannhttp/domain/metric"
  "riemannhttp/domain/asn"
)

type Server struct {
  app *chi.Mux
  cfg ApiConfig
}

func NewServer(rc *riemann.TCPClient, redisClient *redis.Client, cfg ApiConfig) *Server {
  creds := cfg.GetApiCredential()
  app := chi.NewRouter()
  app.Use(middleware.Logger)
  app.Use(middleware.BasicAuth("Realm", creds))
  app.Use(render.SetContentType(render.ContentTypeJSON))

  asnSvc := asn.NewService(redisClient)
  asnHttp := asn.NewHTTP(asnSvc)
  app.Get("/asn", asnHttp.Get)

  metricSvc := metric.NewService(rc, asnSvc)
  metricHttp := metric.NewHTTP(metricSvc)
  app.Post("/metric", metricHttp.Create)

  log.Print("Server ready")

  return &Server{
    app: app,
    cfg: cfg,
  }
}

func (s Server) Run() error {
  httpPort := s.cfg.GetApiPort()
  return http.ListenAndServe(fmt.Sprintf(":%d", httpPort), s.app)
}
