package main

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Port           int  `env:"PORT" envDefault:"1323"`
	AllowPrivateIP bool `env:"ALLOW_PRIVATE_IP" envDefault:"false"`
}

var config Config

func loadConfig() {
	if err := env.Parse(&config); err != nil {
		fmt.Printf("%+v\n", err)
	}
}
