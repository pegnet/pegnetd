# Addressing in Pegnet

Factom addressing supports multiple address types "under the hood" via a Redeem Condition Datastructure (RCD). There is only 1 supported RCD type by the Factom Protocol at the moment, RCD-1. It uses the ed25519 signature scheme. In order to support an ethereum bridge, a new RCD type is proposed.

## RCD-e

RCD-e uses the RCD `0x0e + 64 byte ecdsa public key`. Signature validation will use ecdsa to be compatible with Ethereum. The encoding of the resulting address is the same as before, `0x5fb1 + sha256d(rcd) + checksum`. This means all addresses on Factom will still follow the standard `FA...` format.

This RCD will be implemented in Pegnet before being implemented into factomd and the Factom Ecosystem. Because of this, this RCD format cannot move factoids at this time. There is a risk users will send FCT to this new format. To combat user mistakes, when making this new RCD-e into an address, it is recommended to prefix the address with `e`. So the resulting address will be something like `eFA3ZeHBmPHiFeMaTmGiArgNR33VM2zRqN2LjzP5FTf18AVR5dkqC`. This will prevent any older wallets from sending FCT, as the prefix makes the address no longer valid. The pegnet ecosystem will know how to interpret the `e` prefix and handle the checks accordingly.

The purpose of this new rcd is to enable an Ethereum gateway. In order for the Gateway to work, only `eFA` addresses can enter the gateway. So another user mistake would be to send funds from an `FA` address. To prevent this second user mistake, addresses prefixed with a capital `E` will only accept funds from an `eFA` address. This means the Gateway can provide an `EFA` address to move into ethereum, and the pegnet applications will better protect against mistakes. Any user can change their `FA` address as an `eFA` or `EFA`. This is only to enforce some client side protections.

**All these user protections are enforced by the wallet, not by the protocol**


#### Implementation Details

The RCD will behave as another normal rcd type for all outputs. Inputs will use ecdsa vs ed25519. Users will be presented with an `eFA` address, but this `eFA` address should correspond to a `0x123...` address on ethereum. To do this, all `eFA` addresses should also somehow indicate their ethereum equivalent address. No pegnet/factom application should accept the eth address as an output, but it should be displayed for comparison purposes to ensure the other side of the gateway will line up.

All generated `eFA` addresses should use the Ethereum bip44 derivation path.

All imported `eFA` addresses should use the Ethereum hex format.

#### Clientside Transaction Rules

A quick table of the allowed transfers. The allowed column allows what addresses can receive the given currency from the input address. Remember all these are client side enforced.

| Input | Currency |    Allowed   |
|-------|----------|:------------:|
| FA    | FCT      | FA           |
| FA    | pXXX     | FA, eFA      |
|       |          |              |
| eFA   | FCT      |      x       |
| eFA   | pXXX     | FA, eFA, EFA |
|       |          |              |
| EFA   | FCT      |      x       |
| EFA   | pXXX     | eFA, EFA     |



# Technical Information

**RCD Type e**:

| data | Field Name | Description |
| ----------------- | ---------------- | --------------- |
| varInt_F | Type | The RCD type.  This specifies how the datastructure should be interpreted.  Type 14 `(0x0e)` indicates the Ethereum linked rcd.|
| 64 bytes | Pubkey 0 | The 64 byte ecdsa uncompressed public key. If generating an uncompressed key, and the length is 65, remove the `0x04` prefix indicating an uncompressed key. |

The private key is the 32 byte ecdsa private key. When represented as text, the Ethereum hex style `0x00..` should be used.

Signatures are 65 bytes in length. The first 64 bytes is the signature itself, with the 65th byte being the recovery byte. The recovery byte allows the operation `pubkey = recover(digest, signature)`.

Example Vector:
```
Private Key: 0xbde0723b3236d7b7613d11c6c93c57ad89fd7f7c586aaa18f7a1b392aa2c39fd
 Public Key: 0x25892ecbaf10c71d52f260c0b43a8e6b2384324b02c0d4b576d74452b81c359068e584b8831287a1a21b21fd2ac0c0e326cfcdaff27df14261a78bee0e0177d4
  FAAddress: FA2M5uK3aPrJ8RfHdbTUL1ySxp22s3TwQ3Js6n7jss7QErw3S6ae
 EthAddress: 0xD8A27BdCA2067F233551F29231CD2cDa0828bF62

     Digest: 0000000000000000000000000000000000000000000000000000000000000000
  Signature: 4bc4cee5b07114dcc0dd4a7654d42a1abc51f8d5c417fde8413686126a8cfb0633fac598664c46ec88be878f8fe9c3245e8385e00be107f021a937ab73d09d7a00
```
