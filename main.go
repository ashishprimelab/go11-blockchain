package main

import (
	"os"

	"github.com/tensor-programming/golang-blockchain/cli"
)

func main() {
	defer os.Exit(0)

	// api := api.Api{}
	// api.Run()
	cmd := cli.CommandLine{}
	cmd.Run()
}
