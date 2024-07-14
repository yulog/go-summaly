package main

import (
	"github.com/yulog/go-summaly/server"
)

const name = "summaly"

const version = "0.0.1"

var revision = "HEAD"

func main() {
	server.New().Start()
}
