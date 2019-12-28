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

func authLogin(client *nigonigo.Client, sessionFile, accountFile, id, password string) error {
	if sessionFile != "" {
		err := client.LoadLoginSession(sessionFile)
		if err == nil {
			return nil
		}
	}
	if accountFile != "" {
		if _, err := os.Stat(accountFile); err != nil {
			return err
		}
		return client.LoginWithJsonFile(accountFile)
	} else if id != "" || sessionFile == "" {
		if id == "" {
			fmt.Print(" Account: ")
			id, _ = bufio.NewReader(os.Stdin).ReadString('\n')
		}
		if password == "" {
			fmt.Print(" Password: ")
			password, _ = bufio.NewReader(os.Stdin).ReadString('\n')
		}
		return client.Login(strings.TrimSpace(id), strings.TrimSpace(password))
	}
	return fmt.Errorf("no account, login to 'nigo auth'")
}

func authLogout(client *nigonigo.Client, sessionFile string) {
	err := client.LoadLoginSession(sessionFile)
	if err == nil {
		client.Logout()
		log.Println("Logged off")
	}
	os.Remove(sessionFile)
}

func cmdAuth() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s auth [options]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "       %s auth logout\n", os.Args[0])
		flag.PrintDefaults()
	}
	id := flag.String("i", "", "mail address")
	password := flag.String("p", "", "password")
	accountFile := flag.String("a", "", "load account.json")
	sessionFile := flag.String("s", defaultSessionFilePath, "save session.json")
	// flag.Parse()
	flag.CommandLine.Parse(os.Args[2:])

	if flag.Arg(0) == "logout" {
		authLogout(nigonigo.NewClient(), *sessionFile)
		return
	}

	if flag.NArg() != 0 {
		flag.Usage()
		return
	}

	client := nigonigo.NewClient()
	err := authLogin(client, "", *accountFile, *id, *password)
	if err != nil {
		log.Fatalf("Failed to login: %v", err)
	}

	if *sessionFile != "" {
		client.SaveLoginSession(*sessionFile)
		log.Printf("Saved: %v", *sessionFile)
	}
}
