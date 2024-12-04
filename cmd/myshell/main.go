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
	"pwd":  {},
	"cd":   {},
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
	case "pwd":
		executePwd()
	case "type":
		executeType(cmds[1:])
	case "cd":
		executeCd(cmds[1:])
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

func executeCd(args []string) {
	if len(args) == 0 {
		return
	}

	err := os.Chdir(args[0])
	if err != nil {
		if strings.Contains(err.Error(), "no such file or directory") {
			fmt.Printf("cd: %s: No such file or directory\n", args[0])
			return
		}
		fmt.Printf("cd: %s: %s\n", args[0], err.Error())
	}
}

func executePwd() {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(dir)
}

func executeType(args []string) {
	if len(args) == 0 {
		fmt.Println("Empty type command")
		return
	}
	if _, ok := BUILTINS[args[0]]; ok {
		fmt.Printf("%s is a shell builtin\n", args[0])
		return
	}
	cmdPath, err := exec.LookPath(args[0])
	if err != nil {
		fmt.Printf("%s: %s\n", args[0], "not found")
		return
	}
	fmt.Printf("%s is %s\n", args[0], cmdPath)
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
