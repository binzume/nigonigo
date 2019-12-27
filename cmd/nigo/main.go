package main

import (
	"fmt"
	"os"
)

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage: %v search|auth|download [params]", os.Args[0])
	os.Exit(1)
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
	}
	cmd := os.Args[1]

	if cmd == "search" {
		cmdSearch()
	} else if cmd == "download" {
		cmdDownload()
	} else if cmd == "auth" {
		cmdAuth()
	} else {
		printUsage()
	}
}
