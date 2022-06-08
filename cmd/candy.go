/*
Copyright © 2022 Daniel Charpentier daniel@kreechures.com

*/
package cmd

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/portto/solana-go-sdk/client"
	"github.com/portto/solana-go-sdk/common"
	"github.com/portto/solana-go-sdk/program/metaplex/tokenmeta"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// candyCmd represents the candy command
var candyCmd = &cobra.Command{
	Use:   "candy",
	Short: "Returns the candy machine account address of a token mint or Magic Eden collection, if one exists.",
	Long:  `Returns the candy machine account address of a token mint or Magic Eden collection, if one exists.`,
	Run: func(cmd *cobra.Command, args []string) {
		tokenAddress, _ := cmd.Flags().GetString("mint")
		meCollection, _ := cmd.Flags().GetString("meCOLL")
		meBearToken, _ := cmd.Flags().GetString("meapi")
		rateLimit, _ := cmd.Flags().GetInt("ratelimit")
		rpcProviderURL, _ := cmd.Flags().GetString("rpc")
		candyMachine, _ := cmd.Flags().GetString("cndy")
		if rpcProviderURL == "" {
			rpcProviderURL = rpcDefault
		}
		if tokenAddress != "" {
			candyMachine, err := getCandy(tokenAddress, rpcProviderURL)
			if err != nil {
				fmt.Println("Error in attempting to discover candy machine address:", err.Error())
			} else {
				if candyMachine != "" {
					fmt.Println(candyMachine)
				} else {
					fmt.Println("No candy machine found.")
				}
			}
		} else {
			if meCollection != "" {
				candyMap := make(map[string]string)
				candyMap, err := getMagicEdenCollectionNFTs(meCollection, meBearToken, rateLimit, rpcProviderURL)
				if err != nil {
					fmt.Println("")
					fmt.Println("Error when trying to retrieve candy machine IDs")
				} else {
					fmt.Println("")
					fmt.Println("Candy machine ID's found are:")
					for k, _ := range candyMap {
						if len(k) > 0 {
							fmt.Println(k)
						}
					}
				}
			} else {
				if candyMachine != "" {
					nftMap := make(map[string]string)
					nftMap, err := getNFTfromCandy(candyMachine, rpcProviderURL, rateLimit)
					if err != nil {
						fmt.Println("")
						fmt.Println("Error when trying to retrieve NFTs from Candy Machine")
					} else {
						fmt.Println("")
						fmt.Println("NFTs from Candy Machine found are:")
						for k, _ := range nftMap {
							if len(k) > 0 {
								fmt.Println(k)
							}
						}
					}
				} else {
					fmt.Println("Missing option. Please refer to --help")
				}
			}
		}
	},
}

func getNFTfromCandy(cndyAddress string, rpcProviderURL string, rateLimit int) (map[string]string, error) {

	nftMintAddresses := make(map[string]string)
	cndyTransactions := make(map[string]int64)

	cndyTransactions, err := getTransactions(cndyAddress, rpcProviderURL, rateLimit)
	if err != nil {
		return nftMintAddresses, err
	} else {
		for cndySignature, _ := range cndyTransactions {
			cndyNFT, err := getMintFromCandyTransaction(cndySignature, rpcProviderURL)
			if err != nil {
				// Error on attempt. Move on to next attempt.
			} else {
				if cndyNFT != "" {
					_, ok := nftMintAddresses[cndyNFT]
					if !ok {
						nftMintAddresses[cndyNFT] = "found"
						_ = saveCandyNFTs(cndyNFT, cndyAddress)
						fmt.Println(cndyNFT)
					}
				}
			}
		}
	}
	return nftMintAddresses, errors.New("No data returned")
}

func getMagicEdenCollectionNFTs(meCollection string, meAuthToken string, rateLimit int, rpcProviderURL string) (map[string]string, error) {

	// get unique NFT mint addresses from collection activities
	// get unique list of candy machine addresses for all NFTs

	const magicEdenCollectionActivitiesURL string = "https://api-mainnet.magiceden.dev/v2/collections/"

	nftMintAddresses := make(map[string]string)
	candyAddresses := make(map[string]string)

	currentOffset := 0
	returnedRecords := 1

	type MEactivity struct {
		TokenMint string `json:"tokenMint"`
	}

	fmt.Print("Working")

	// Getting the unique NFT mint addresses from collection activities
	client := resty.New()
	for returnedRecords > 0 {
		var activities []MEactivity
		_, err := client.R().
			SetAuthToken(meAuthToken).
			SetQueryParams(map[string]string{
				"offset": strconv.Itoa(currentOffset),
				"limit":  "1000",
			}).
			SetResult(&activities).
			Get(magicEdenCollectionActivitiesURL + meCollection + "/activities")
		if err != nil {
			return candyAddresses, err
		} else {
			returnedRecords = len(activities)
			for _, activity := range activities {
				_, ok := nftMintAddresses[activity.TokenMint]
				if !ok && len(activity.TokenMint) > 0 {
					nftMintAddresses[activity.TokenMint] = meCollection
				}
			}
		}
		currentOffset = currentOffset + 1000
		if rateLimit > 0 {
			time.Sleep(time.Duration(1000/rateLimit) * time.Millisecond)
		}
	}

	// Getting the unique list of candy machine addresses for all NFTs
	for mintAddress, _ := range nftMintAddresses {
		candyMachineAddress, err := getCandy(mintAddress, rpcProviderURL)
		if err != nil {
			// Error from attempt. Move to the next one.
		} else {
			fmt.Print(".")
			_, ok := candyAddresses[candyMachineAddress]
			if !ok {
				candyAddresses[candyMachineAddress] = "found"
			}
		}
	}

	return candyAddresses, nil
}

func getCandy(NFT string, rpcProviderURL string) (string, error) {

	metaAddress, err := tokenmeta.GetTokenMetaPubkey(common.PublicKeyFromString(NFT))

	c := client.NewClient(rpcProviderURL)
	accountInfo, err := c.GetAccountInfo(context.Background(), metaAddress.String())
	if err != nil {
		return "", err
	} else {
		metadata, err := tokenmeta.MetadataDeserialize(accountInfo.Data)
		if err != nil {
			return "", err
		} else {
			if &metadata == nil {
				return "", errors.New("No data returned")
			}
			for _, creator := range *metadata.Data.Creators {
				if creator.Share == 0 && creator.Verified {
					cndyInfo, err := c.GetAccountInfo(context.Background(), creator.Address.String())
					if err != nil {
						// Error in getting account info. Move on to the next one.
					} else {
						if strings.HasPrefix(cndyInfo.Owner, "cndy") {
							return creator.Address.String(), nil
						}
					}
				}
			}
		}
	}
	return "", nil
}

func saveCandyNFTs(nft string, cndyMachine string) error {

	f, err := os.OpenFile(cndyMachine+".csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	} else {
		dataLine := cndyMachine + "," +
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
	rootCmd.AddCommand(candyCmd)
	candyCmd.Flags().String("mint", "", "NFT to check for candy machine address")
	candyCmd.Flags().String("meCOLL", "", "Collection from Magic Eden to check for all candy machines")
	candyCmd.Flags().String("cndy", "", "List NFTs produced from this candy machine and saves to csv")
}
