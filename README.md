# pegnetd

A light weight daemon that executes the PegNet protocol on Factom chains and maintains balances for addresses.

**Note:** This program is intended to be temporary and will be replaced when PegNet is integrated into the [Factom Asset Tokens daemon](https://github.com/Factom-Asset-Tokens/fatd). Usage and JSON-RPC functions are being written as similarly as possible to fatd in an effort to minimize friction during that switchover period.

## Building From Source

Ensure that [Golang](https://golang.org/) or later is installed. The latest official release of Golang is always recommended.

Clone the project and build the binary:
```
$ git clone https://github.com/pegnet/pegnetd
$ cd pegnetd
$ go build
```

If successful, there will now be a `pegnetd` executable file in the working directory.

## Configuration

`// TODO: add documentation around config options here`

## Running

To start the daemon, run: `$ ./pegnetd --log=debug` using your preferred log verbosity level.

To exit `pegnetd`, send a `SIGINT` (commonly done by pressing `<ctrl> + <c>` within the terminal).

## RPC API Documentation

`// TODO: add documentation around how to use the RPC API, keeping it as close to fatd as possible`
