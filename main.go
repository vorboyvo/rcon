package main

import (
	"bufio"
	"fmt"
	flag "github.com/spf13/pflag"
	"io"
	"net"
	"os"
	"strings"
)

var debug bool

func main() {
	os.Exit(mainWithCode())
}

func mainWithCode() int {
	// Interpret flags
	flagHost := flag.StringP("host", "H", "", "Hostname or IP")
	flagPort := flag.IntP("port", "p", 27015, "Port")
	flagPassword := flag.StringP("password", "P", "", "RCON Password")
	flagDebug := flag.BoolP("debug", "d", false, "Additional output for debug purposes")
	flag.Parse()
	args := flag.Args()
	debug = *flagDebug

	// Check for legal arguments
	{
		var illegalArguments bool
		if *flagHost == "" {
			_, _ = fmt.Fprintln(os.Stderr, "Hostname not provided")
			illegalArguments = true
		}
		if *flagPort < 1 || *flagPort > 65535 {
			_, _ = fmt.Fprintln(os.Stderr, "Invalid port provided")
			illegalArguments = true
		}
		if *flagPassword == "" {
			_, _ = fmt.Fprintln(os.Stderr, "Password not provided")
			illegalArguments = true
		}
		if illegalArguments {
			return -1
		}
	}

	// Create connection, handle failure, defer closure
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
			if err == io.EOF {
				fmt.Print(result)
				_, _ = fmt.Fprintln(os.Stderr, "Connection closed by remote host")
				return 4
			} else if opErr, ok := err.(*net.OpError); ok {
				fmt.Print(result)
				_, _ = fmt.Fprintln(os.Stderr, opErr)
				return 4
			} else {
				panic(err)
			}
		}
		fmt.Print(result)
	}
	if err := scanner.Err(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		return 5
	}
	return 0
}
