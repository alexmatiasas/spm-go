// Package main
package main

import (
	"fmt"
	"os"
)

var (
	version = "dev"
	commit  = "unknown"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Printf("spm %s (commit %s)\n", version, commit)
		return
	}
	fmt.Println("spm — Smart Project Manager")
	fmt.Println("Run 'spm --help' for usage information.")
}
