package main

import "github.com/spf13/cobra"

var RootCommand = &cobra.Command{
	Use: "wxcallback",
}

func main() {
	RootCommand.Execute()
}
