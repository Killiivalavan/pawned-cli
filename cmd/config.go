package cmd

import (
	"chesshell-cli/internal/store"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage local configuration.",
}

var configUnicodeCmd = &cobra.Command{
	Use:   "unicode [on|off]",
	Short: "Turn Unicode board rendering on or off.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		val := args[0]
		var unicode bool
		if val == "on" {
			unicode = true
		} else if val == "off" {
			unicode = false
		} else {
			fmt.Println("Error: Invalid value. Use 'on' or 'off'.")
			os.Exit(1)
		}

		data, err := store.Load()
		if err != nil && err != store.ErrCorruptedFile {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		data.Config.Unicode = &unicode

		if err := store.Save(data); err != nil {
			fmt.Printf("Error saving config: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Unicode mode turned %s.\n", val)
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configUnicodeCmd)
}
