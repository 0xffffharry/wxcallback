package main

import (
	"fmt"
	"github.com/spf13/cobra"
)

const Version = "v0.0.1-alpha"

var versionCommand = &cobra.Command{
	Use:   "version",
	Short: "Show Version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(fmt.Sprintf("%s %s", "wxcallback", Version))
	},
}

func init() {
	RootCommand.AddCommand(versionCommand)
}
