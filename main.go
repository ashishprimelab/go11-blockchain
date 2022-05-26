package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/tensor-programming/golang-blockchain/cli"
)

func main() {
	defer os.Exit(0)
	err := godotenv.Load()

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// api := api.Api{}
	// api.Run()
	cmd := cli.CommandLine{}
	cmd.Run()
}
