/*
Copyright © 2022 Daniel Charpentier daniel@kreechures.com

*/
package cmd

import (
	"github.com/spf13/cobra"
)

// candyV1Cmd represents the candyV1 command
var candyV1Cmd = &cobra.Command{
	Use:   "candyV1",
	Short: "Not yet implemented.",
	Long:  `Not yet implemented.`,
	Run: func(cmd *cobra.Command, args []string) {
	},
}

func init() {
	rootCmd.AddCommand(candyV1Cmd)
}
