package cmd

import (
	"github.com/spf13/cobra"
)

var fileCmd = &cobra.Command{
	Use:   "file",
	Short: "Push/pull files",

	Run: func(cmd *cobra.Command, args []string) {
		cmd.Usage()
	},
}
