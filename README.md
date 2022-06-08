
<h1 align="center">GoMergen</h1>
<div align="center">
	<img alt="GitHub go.mod Go version" src="https://img.shields.io/github/go-mod/m4thfr34k/GoMergen">
</div>

GoMergen is a CLI tool to gather data related to NFTs on the Solana blockchain

# Guide

## TODO
- [ ] Code cleanup
- [ ] Refactor flags to align with specific commands. Some flags are currently globally available when they only make sense for certain commands.
- [ ] Add global flag to save output to csv
- [ ] Better error checking and handling
- [ ] Consolodate file save functions 

## Examples

Retrieves NFTs from Magic Eden by:
* scanning all ME activities related to the collection
* scanning wallets associated to these activities for unlisted NFTs in this collection
Saves to a csv file named {collection name}.csv
```sh
gomergen magiceden --coll=kreechures
```

Returns the candy machine account address of a token mint or Magic Eden collection, if one exists.

```sh
gomergen candy --mint=FggC7na23EqzC3kjezjLbEhkcfFiBLZVUqNVrrBAx5UU
```

Returns the candy machine account address of a Magic Eden collection, if one exists.

```sh
gomergen candy --meCOLL=kreechures
```