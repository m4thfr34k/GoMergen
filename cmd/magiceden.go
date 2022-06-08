/*
Copyright © 2022 Daniel Charpentier daniel@kreechures.com

*/
package cmd

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
	"os"
	"strconv"
	"time"
)

// magicedenCmd represents the magiceden command
var magicedenCmd = &cobra.Command{
	Use:   "magiceden",
	Short: "Lists all NFTs for a given collection on Magic Eden.",
	Long:  `Lists all NFTs for a given collection on Magic Eden by scanning ME activities.`,
	Run: func(cmd *cobra.Command, args []string) {
		meCollection, _ := cmd.Flags().GetString("coll")
		meBearToken, _ := cmd.Flags().GetString("meapi")
		rateLimit, _ := cmd.Flags().GetInt("ratelimit")
		retryLimit, _ := cmd.Flags().GetInt("retry")
		rpcProviderURL, _ := cmd.Flags().GetString("rpc")
		if rpcProviderURL == "" {
			rpcProviderURL = rpcDefault
		}
		if meCollection != "" {
			nftMintAddresses := make(map[string]string)
			nftMintAddresses, err := getMagicEdenNFTs(meCollection, meBearToken, rateLimit, retryLimit, rpcProviderURL)
			if err != nil {
				fmt.Println("")
				fmt.Println("Error in attempting to discover NFTs in collection:", err.Error())
			} else {
				fmt.Println("")
				fmt.Println(len(nftMintAddresses), " found during this run")
			}

		} else {
			fmt.Println("Please provide the Magic Eden collection to search.")
		}

	},
}

func getMagicEdenNFTs(meCollection string, meAuthToken string, rateLimit int, retryLimit int, rpcProviderURL string) (map[string]string, error) {

	// TODO Refactor this func. It's too long and insane. Working <> good
	// get unique NFT mint addresses from collection activities
	// get unique NFT mint addresses from wallets found in collections

	const magicEdenCollectionActivitiesURL string = "https://api-mainnet.magiceden.dev/v2/collections/"
	const magicEdenWalletContentsURL string = "https://api-mainnet.magiceden.dev/v2/wallets/"
	const magicEdenTokenMintURL string = "https://api-mainnet.magiceden.dev/v2/tokens/"

	nftMintAddresses := make(map[string]string)

	for i := 1; i <= (retryLimit + 1); i++ {

		nftWallet := make(map[string]string)
		allNFTsFromWallets := make(map[string]string)

		currentOffset := 0
		returnedRecords := 1

		type MEactivity struct {
			TokenMint string `json:"tokenMint"`
			Seller    string `json:"seller"`
			Buyer     string `json:"buyer"`
		}

		fmt.Print("Working")

		// Getting the unique NFT mint addresses from collection activities
		clientResty := resty.New()
		for returnedRecords > 0 {
			var activities []MEactivity
			_, err := clientResty.R().
				SetAuthToken(meAuthToken).
				SetQueryParams(map[string]string{
					"offset": strconv.Itoa(currentOffset),
					"limit":  "1000",
				}).
				SetResult(&activities).
				Get(magicEdenCollectionActivitiesURL + meCollection + "/activities")
			if err != nil {
				return nftMintAddresses, err
			} else {
				returnedRecords = len(activities)
				for _, activity := range activities {
					fmt.Print(".")
					_, ok := nftMintAddresses[activity.TokenMint]
					if !ok && len(activity.TokenMint) > 0 {
						nftMintAddresses[activity.TokenMint] = meCollection
						_ = saveMECollectionNFTs(activity.TokenMint, meCollection)
					}

					//Getting unique wallets that hold these NFTs so we can check for more from the same collection
					_, ok = nftWallet[activity.Seller]
					if !ok && len(activity.TokenMint) > 0 && len(activity.Seller) > 0 {
						nftWallet[activity.Seller] = "found"
					}

					_, ok = nftWallet[activity.Buyer]
					if !ok && len(activity.TokenMint) > 0 && len(activity.Buyer) > 0 {
						nftWallet[activity.Buyer] = "found"
					}
				}
			}
			currentOffset = currentOffset + 1000
			if rateLimit > 0 {
				time.Sleep(time.Duration(1000/rateLimit) * time.Millisecond)
			}
		}

		//loop thru wallets and get all nft's stored in the wallets
		//this data is coming from Magic Eden
		for walletAddress, _ := range nftWallet {
			currentWalletOffset := 0
			returnedWalletRecords := 1
			type WalletContents struct {
				NFTaddress string `json:"mintAddress"`
			}
			for returnedWalletRecords > 0 {
				var nftsFromWallet []WalletContents
				_, err := clientResty.R().
					SetAuthToken(meAuthToken).
					SetQueryParams(map[string]string{
						"offset":     strconv.Itoa(currentWalletOffset),
						"limit":      "500",
						"listStatus": "both",
					}).
					SetResult(&nftsFromWallet).
					Get(magicEdenWalletContentsURL + walletAddress + "/tokens")
				if err != nil {
					// Error in current attempt. Move to next attempt.
				} else {
					returnedWalletRecords = len(nftsFromWallet)
					for _, walletData := range nftsFromWallet {
						fmt.Print(".")
						_, ok := allNFTsFromWallets[walletData.NFTaddress]
						if !ok && len(walletData.NFTaddress) > 0 {
							allNFTsFromWallets[walletData.NFTaddress] = "found"
						}
					}
				}
				currentWalletOffset = currentWalletOffset + 500
				if rateLimit > 0 {
					time.Sleep(time.Duration(1000/rateLimit) * time.Millisecond)
				}
			}
		}

		//loop thru NFTs from wallets and get all nft's that match the collection we are looking for
		for singleNFT, _ := range allNFTsFromWallets {
			type NFTData struct {
				Collection      string `json:"collection"`
				UpdateAuthority string `json:"updateAuthority"`
			}

			var nftsData NFTData
			_, err := clientResty.R().
				SetAuthToken(meAuthToken).
				SetResult(&nftsData).
				Get(magicEdenTokenMintURL + singleNFT)
			if err != nil {
				// Error in current attempt. Move to next attempt.
			} else {
				fmt.Print(".")
				if nftsData.Collection == meCollection {
					_, ok := nftMintAddresses[singleNFT]
					if !ok && len(singleNFT) > 0 {
						nftMintAddresses[singleNFT] = "found"
						_ = saveMECollectionNFTs(singleNFT, meCollection)
					}
				}
			}
			if rateLimit > 0 {
				time.Sleep(time.Duration(1000/rateLimit) * time.Millisecond)
			}
		}
	}

	return nftMintAddresses, nil
}

func saveMECollectionNFTs(nft string, meCollection string) error {

	f, err := os.OpenFile(meCollection+".csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	} else {
		dataLine := meCollection + "," +
			nft + "\n"
		if _, err := f.Write([]byte(dataLine)); err != nil {
			return err
		}
		if err = f.Close(); err != nil {
			return err
		}
	}
	return nil
}

func init() {
	rootCmd.AddCommand(magicedenCmd)
	magicedenCmd.Flags().String("coll", "", "Collection from Magic Eden")
}
