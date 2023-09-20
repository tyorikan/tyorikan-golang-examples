package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "app",
	Short: "App CLI is the main interface to the backend",
}

// Execute is called just once, by main.main()
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}
