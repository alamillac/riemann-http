package main

import (
	"log"
	"os"
	"riemannhttp/apiserver"
	"riemannhttp/domain/cerberus"
	config "riemannhttp/internal"

	"context"
	"github.com/go-redis/redis/v8"
	riemann "github.com/riemann/riemann-go-client"
)

func createCerberus(rc *riemann.TCPClient, cfg *config.Config) *cerberus.Cerberus {
	cubaAsn := "27725"
	jenkins := cerberus.Jenkins{
		BaseUrl:  cfg.GetJenkinsBaseUrl(),
		Username: cfg.GetJenkinsUsername(),
		Token:    cfg.GetJenkinsToken(),
		Password: cfg.GetJenkinsPassword(),
	}
	guardian := cerberus.NewCerberus(&cerberus.Options{
		Rules: []cerberus.RuleOpts{
			{
				Name:   "ip",
				Type:   cerberus.IpRule,
				Window: cerberus.WindowOpts{Size: 600, Tick: 5}, // 10 minutes window (600 seconds) with tick every 5 seconds
				Trigger: cerberus.LoginTriggerOpts{ // Min 15 requests, 90% of login requests, 90% of login errors
					MinRequests:       15,
					MinRateLogin:      0.9,
					MinRateLoginError: 0.9,
					Action: &cerberus.BlockIp{
						Client:  rc,
						Jenkins: &jenkins,
					},
				},
			},
			{
				Name:   "asn-low",
				Type:   cerberus.AsnRule,
				Window: cerberus.WindowOpts{Size: 300, Tick: 5}, // 5 minutes window (300 seconds) with tick every 5 seconds
				Trigger: cerberus.RateTriggerOpts{ // Min 30 requests, 80% of errors
					MinRequests:  30,
					MinRateError: 0.8,
					Action: &cerberus.BlockAsn{
						Client:  rc,
						Jenkins: &jenkins,
					},
				},
				Ignored: []string{cubaAsn},
			},
			{
				Name:   "asn-high",
				Type:   cerberus.AsnRule,
				Window: cerberus.WindowOpts{Size: 30, Tick: 1}, // 30 seconds window with tick every 1 second
				Trigger: cerberus.RateTriggerOpts{ // Min 30 requests, 80% of errors
					MinRequests:  30,
					MinRateError: 0.8,
					Action: &cerberus.BlockAsn{
						Client:  rc,
						Jenkins: &jenkins,
					},
				},
				Ignored: []string{cubaAsn},
			},
		},
	})
	return guardian
}

func main() {
	cfg := config.GetConfig()
	rc := riemann.NewTCPClient(cfg.GetRiemannAddress(), cfg.GetRiemannConnectTimeout())
	if err := rc.Connect(); err != nil {
		log.Printf("Failed to connect to riemann server. %s\n", err)
		os.Exit(1)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.GetRedisAddress(),
		Password: cfg.GetRedisPassword(),
		DB:       cfg.GetRedisDB(),
	})

	var ctx = context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Printf("Failed to connect to redis server. %s\n", err)
		os.Exit(1)
	}

	guardian := createCerberus(rc, cfg)
	server := apiserver.NewServer(rc, guardian, redisClient, cfg)
	if err := server.Run(); err != nil {
		log.Fatalf("Error starting http server <%s>", err)
		os.Exit(1)
	}
}
