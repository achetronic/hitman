// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

package version

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

const (
	descriptionShort = `Print the current version`
	descriptionLong  = `
	Version show the current hitman version client.`
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "version",
		DisableFlagsInUseLine: true,
		Short:                 descriptionShort,
		Long:                  strings.ReplaceAll(descriptionLong, "\t", ""),

		Run: RunCommand,
	}

	return cmd
}

func RunCommand(cmd *cobra.Command, args []string) {
	fmt.Print("version: {VERSION}\n")
}
