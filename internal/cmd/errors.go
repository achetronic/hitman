// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"fmt"
	"os"
)

// CheckError prints err to stderr and exits with code 1 if err is not nil. Otherwise, it is a no-op.
func CheckError(err error) {
	if err != nil {
		if err != context.Canceled {
			fmt.Fprintf(os.Stderr, "An error occurred: %v\n", err)
		}
		os.Exit(1)
	}
}
