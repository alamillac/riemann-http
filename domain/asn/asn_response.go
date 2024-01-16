package asn

import (
  "net/http"
  "github.com/go-chi/render"
)

type ASNResponse struct {
  ASN string `json:"asn"`
  IP string  `json:"ip"`
}

func (asn *ASNResponse) Render(w http.ResponseWriter, r *http.Request) error {
  render.Status(r, http.StatusOK)
  return nil
}
