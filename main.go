package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
)

func main() {
	// Interpret flags
	flagHost := flag.String("H", "", "Hostname or IP")
	flagPort := flag.Int("p", 27015, "Port")
	flagPassword := flag.String("P", "", "RCON Password")
	flag.Parse()
	args := flag.Args()

	// Check for flag validity: TODO

	conn := NewRCONConnection(*flagHost, *flagPort, *flagPassword)
	defer conn.Close()

	// If command passed, just run it and be done
	if len(args) != 0 {
		cmd := strings.Join(args, " ")
		result := conn.SendCommand(cmd)
		fmt.Print(result)
		return
	}

	// Scanner for input
	scanner := bufio.NewScanner(os.Stdin)
	for scan := true; scan; {
		fmt.Print("> ")
		scan = scanner.Scan()
		result := conn.SendCommand(scanner.Text())
		fmt.Print(result)
	}
	if err := scanner.Err(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(5)
	}
}
