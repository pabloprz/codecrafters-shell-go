package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"unicode"
)

type output struct {
	stdOut    string
	stdErr    string
	appendOut bool
	appendErr bool
}

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

var REDIRECTS = map[string]struct{}{
	">":   {},
	"1>":  {},
	"2>":  {},
	">>":  {},
	"1>>": {},
	"2>>": {},
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
	cmds, redir := parseInput(input)

	switch cmds[0] {
	case "exit":
		os.Exit(0)
	case "echo":
		var sb strings.Builder
		for i, cmd := range cmds[1:] {
			if i == len(cmds)-2 {
				sb.WriteString(cmd)
			} else {
				sb.WriteString(fmt.Sprintf("%s ", cmd))
			}
		}
		printResult(sb.String(), nil, redir)
	case "pwd":
		res, err := executePwd()
		printResult(res, err, redir)
	case "type":
		res, err := executeType(cmds[1:])
		printResult(res, err, redir)
	case "cd":
		res, err := executeCd(cmds[1:])
		printResult(res, err, redir)
	default:
		cmd := exec.Command(cmds[0], cmds[1:]...)
		var out bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &stderr
		err := cmd.Run()
		if err != nil {
			if strings.Contains(err.Error(), "executable file not found in $PATH") {
				printResult("", errors.New(fmt.Sprintf("%s: not found", input)), redir)
			} else {
				printResult("", errors.New(strings.TrimSpace(stderr.String())), redir)
			}
		}

		output := out.Bytes()
		if len(output) == 0 {
			return
		}

		printResult(strings.TrimSpace(string(output)), nil, redir)
	}
}

func printResult(res string, error error, output output) {
	file, err := openFile(output)
	if err != nil {
		fmt.Println(err)
		return
	}
	if file != nil {
		defer file.Close()
	}

	if error != nil {
		if output.stdErr == "" {
			fmt.Println(error)
		} else {
			writeToFile(file, error.Error())
		}
	}
	if res != "" {
		if output.stdOut == "" {
			fmt.Println(res)
		} else {
			writeToFile(file, res)
		}
	}
}

func openFile(output output) (*os.File, error) {
	if output.stdErr == "" && output.stdOut == "" {
		return nil, nil
	}

	flags := os.O_RDWR | os.O_CREATE
	if output.appendOut || output.appendErr {
		flags = os.O_APPEND | os.O_CREATE | os.O_WRONLY
	}

	filename := output.stdErr
	if output.stdOut != "" {
		filename = output.stdOut
	}

	file, err := os.OpenFile(filename, flags, 0o777)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func writeToFile(file *os.File, content string) {
	if _, err := file.WriteString("\n" + content); err != nil {
		fmt.Println(err)
	}
}

func executeCd(args []string) (string, error) {
	if len(args) == 0 {
		return "", nil
	}

	if args[0] == "~" {
		args[0] = os.Getenv("HOME")
	}

	err := os.Chdir(args[0])
	if err != nil {
		if strings.Contains(err.Error(), "no such file or directory") {
			return "", errors.New(fmt.Sprintf("cd: %s: No such file or directory", args[0]))
		}
		return "", errors.New(fmt.Sprintf("cd: %s: %s", args[0], err.Error()))
	}
	return "", nil
}

func executePwd() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return dir, nil
}

func executeType(args []string) (string, error) {
	if len(args) == 0 {
		return "", errors.New("Empty type command")
	}
	if _, ok := BUILTINS[args[0]]; ok {
		return fmt.Sprintf("%s is a shell builtin", args[0]), nil
	}
	cmdPath, err := exec.LookPath(args[0])
	if err != nil {
		return fmt.Sprintf("%s: %s", args[0], "not found"), nil
	}
	return fmt.Sprintf("%s is %s", args[0], cmdPath), nil
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

func parseInput(input string) ([]string, output) {
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

	finalTokens := make([]string, 0)
	output := output{}
	for i := 0; i < len(tokens); {
		t := tokens[i]
		if _, ok := REDIRECTS[t]; ok {
			if i == len(tokens)-1 {
				continue
			}

			sendTo := tokens[i+1]
			switch t {
			case ">", "1>":
				output.stdOut = sendTo
				output.appendOut = false
			case "2>":
				output.stdErr = sendTo
				output.appendErr = false
			case ">>", "1>>":
				output.stdOut = sendTo
				output.appendOut = true
			case "2>>":
				output.stdErr = sendTo
				output.appendOut = true
			}
			i += 2
		} else {
			finalTokens = append(finalTokens, t)
			i++
		}

	}
	return finalTokens, output
}
