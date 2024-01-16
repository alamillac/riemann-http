package metric

import (
  "net/http"

  "gopkg.in/go-playground/validator.v9"
)

type MetricState string

const (
  MetricOK       = MetricState("ok")
  MetricWarning  = MetricState("warning")
  MetricError    = MetricState("error")
  MetricCritical = MetricState("critical")
)

type Metric struct {
  Service     string            `json:"service" validate:"required"`
  Description string            `json:"description" validate:"required"`
  Metric      *int64            `json:"metric" validate:"required"`
  State       MetricState       `json:"state" validate:"required"`
  Host        string            `json:"host" validate:"required"`
  Tags        []string          `json:"tags,omitempty"`
  TTL         int64             `json:"ttl,omitempty"`
  Attributes  map[string]string `json:"attributes,omitempty"`
}

type MetricPayload struct {
  *Metric
}

func (mp *MetricPayload) Render(w http.ResponseWriter, r *http.Request) error {
  return nil
}

func (mp *MetricPayload) Bind(r *http.Request) error {
  v := validator.New()
  return v.Struct(mp)
}
