// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

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
