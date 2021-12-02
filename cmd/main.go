package main

import (
	"log"
	"os"
	"riemannhttp/apiserver"
	config "riemannhttp/internal"

	riemann "github.com/riemann/riemann-go-client"
)

func main() {
	cfg := config.GetConfig()
	rc := riemann.NewTCPClient(cfg.GetRiemannAddress(), cfg.GetRiemannConnectTimeout())
	if err := rc.Connect(); err != nil {
		log.Printf("Failed to connect to riemann server. %s\n", err)
		os.Exit(1)
	}

	server := apiserver.NewServer(rc, cfg)
	if err := server.Run(); err != nil {
		log.Fatalf("Error starting http server <%s>", err)
		os.Exit(1)
	}
}
