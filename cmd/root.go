package cmd

import (
	"fmt"
	"os"

	"github.com/ramonvermeulen/whosthere/internal/config"
	"github.com/ramonvermeulen/whosthere/internal/ui"
	"github.com/spf13/cobra"
)

const (
	appName      = "whosthere"
	shortAppDesc = "Local network discovery tool with a modern TUI interface."
	longAppDesc  = `Local network discovery tool with a modern TUI interface written in Go.
Discover, explore, and understand your Local Area Network in an intuitive way.

Knock Knock... who's there? 🚪`
)

var (
	rootCmd = &cobra.Command{
		Use:   appName,
		Short: shortAppDesc,
		Long:  longAppDesc,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		RunE: run,
	}
	whosthereFlags = config.NewFlags()
)

func init() {
	initWhosthereFlags()
}

// Execute is the entrypoint for the CLI application
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func run(*cobra.Command, []string) error {
	fmt.Println("Welcome to whosthere!")

	cfg, _, err := config.Load(whosthereFlags.ConfigFile)
	if err != nil {
		fmt.Println("Error loading config:", err)
		fmt.Println("Falling-back to default configuration.")
	}

	app := ui.NewApp(cfg)
	if err := app.Init(); err != nil {
		return err
	}

	if err := app.Run(); err != nil {
		return err
	}

	return nil
}

func initWhosthereFlags() {
	rootCmd.Flags().StringVarP(
		&whosthereFlags.ConfigFile,
		"config-file", "c",
		"",
		"Path to config file.",
	)
}
