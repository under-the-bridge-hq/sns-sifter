package main

import (
	"os"

	"github.com/under-the-bridge-hq/sns-sifter/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:]))
}
