package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

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
	if strings.HasPrefix(input, "exit") {
		os.Exit(0)
	}
	fmt.Printf("%s: not found\n", input)
}

func readInput(scanner *bufio.Scanner) (string, error) {
	scanner.Scan()
	input := scanner.Text()

	if len(input) == 0 {
		return "", errors.New("Empty input")
	}
	return input, nil
}
