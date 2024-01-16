package asn

import (
  "log"
  "net/http"
  "errors"

  "github.com/go-chi/render"
)

type HttpTransport interface {
  Get(w http.ResponseWriter, r *http.Request)
}

type httpTransport struct {
  svc Service
}

func NewHTTP(svc Service) HttpTransport {
  return &httpTransport{
    svc: svc,
  }
}

func (h httpTransport) Get(w http.ResponseWriter, r *http.Request) {
  log.Print("ASN request received")
  ip := r.URL.Query().Get("ip")
  if ip == "" {
    err := errors.New("IP is required")
    render.Render(w, r, ErrInvalidRequest(err))
    return
  }

  asn, err := h.svc.GetASNForIP(ip)
  if err != nil {
    render.Render(w, r, ErrOperationError(err))
    log.Printf("Error getting ASN from ip %s: %s", ip, err)
    return
  }

  log.Printf("ASN: %s IP: %s", asn, ip)
  render.Render(w, r, &ASNResponse{
    ASN: asn,
    IP: ip,
  })
}
