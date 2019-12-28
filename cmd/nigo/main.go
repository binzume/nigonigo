package main

import (
	"fmt"
	"os"
)

func printMainUsage() {
	fmt.Fprintf(os.Stderr, "Usage: %v search|auth|download [params] [-help]", os.Args[0])
	os.Exit(1)
}

func main() {
	if len(os.Args) < 2 {
		printMainUsage()
	}

	switch os.Args[1] {
	case "auth":
		cmdAuth()
	case "search":
		cmdSearch()
	case "download":
		cmdDownload()
	default:
		printMainUsage()
	}
}
