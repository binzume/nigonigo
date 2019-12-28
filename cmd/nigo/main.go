package main

import (
	"fmt"
	"os"
)

func printMainUsage() {
	fmt.Fprintf(os.Stderr, "Usage: %v auth|search|mylist|download [params] [-help]", os.Args[0])
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
	case "mylist":
		cmdMylist()
	case "download":
		cmdDownload()
	default:
		printMainUsage()
	}
}
