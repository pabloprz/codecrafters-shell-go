package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	fmt.Fprint(os.Stdout, "$ ")

	scanner := bufio.NewScanner(os.Stdin)

	scanner.Scan()
	input := scanner.Text()
	if len(input) != 0 {
		input = strings.TrimSuffix(input, "\n")
		fmt.Printf("%s: not found\n", input)
	}
}
