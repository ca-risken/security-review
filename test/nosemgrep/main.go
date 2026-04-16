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
	// cmd1: All ignored semgrep rules. (want: ignored)
	// nosemgrep
	cmd1 := exec.Command(userInput) // OS command injection
	cmd1.Stdout = os.Stdout
	cmd1.Stderr = os.Stdout
	if err := cmd1.Run(); err != nil {
		return err
	}

	// cmd2: specific ignored semgrep rule. (want: ignored)
	// nosemgrep: go.lang.security.audit.dangerous-exec-command.dangerous-exec-command
	cmd2 := exec.Command(userInput) // OS command injection
	cmd2.Stdout = os.Stdout
	cmd2.Stderr = os.Stdout
	if err := cmd1.Run(); err != nil {
		return err
	}

	// cmd3: other rule ignored. (want: detected)
	// nosemgrep: go.lang.security.injection.tainted-sql-string.tainted-sql-string
	cmd3 := exec.Command(userInput) // OS command injection
	cmd3.Stdout = os.Stdout
	cmd3.Stderr = os.Stdout
	if err := cmd1.Run(); err != nil {
		return err
	}
	return nil
}
