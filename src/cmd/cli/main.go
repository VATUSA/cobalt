package main

import (
	"fmt"
	"os"
	"vatusa-cobalt/cli"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: cli <command>")
		fmt.Println("Commands:")
		for command, _ := range cli.Commands {
			fmt.Printf("\t%s\n", command)
		}
		return
	}

	args := os.Args[1:]
	command := args[0]
	var commandArgs []string
	if len(args) > 1 {
		commandArgs = args[1:]
	}

	commandFunc, ok := cli.Commands[command]
	if !ok {
		fmt.Printf("Unknown command: %s\n", command)
		return
	}

	err := commandFunc(commandArgs)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}
	return
}
