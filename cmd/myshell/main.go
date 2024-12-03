package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

var BUILTINS = map[string]struct{}{
	"exit": {},
	"echo": {},
	"type": {},
}

func main() {
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
		}
		if _, ok := BUILTINS[cmds[1]]; ok {
			fmt.Printf("%s is a shell builtin\n", cmds[1])
		} else {
			fmt.Printf("%s: not found\n", cmds[1])
		}
	default:
		fmt.Printf("%s: not found\n", input)
	}
}

func readInput(scanner *bufio.Scanner) (string, error) {
	scanner.Scan()
	input := scanner.Text()

	if len(input) == 0 {
		return "", errors.New("Empty input")
	}
	return input, nil
}
