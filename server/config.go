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
	// BotUA
	BotUA string `env:"BOT_UA" envDefault:"Mozilla/5.0 (compatible; SummalyBot/0.0.1; +https://github.com/yulog/go-summaly)"`
	// NonBotUA
	NonBotUA string `env:"NON_BOT_UA" envDefault:"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36"`
	// RequireNonBotUAFile
	RequireNonBotUAFile string `env:"REQUIRE_NON_BOT_UA_FILE" envDefault:"./nonbot.txt"`
	// RequireNonBotUA
	RequireNonBotUA []string `env:"REQUIRE_NON_BOT_UA,file,expand" envDefault:"${REQUIRE_NON_BOT_UA_FILE}"`
	// AllowPrivateIP to connect private ip for test
	AllowPrivateIP bool `env:"ALLOW_PRIVATE_IP" envDefault:"false"`
}
