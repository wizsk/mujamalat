// make file
package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usages: go run m/m.go <command>")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "ins":
		installCurr()
	default:
		fmt.Fprintln(os.Stderr, "Unknown command:", os.Args[1])
		os.Exit(1)
	}
}
