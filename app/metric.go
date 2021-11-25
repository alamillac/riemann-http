package app

import (
	"net/http"

	"gopkg.in/go-playground/validator.v9"

	"riemannhttp/metric"
)

type MetricPayload struct {
	*metric.Metric
}

func (mp *MetricPayload) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (mp *MetricPayload) Bind(r *http.Request) error {
	v := validator.New()
	return v.Struct(mp)
}
