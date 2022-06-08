/*
Copyright © 2022 Daniel Charpentier daniel@kreechures.com

*/

package cmd

import (
	"context"
	"errors"
	"github.com/go-resty/resty/v2"
	"github.com/portto/solana-go-sdk/client"
	"github.com/portto/solana-go-sdk/common"
	"github.com/portto/solana-go-sdk/program/metaplex/tokenmeta"
	"github.com/portto/solana-go-sdk/rpc"
	"os"
	"time"
)

const rpcDefault string = "https://ssc-dao.genesysgo.net/"

func getTransactions(acctAddress string, rpcProviderURL string, rateLimit int) (map[string]int64, error) {

	/*
		Returns signatures of all transactions related to acctAddress
		getSignaturesForAddress returns a max of 1k signatures so must loop until no further results are returned
	*/

	var totalSignatures int = 1
	var earliestSignature string = ""
	var earliestBlocktime int64 = 99999999999999999

	var tempCounter int = 0

	transactionsMap := make(map[string]int64)

	c := client.NewClient(rpcProviderURL)

	for totalSignatures > 0 {
		txInfo, err := c.GetSignaturesForAddressWithConfig(context.Background(), acctAddress, rpc.GetSignaturesForAddressConfig{
			Limit:  1000,
			Before: earliestSignature,
		})
		if err != nil {
			return transactionsMap, err
		} else {
			totalSignatures = len(txInfo)
			for _, result := range txInfo {
				tempCounter++
				transactionsMap[result.Signature] = *result.BlockTime
				if *result.BlockTime <= earliestBlocktime {
					earliestBlocktime = *result.BlockTime
					earliestSignature = result.Signature
				}
			}
		}
		if rateLimit > 0 {
			time.Sleep(time.Duration(1000/rateLimit) * time.Millisecond)
		}
	}
	return transactionsMap, nil
}

func getMintFromTransaction(txSignature string, rpcProviderURL string) (string, error) {

	c := client.NewClient(rpcProviderURL)
	txInfo, err := c.GetTransaction(context.Background(), txSignature)
	if err != nil {
		return "", err
	} else {
		if txInfo == nil {
			return "", nil
		}
		if txInfo.Transaction.Message.DecompileInstructions() != nil {
			for _, instructV := range txInfo.Transaction.Message.DecompileInstructions() {
				for _, singleAccount := range instructV.Accounts {
					if (singleAccount.IsWritable) && (!singleAccount.IsSigner) {
						metaAddress := singleAccount.PubKey.String()
						accountInfo, err := c.GetAccountInfo(context.Background(), metaAddress)
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
								return metadata.Mint.String(), nil
							}
						}
					}
				}
			}
		}
		return "", nil
	}
}

func getMintFromCandyTransaction(txSignature string, rpcProviderURL string) (string, error) {

	c := client.NewClient(rpcProviderURL)
	txInfo, err := c.GetTransaction(context.Background(), txSignature)
	if err != nil {
		return "", err
	} else {
		if txInfo == nil {
			return "", nil
		}
		for _, balance := range txInfo.Meta.PostTokenBalances {
			if balance.UITokenAmount.Amount == "1" {
				// Checking to see if this item has an actual metadata account associated to it
				metadataAccount, err := tokenmeta.GetTokenMetaPubkey(common.PublicKeyFromString(balance.Mint))
				if err != nil {
					return "", err
				} else {
					accountInfo, err := c.GetAccountInfo(context.Background(), metadataAccount.String())
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
							return metadata.Mint.String(), nil
						}
					}
				}
			}
		}
		return "", nil
	}
}

func getMagicEdenData(mintAddress string, meAuthToken string) (collection string, owner string, err error) {

	const magicEdenMintInfoURL string = "https://api-mainnet.magiceden.dev/v2/tokens/"

	type NFT struct {
		Owner      string `json:"owner"`
		Collection string `json:"collection"`
		Name       string `json:"name"`
		Image      string `json:"image"`
	}

	client := resty.New()
	resp, err := client.R().
		SetAuthToken(meAuthToken).
		SetResult(&NFT{}).
		Get(magicEdenMintInfoURL + mintAddress)
	if err != nil {
		return "", "", err
	} else {
		nft := resp.Result().(*NFT)
		if nft == nil {
			return "", "", errors.New("No data returned")
		} else {
			return nft.Collection, nft.Owner, nil
		}
	}
}

func saveCollection(nft string, owner string, collection string, meCollectionName string) error {

	f, err := os.OpenFile(collection+".csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	} else {
		dataLine := meCollectionName + "," +
			collection + "," +
			nft + "," +
			owner + "\n"
		if _, err := f.Write([]byte(dataLine)); err != nil {
			return err
		}
		if err = f.Close(); err != nil {
			return err
		}
	}
	return nil
}
