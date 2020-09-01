# pegnetd

A light weight daemon that executes the PegNet protocol on Factom chains and maintains balances for addresses.

**Note:** This program is intended to be temporary and will be replaced when PegNet is integrated into the [Factom Asset Tokens daemon](https://github.com/Factom-Asset-Tokens/fatd). Usage and JSON-RPC functions are being written as similarly as possible to fatd in an effort to minimize friction during that switchover period.

## Building From Source

Ensure that [Golang 1.13](https://golang.org/) or later is installed. The latest official release of Golang is always recommended.

Clone the project and build the binary:
```
$ git clone https://github.com/pegnet/pegnetd
$ cd pegnetd
$ go build
```

If successful, there will now be a `pegnetd` executable file in the working directory.

## Configuration

By default `pegnetd` will search `$HOME/.pegnetd/pegnetd-conf.toml` or `./pegnetd-conf.toml` for your config file. The default config file is in the root directory of this repo, and can be copied to your home folder. Alternatively the `--config pegnetd-conf.toml` can also be specified. 

Once you have your config in place, running `pegnetd` will begin syncing to the current network your factomd is on.

## Running

To start the daemon, run: `$ ./pegnetd --log=debug` using your preferred log verbosity level.

To exit `pegnetd`, send a `SIGINT` (commonly done by pressing `<ctrl> + <c>` within the terminal).

## Running in Development

To run in development, the `--testing` flag will set the activation heights to 0, and grading versions to 2. So if you have a local factomd running, you can do the following:

```bash
# Assuming you have `pegnet` installed for mining and factom + factom-walletd running
cd $GOPATH/src/github.com/pegnet/pegnet
cd initialization
go build
# Get some entry credits to build the initial chains
./fundEC.sh
./initialization

# Now we can run a miner to get us some rates (you will need to configure a miner)
pegnet --testing --top 50 --miners 1 --log debug

# Now you have a chain to read and can run the node
pegnetd --testing --log debug
```

## RPC API Documentation

`// TODO: add documentation around how to use the RPC API, keeping it as close to fatd as possible`
