/*
Copyright © 2022 Daniel Charpentier daniel@kreechures.com

*/
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"time"
)

// collectionCmd represents the collection command
var collectionCmd = &cobra.Command{
	Use:   "collection",
	Short: "Saves csv file of NFT Mint addresses associated to given collection",
	Long: `Saves csv file of NFT Mint addresses associated to given collection.
File format is Magic Eden Collection Name,Collection address,NFT Address,Owner Address`,
	Run: func(cmd *cobra.Command, args []string) {
		collectionAddress, _ := cmd.Flags().GetString("coll")
		if collectionAddress != "" {
			rpcProviderURL, _ := cmd.Flags().GetString("rpc")
			if rpcProviderURL == "" {
				rpcProviderURL = rpcDefault
			}
			meBearToken, _ := cmd.Flags().GetString("meapi")
			rateLimit, _ := cmd.Flags().GetInt("ratelimit")
			retryLimit, _ := cmd.Flags().GetInt("retry")
			getCollection(collectionAddress, rpcProviderURL, meBearToken, rateLimit, retryLimit)
		} else {
			fmt.Println("collection command is missing required flag --coll")
			fmt.Println("Example: gomergen collection --coll=CRzPn8YDgZnEyaJHoYR38NphzWWRfiCEDSSAU5SPdkri")
		}
	},
}

func getCollection(collectionAddress string, rpcProvider string, meBearToken string, rateLimit int, retryLimit int) {

	//Save meta to file named collectionAddress.csv
	fmt.Print("Working")
	for i := 1; i <= (retryLimit + 1); i++ {

		transactionsMap := make(map[string]int64)

		transactionsMap, err := getTransactions(collectionAddress, rpcProvider, rateLimit)
		if err != nil {
			fmt.Println("Error occurred while getting transactions:", err.Error())
		} else {
			for txSig, _ := range transactionsMap {
				mintAddress, err := getMintFromTransaction(txSig, rpcProvider)
				if err != nil {
					fmt.Println("Error occurred while getting mint address from a transaction:", err.Error())
				} else {
					if &mintAddress == nil || mintAddress == "" {
						// Skipping due to no mintAddress to check
					} else {
						collection, owner, err := getMagicEdenData(mintAddress, meBearToken)
						if err != nil {
							fmt.Println("Error occurred while getting Magic Eden data for mint address:"+mintAddress, " - Error is:", err.Error())
						}
						err = saveCollection(mintAddress, owner, collectionAddress, collection)
						if err != nil {
							fmt.Println("Save file error when saving:", mintAddress, owner, collectionAddress, collection)
						}
					}
				}
				if rateLimit > 0 {
					time.Sleep(time.Duration(1000/rateLimit) * time.Millisecond)
				}
				print(".")
			}
		}
	}
}

func init() {
	rootCmd.AddCommand(collectionCmd)
	collectionCmd.Flags().String("coll", "", "Collection to retrieve")
}
