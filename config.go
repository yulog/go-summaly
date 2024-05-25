package main

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

//go:generate go run github.com/g4s8/envdoc@latest -output environments.md -type Config
type Config struct {
	// Port to listen for incoming connections
	Port int `env:"PORT" envDefault:"1323"`
	// AllowPrivateIP to connect private ip for test
	AllowPrivateIP bool `env:"ALLOW_PRIVATE_IP" envDefault:"false"`
}

var config Config

func loadConfig() {
	if err := env.Parse(&config); err != nil {
		fmt.Printf("%+v\n", err)
	}
}
