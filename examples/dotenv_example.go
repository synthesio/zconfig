// Example demonstrating dotenv support in zconfig
// Run with: go run examples/dotenv_example.go
// Or with custom dotenv: go run examples/dotenv_example.go --dotenv=examples/custom.env

package main

import (
	"context"
	"fmt"
	"log"

	"github.com/synthesio/zconfig/v2"
)

type Config struct {
	Database struct {
		URL string `key:"url" description:"Database connection URL"`
	} `key:"database"`

	API struct {
		Key string `key:"key" description:"API secret key"`
	} `key:"api"`

	Debug bool `key:"debug" description:"Enable debug mode" default:"false"`
	Port  int  `key:"port" description:"Server port" default:"8080"`
}

func main() {
	var config Config

	err := zconfig.Configure(context.Background(), &config)
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	fmt.Printf("Configuration loaded:\n")
	fmt.Printf("  Database URL: %s\n", config.Database.URL)
	fmt.Printf("  API Key: %s\n", config.API.Key)
	fmt.Printf("  Debug: %t\n", config.Debug)
	fmt.Printf("  Port: %d\n", config.Port)
}
