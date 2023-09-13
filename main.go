// Super simple file list builder.
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/ziutek/telnet"
)

var (
	// Version is exported during build
	Version   string
	isVerbose bool
	buffer    string
)

func main() {
	var err error

	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Println("telleq version:", Version)
		os.Exit(0)
	}

	username := ""
	password := ""
	host := "localhost"
	port := "9000"
	flag.StringVar(&host, "host", "localhost", "Host to connect to")
	flag.StringVar(&port, "port", "9000", "Port to connect to")
	flag.StringVar(&username, "username", "", "Username to use")
	flag.StringVar(&password, "password", "", "Password to use")
	isVerbose = *flag.Bool("v", false, "Verbose output")
	flag.Parse()

	if len(flag.Args()) == 0 {
		usage()
		os.Exit(1)
	}

	conn, err := telnet.Dial("tcp", host+":"+port)
	if err != nil {
		println("Error connecting to telnet:", err)
		failDump()
		os.Exit(1)
	}
	defer conn.Close()

	err = conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	if err != nil {
		println("Error setting read deadline:", err)
		failDump()
		os.Exit(1)
	}
	err = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	if err != nil {
		println("Error setting write deadline:", err)
		failDump()
		os.Exit(1)
	}

	index := 0
	skipAuth := false

	index, err = conn.SkipUntilIndex("Username:", "Connection established from localhost, assuming admin")
	if err != nil {
		println("Error waiting for username prompt:", err)
		failDump()
		os.Exit(1)
	}
	if index != 0 {
		skipAuth = true
	}

	if !skipAuth {
		err = sendLn(conn, username)
		if err != nil {
			println("Error sending username:", err)
			failDump()
			os.Exit(1)
		}

		err = conn.SkipUntil("Password:")
		if err != nil {
			println("Error waiting for password prompt:", err)
			failDump()
			os.Exit(1)
		}

		err = sendLn(conn, password)
		if err != nil {
			println("Error sending password:", err)
			failDump()
			os.Exit(1)
		}
	}

	err = sendLn(conn, "echo off")
	if err != nil {
		println("Error sending echo off:", err)
		failDump()
		os.Exit(1)
	}

	err = sendLn(conn, "acceptmessages on")
	if err != nil {
		println("Error sending acceptmessages on:", err)
		failDump()
		os.Exit(1)
	}

	for _, command := range flag.Args() {
		err = sendLn(conn, command)
		if err != nil {
			println("Error sending command:", err)
			failDump()
			os.Exit(1)
		}
	}
}

func usage() {
	fmt.Println(`Usage: telleq [-host "localhost"] [-port "9000"] [-username "username"] [-password "password"] [commands...]`)
	fmt.Println("This program runs build if the git version on folders... matches the latest commit, otherwise it fetches the download url. (You can set -url none to do nothing if matches)")
}

func println(a ...interface{}) {
	if !isVerbose {
		buffer += fmt.Sprintln(a...)
		return
	}
	fmt.Printf("[telleq] ")
	fmt.Println(a...)
}

func failDump() {
	if buffer == "" {
		return
	}
	fmt.Println(buffer)
	buffer = ""
}

func sendLn(conn *telnet.Conn, s string) (err error) {
	if conn == nil {
		return fmt.Errorf("no connection created")
	}
	buf := make([]byte, len(s)+1)
	copy(buf, s)
	buf[len(s)] = '\n'

	_, err = conn.Write(buf)
	if err != nil {
		return fmt.Errorf("sendLn: %s: %w", s, err)
	}
	return
}
