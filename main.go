package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
)

func main() {
	os.Exit(mainWithCode())
}

func mainWithCode() int {
	// Interpret flags
	flagHost := flag.String("H", "", "Hostname or IP")
	flagPort := flag.Int("p", 27015, "Port")
	flagPassword := flag.String("P", "", "RCON Password")
	flag.Parse()
	args := flag.Args()

	// Check for flag validity
	// Host

	conn, err := NewRCONConnection(*flagHost, *flagPort, *flagPassword)
	if err != nil {
		if connFailure, ok := err.(ConnectionFailure); ok {
			_, err := fmt.Fprintln(os.Stderr, connFailure.Error())
			if err != nil {
				panic(err)
			}
			return 2
		} else if authFailure, ok := err.(AuthenticationFailure); ok {
			_, err := fmt.Fprintln(os.Stderr, authFailure.Error())
			if err != nil {
				panic(err)
			}
			return 3
		} else {
			panic(err)
		}
	}
	defer conn.Close()

	// If command passed, just run it and be done
	if len(args) != 0 {
		cmd := strings.Join(args, " ")
		result, err := conn.SendCommand(cmd)
		if err != nil {
			panic(err)
		}
		fmt.Print(result)
		return 0
	}

	// Scanner for input
	scanner := bufio.NewScanner(os.Stdin)
	for scan := true; scan; {
		fmt.Print("> ")
		scan = scanner.Scan()
		result, err := conn.SendCommand(scanner.Text())
		if err != nil {
			panic(err)
		}
		fmt.Print(result)
	}
	if err := scanner.Err(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		return 5
	}
	return 0
}
