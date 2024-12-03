package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

var BUILTINS = map[string]struct{}{
	"exit": {},
	"echo": {},
	"type": {},
}
var PATH []string = make([]string, 0)

func main() {
	initializePath()
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Fprint(os.Stdout, "$ ")
		input, err := readInput(scanner)
		if err != nil {
			fmt.Println(err)
			continue
		}
		handleInput(input)
	}
}

func handleInput(input string) {
	cmds := strings.Fields(input)

	switch cmds[0] {
	case "exit":
		os.Exit(0)
	case "echo":
		fmt.Println(input[5:])
	case "type":
		if len(cmds) < 2 {
			fmt.Println("Empty type command")
			return
		}
		if _, ok := BUILTINS[cmds[1]]; ok {
			fmt.Printf("%s is a shell builtin\n", cmds[1])
			return
		}
		cmdPath, err := exec.LookPath(cmds[1])
		if err != nil {
			fmt.Printf("%s: %s\n", cmds[1], "not found")
			return
		}
		fmt.Printf("%s is %s\n", cmds[1], cmdPath)
	default:
		cmd := exec.Command(cmds[0], cmds[1:]...)
		output, err := cmd.Output()
		if err != nil {
			if strings.Contains(err.Error(), "executable file not found in $PATH") {
				fmt.Printf("%s: not found\n", input)
			} else {
				fmt.Println("error executing command: ", err)
			}
			return
		}

		fmt.Println(strings.TrimSpace(string(output)))
	}
}

func initializePath() {
	pathStr := os.Getenv("PATH")
	if pathStr == "" {
		return
	}

	PATH = strings.Split(pathStr, ":")
}

func readInput(scanner *bufio.Scanner) (string, error) {
	scanner.Scan()
	input := scanner.Text()

	if len(input) == 0 {
		return "", errors.New("Empty input")
	}
	return input, nil
}
