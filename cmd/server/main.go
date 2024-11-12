package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/mwantia/valkey-snapshot/pkg/config"
	"github.com/mwantia/valkey-snapshot/pkg/redis"
)

func init() {
	flag.Set("logtostderr", "true")
}

var (
	Config = flag.String("config", "", "Defines the configuration path used by this application")
)

func main() {
	flag.Parse()

	if strings.TrimSpace(*Config) == "" {
		log.Fatal(fmt.Errorf("configuration path has not been defined"))
	}

	cfg, err := config.LoadConfig(*Config)
	if err != nil {
		log.Fatal(err)
	}

	if err := redis.Start(cfg); err != nil {
		log.Fatal(err)
	}
}
