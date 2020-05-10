package main

import (
	"os"

	"github.com/mygitlab/iibbeat/cmd"

	_ "github.com/mygitlab/iibbeat/include"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
