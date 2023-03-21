/*
Copyright 2023 vorboyvo.

This file is part of rcon.

rcon is free software: you can redistribute it and/or modify it under the terms of the GNU General Public
License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later
version.

rcon is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied
warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with rcon. If not, see
https://www.gnu.org/licenses.
*/

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/cheynewallace/tabby"
	flag "github.com/spf13/pflag"
	"io"
	"net"
	"os"
	"strings"
	"text/tabwriter"
)

var debug bool

func usage() {
	var usageString string
	usageString += "Usage:\n"
	usageString += " rcon [options]\n"
	usageString += " rcon [options] command\n"

	var optionsString string
	optionsString += "Options:\n"
	buf := bytes.Buffer{}
	writer := tabwriter.NewWriter(&buf, 0, 0, 1, ' ', 0)
	options := tabby.NewCustom(writer)
	flag.VisitAll(func(f *flag.Flag) {
		if f.Hidden {
			return
		}
		options.AddLine("-"+f.Shorthand+",", "--"+f.Name, f.Usage)
	})
	options.Print()
	optionsString += buf.String()
	_, _ = fmt.Fprintln(os.Stderr, usageString+"\n"+optionsString)
}

func mainWithCode() int {
	// Interpret flags
	flagHost := flag.StringP("host", "H", "", "Hostname or IP")
	flagPort := flag.IntP("port", "p", 27015, "Port")
	flagPassword := flag.StringP("password", "P", "", "RCON Password")
	flagDebug := flag.BoolP("debug", "d", false, "Additional output for debug purposes")
	flagHelp := flag.BoolP("help", "h", false, "Show this help text")
	flag.CommandLine.SortFlags = false
	flag.CommandLine.Usage = usage
	flag.Parse()
	args := flag.Args()
	debug = *flagDebug

	// Show help text if requested, then exit
	if *flagHelp {
		usage()
		return -9
	}

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
			fmt.Println()
			usage()
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

func main() {
	os.Exit(mainWithCode())
}
