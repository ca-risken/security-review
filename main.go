package main

import (
	"os"

	"github.com/ca-risken/security-review/pkg/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
