package server

import (
	"time"
)

//go:generate go run github.com/g4s8/envdoc@latest -output ../environments.md -type Config
type Config struct {
	// Port to listen for incoming connections
	Port int `env:"PORT" envDefault:"1323"`
	// Timeout for outgoing http requests
	Timeout time.Duration `env:"TIMEOUT" envDefault:"60s"`
	// AllowPrivateIP to connect private ip for test
	AllowPrivateIP bool `env:"ALLOW_PRIVATE_IP" envDefault:"false"`
}
