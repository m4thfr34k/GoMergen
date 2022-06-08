/*
Copyright © 2022 Daniel Charpentier daniel@kreechures.com

*/
package cmd

import (
	"github.com/spf13/cobra"
)

// candyV2Cmd represents the candyV2 command
var candyV2Cmd = &cobra.Command{
	Use:   "candyV2",
	Short: "Not yet implemented.",
	Long:  `Not yet implemented.`,
	Run: func(cmd *cobra.Command, args []string) {
	},
}

func init() {
	rootCmd.AddCommand(candyV2Cmd)
}
