package main

import (
	"os"
	"path/filepath"

	"hitman/internal/cmd"
)

func main() {
	baseName := filepath.Base(os.Args[0])

	err := cmd.NewRootCommand(baseName).Execute()
	cmd.CheckError(err)
}
