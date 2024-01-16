package config

import (
  "os"
  "time"
)

type ApiConfig struct {
  User     string
  Password string
  Port     int
}

type RiemannConfig struct {
  Address        string
  ConnectTimeout time.Duration
}

type RedisConfig struct {
  Address     string
  Password    string
  DB          int
}

type Config struct {
  apiConfig     ApiConfig
  riemannConfig RiemannConfig
  redisConfig   RedisConfig
}

func (c *Config) GetApiCredential() map[string]string {
  return map[string]string{c.apiConfig.User: c.apiConfig.Password}
}

func (c *Config) GetApiPort() int {
  return c.apiConfig.Port
}

func (c *Config) GetRiemannAddress() string {
  return c.riemannConfig.Address
}

func (c *Config) GetRiemannConnectTimeout() time.Duration {
  return c.riemannConfig.ConnectTimeout
}

func (c *Config) GetRedisAddress() string {
  return c.redisConfig.Address
}

func (c *Config) GetRedisPassword() string {
  return c.redisConfig.Password
}

func (c *Config) GetRedisDB() int {
  return c.redisConfig.DB
}

func GetConfig() *Config {
  return &Config{
    apiConfig: ApiConfig{
      User:     os.Getenv("AUTH_USER"),
      Password: os.Getenv("AUTH_PASSWORD"),
      Port:     8080,
    },
    riemannConfig: RiemannConfig{
      Address:        "127.0.0.1:5555",
      ConnectTimeout: 10 * time.Second,
    },
    redisConfig: RedisConfig{
      Address:  "127.0.0.1:6379",
      Password: "",
      DB:       0,
    },
  }
}
