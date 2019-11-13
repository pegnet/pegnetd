# Addressing in Pegnet

Factom addressing supports multiple address types "under the hood" via a Redeem Condition Datastructure (RCD). There is only 1 supported RCD type by the Factom Protocol at the moment, RCD-1. It uses the ed25519 signature scheme. In order to support an etheruem bridge, a new RCD type is proposed.

## RCD-e

RCD-e uses the RCD `0x0e + 64 byte ecdsa public key` (TODO: Check this is enough). Signature validation will use ecdsa to be compatible with Etheruem. The encoding of the resulting address is the same as before, `0x5fb1 + sha256d(rcd) + checksum`. This means all addresses on Factom will still follow the standard `FA...` format.

This RCD will be implemented in Pegnet before being implemented into factomd and the Factom Ecosystem. Because of this, this RCD format cannot move factoids at this time. There is a risk users will send FCT to this new format. To combat user mistakes, when making this new RCD-e into an address, it is recommended to prefix the address with `e`. So the resulting address will be something like `eFA3ZeHBmPHiFeMaTmGiArgNR33VM2zRqN2LjzP5FTf18AVR5dkqC`. This will prevent any older wallets from sending FCT, as the prefix makes the address no longer valid. The pegnet ecosystem will know how to interpret the `e` prefix and handle the checks accordingly.

The purpose of this new rcd is to enable an Etheruem gateway. In order for the Gateway to work, only `eFA` addresses can enter the gateway. So another user mistake would be to send funds from an `FA` address. To prevent this second user mistake, addresses prefixed with a capital `E` will only accept funds from an `eFA` address. This means the Gateway can provide an `EFA` address to move into etheruem, and the pegnet applications will better protect against mistakes. Any user can change their `FA` address as an `eFA` or `EFA`. This is only to enforce some client side protections.

**All these user protections are enforced by the wallet, not by the protocol**


#### Implementation Details

The RCD will behave as another normal rcd type for all outputs. Inputs will use ecdsa vs ed25519. Users will be presented with an `eFA` address, but this `eFA` address should correspond to a `0x123...` address on etheruem. To do this, all `eFA` addresses should also somehow indicate their etheruem equivalent address. No pegnet/factom application should accept the eth address as an output, but it should be displayed for comparison purposes to ensure the other side of the gateway will line up.

All generated `eFA` addresses should use the Etheruem bip44 derivation path.

TODO: How to import an etheruem private key

#### Clientside Transaction Rules

A quick table of the allowed transfers. The allowed column allows what addresses can receive the given currency from the input address. Remember all these are client side enforced.

| Input | Currency |    Allowed   |
|-------|----------|:------------:|
| FA    | FCT      | FA           |
| FA    | pXXX     | FA, eFA      |
|       |          |              |
| eFA   | FCT      |              |
| eFA   | pXXX     | FA, eFA, EFA |
|       |          |              |
| EFA   | FCT      |              |
| EFA   | pXXX     | eFA, EFA     |