package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"unicode"
)

var BUILTINS = map[string]struct{}{
	"exit": {},
	"echo": {},
	"type": {},
	"pwd":  {},
	"cd":   {},
}

var DOUBLE_SPECIAL_CHARS = map[rune]struct{}{
	'$':  {},
	'\\': {},
	'"':  {},
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
	cmds := parseInput(input)

	switch cmds[0] {
	case "exit":
		os.Exit(0)
	case "echo":
		for i, cmd := range cmds[1:] {
			if i == len(cmds)-2 {
				fmt.Print(cmd)
			} else {
				fmt.Printf("%s ", cmd)
			}
		}
		fmt.Print("\n")
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

	if args[0] == "~" {
		args[0] = os.Getenv("HOME")
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

func parseInput(input string) []string {
	tokens := make([]string, 0)

	for i := 0; i < len(input); {
		isSingle, isDouble, innerDouble := false, false, false
		for i < len(input) && unicode.IsSpace(rune(input[i])) {
			i++
		}

		if i >= len(input) {
			break
		}

		if rune(input[i]) == '\'' {
			isSingle = true
			i++
		} else if rune(input[i]) == '"' {
			isDouble = true
			i++
		}

		var sb strings.Builder

		for i < len(input) {
			curr := rune(input[i])
			// will finish when a space is found unles we are inside quotes
			if !isSingle && !isDouble && unicode.IsSpace(curr) {
				break
			}
			if isSingle && curr == '\'' {
				// skip over the trailing quote and break
				i++
				break
			}
			if isDouble && curr == '"' {
				if !innerDouble {
					i++
					break
				}
				// if innerdouble, we can ignore this quote
				i++
				continue
			}
			if !isSingle && !isDouble && curr == '\\' {
				i++
				if i < len(input) {
					sb.WriteByte(input[i])
					i++
				}
				continue
			}
			if isDouble && rune(input[i]) == '\\' {
				i++
				curr = rune(input[i])
				if i < len(input) {
					if _, ok := DOUBLE_SPECIAL_CHARS[curr]; ok {
						if curr == '"' {
							innerDouble = !innerDouble
						}

						sb.WriteByte(input[i])
					} else {
						sb.WriteByte(input[i-1])
						sb.WriteByte(input[i])
					}
					i++
					continue
				}
			}
			sb.WriteByte(input[i])
			i++
		}

		tokens = append(tokens, sb.String())
	}
	return tokens
}
