# Archethic Command Line Interpreter

This command line interface enables you to interact with the Archethic's blockchain within the terminal by providing snap features:
- Generate transaction address 
- Build & send transaction
- Manage keychain (decentralized wallet)
    - Create and connect to your keychain (decentralized wallet)
    - Build and send transaction from your connected keychain
    - Update your keychain by adding and removing services

![](intro.gif)

## Installation

Install it with go:

```bash
go install github.com/archethic-foundation/archethic-cli@latest
```

Or just build it yourself (requires Go 1.18+):

```bash
git clone https://github.com/archethic-foundation/archethic-cli.git
cd archethic-cli
go build .
```

## Usage

By default the CLI works as TUI (terminal user interface) application allowing the application to be interactive. 

When launching the Archethic CLI you will access to the main menu that allows you to select an action. 
- Generate an address
- Build and send a transaction
    - send uco
    - send tokens
    - interact with smart contract (recipients)
    - add ownerships and secret delegation
    - add abritraty content 
    - add smart contract's code
- Manage keychains
    - create a keychain with a given seed
    - access a keychain
    - add and remove services from a keychain
    - send a keychain transaction for a specific service

## License

[AGPL-3](/LICENCE)