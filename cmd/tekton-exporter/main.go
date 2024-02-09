package main

import (
	"os"
	"path/filepath"
	"tekton-exporter/internal/cmd"
)

func main() {
	baseName := filepath.Base(os.Args[0])

	err := cmd.NewMetricosoCommand(baseName).Execute()
	cmd.CheckError(err)
}
