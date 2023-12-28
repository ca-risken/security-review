package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println("Require argument")
		os.Exit(1)
	}

	if err := osCmdInjection(args[0]); err != nil {
		fmt.Printf("Failed to exec cmd: %s\n", err.Error())
		os.Exit(1)
	}
}

func osCmdInjection(userInput string) error {
	cmd := exec.Command(userInput) // OS command injection
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout
	return cmd.Run()
}
