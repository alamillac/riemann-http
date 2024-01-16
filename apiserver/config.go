package apiserver

type ApiConfig interface {
  GetApiCredential() map[string]string
  GetApiPort() int
}
