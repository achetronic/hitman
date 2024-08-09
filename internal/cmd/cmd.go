package cmd

import (
	"strings"

	"github.com/spf13/cobra"

	"hitman/internal/cmd/run"
	"hitman/internal/cmd/version"
)

const (
	descriptionShort = `TODO`
	descriptionLong  = `
	A super TODO.
	`
)

func NewRootCommand(name string) *cobra.Command {
	c := &cobra.Command{
		Use:   name,
		Short: descriptionShort,
		Long:  strings.ReplaceAll(descriptionLong, "\t", ""),
	}

	c.AddCommand(
		version.NewCommand(),
		run.NewCommand(),
	)

	return c
}
