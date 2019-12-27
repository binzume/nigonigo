package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/binzume/nigonigo"
)

func cmdAuth() {
	id := flag.String("i", "", "mail address")
	password := flag.String("p", "", "password")
	accountFile := flag.String("a", "", "load account.json")
	sessionFile := flag.String("s", "session.json", "save session.json")

	// flag.Parse()
	flag.CommandLine.Parse(os.Args[2:])

	client := nigonigo.NewClient()

	if *accountFile != "" {
		if _, err := os.Stat(*accountFile); err != nil {
			log.Fatal("account file not exists")
		}
		err := client.LoginWithJsonFile(*accountFile)
		if err != nil {
			log.Fatal("login failed")
		}
	} else {
		pass := *password
		if pass == "" {
			fmt.Print(" Password: ")
			pass, _ = bufio.NewReader(os.Stdin).ReadString('\n')
		}
		err := client.Login(*id, strings.TrimSpace(pass))
		if err != nil {
			log.Fatal("login failed")
		}
	}

	if *sessionFile != "" {
		client.SaveLoginSession(*sessionFile)
		log.Printf("Saved: %v", *sessionFile)
	}
}
