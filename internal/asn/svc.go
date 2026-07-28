package asn

import (
  "log"
  "context"
  "time"
  "github.com/go-redis/redis/v8"
)

var ctx = context.Background()

type Service interface {
  GetASNForIP(string) (string, error)
}

func NewService(redisClient *redis.Client) Service {
  return &svc{
    redisClient: redisClient,
  }
}

type svc struct {
  redisClient *redis.Client
}

func (s *svc) GetASNForIP(ip string) (string, error) {
    // Intentar obtener el ASN de Redis
    asn, err := s.redisClient.Get(ctx, ip).Result()
    if err == nil {
        log.Printf("Found in cache")
        return asn, nil
    }

    // Si no está en Redis, realiza la consulta DNS
    log.Printf("Not found in cache")
    asn, err = fetchASNFromDNS(ip)
    if err != nil {
        return "", err
    }

    // Almacenar el resultado en Redis con un tiempo de expiración de 7 días
    err = s.redisClient.Set(ctx, ip, asn, 7*24*time.Hour).Err()
    if err != nil {
        return "", err
    }

    return asn, nil
}
