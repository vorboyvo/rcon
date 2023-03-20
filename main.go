package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	conn := NewRCONConnection("example.com", 27015, "examplePassword")
	defer conn.close()

	// Scanner for input
	scanner := bufio.NewScanner(os.Stdin)
	for scan := true; scan; {
		fmt.Print("> ")
		scan = scanner.Scan()
		result := conn.sendCommand(scanner.Text())
		fmt.Print(result)
	}
	if err := scanner.Err(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(5)
	}
}
