package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"

	"riemannhttp/app"
)

const httpPort = 8080

var user = os.Getenv("AUTH_USER")
var password = os.Getenv("AUTH_PASSWORD")

func main() {
	creds := map[string]string{user: password}
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.BasicAuth("Realm", creds))
	r.Use(render.SetContentType(render.ContentTypeJSON))

	log.Print("Listen to metrics")
	r.Post("/metric", app.MetricHandler)

	if err := http.ListenAndServe(fmt.Sprintf(":%d", httpPort), r); err != http.ErrServerClosed && err != nil {
		log.Fatalf("Error starting http server <%s>", err)
	}
}
