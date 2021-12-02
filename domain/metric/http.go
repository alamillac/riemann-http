package metric

import (
	"log"
	"net/http"

	"github.com/go-chi/render"
)

type HttpTransport interface {
	Create(w http.ResponseWriter, r *http.Request)
}

type httpTransport struct {
	svc Service
}

func NewHTTP(svc Service) HttpTransport {
	return &httpTransport{
		svc: svc,
	}
}

func (h httpTransport) Create(w http.ResponseWriter, r *http.Request) {
	log.Print("metric received")
	metric := &MetricPayload{}
	if err := render.Bind(r, metric); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	if err := h.svc.Send(metric); err != nil {
		render.Render(w, r, ErrOperationError(err))
		log.Printf("Error sending metric: %s", err)
		return
	}

	render.Status(r, http.StatusCreated)
	render.Render(w, r, metric)
}
