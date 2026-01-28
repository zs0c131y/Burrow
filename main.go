package main

import (
	"fmt"
	"os"

	"github.com/zs0c131y/burrow/cmd"
)

const version = "1.0.0"

func main() {
	if err := cmd.Execute(version); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
