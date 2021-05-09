package main

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
)

var defaultSessionFilePath = ".nigo_session.json"

func printMainUsage() {
	fmt.Fprintf(os.Stderr, "Usage: %v auth|search|mylist|download [params] [-help]\n", os.Args[0])
	os.Exit(1)
}

func main() {
	if len(os.Args) < 2 {
		printMainUsage()
	}

	if u, err := user.Current(); err == nil {
		defaultSessionFilePath = filepath.Join(u.HomeDir, ".nigo_session.json")
	}

	switch os.Args[1] {
	case "auth":
		cmdAuth()
	case "search":
		cmdSearch()
	case "mylist":
		cmdMylist()
	case "download":
		cmdDownload()
	default:
		printMainUsage()
	}
}
