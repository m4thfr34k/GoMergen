/*
Copyright © 2022 Daniel Charpentier daniel@kreechures.com

*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "GoMergen",
	Short: "Returns NFT data for a given collection.",
	Long:  `GoMergen is a CLI tool to pull NFT related data for a given collection.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().String("rpc", "", "Specifies Solana RPC server to use. Default is:"+rpcDefault)
	rootCmd.PersistentFlags().String("meapi", "", "Magic Eden API Auth Token")
	rootCmd.PersistentFlags().Int("ratelimit", 0, "What is the max requests per second we should respect. 0 means NO LIMITS on the RPC!!")
	rootCmd.PersistentFlags().Int("retry", 0, "Automatically retry gathering data x number of times")
}
