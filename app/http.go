package app

import (
	"log"
	"net/http"

	"github.com/go-chi/render"
)

func MetricHandler(w http.ResponseWriter, r *http.Request) {
	log.Print("metric received")
	metric := &MetricPayload{}
	if err := render.Bind(r, metric); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	if err := metric.Send(); err != nil {
		render.Render(w, r, ErrOperationError(err))
		return
	}

	render.Status(r, http.StatusCreated)
	render.Render(w, r, metric)
}
