package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/zzxwill/aigit/cmd/auth"
	"github.com/zzxwill/aigit/cmd/commit"
	"github.com/zzxwill/aigit/cmd/version"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "aigit",
		Short: "Generate git commit message including title and body",
		Long:  `AI Git Commi streamlines the git commit process by automatically generating meaningful and standardized commit messages.`,
	}

	rootCmd.AddCommand(auth.SetupCmd())
	rootCmd.AddCommand(commit.SetupCmd())
	rootCmd.AddCommand(version.SetupCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
