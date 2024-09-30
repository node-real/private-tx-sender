# private-tx-sender

## Introduction

The private-tx-sender provides enhanced transaction privacy for the BNB Smart Chain (BSC) network.
By packing a transaction into a bundle and broadcast it to mev-builders who implementing
the [MEV standards](https://docs.bnbchain.org/bnb-smart-chain/validator/mev/overview/),
the following capabilities are provided:

1. Privacy: All transactions sent through this API will not be propagated on the P2P network,
   hence, they won't be detected by any third parties. This effectively prevents transactions from being targeted by
   sandwich attacks.

2. Fast Confirmation: By integrating multiple builders, the transaction will be included in the blocks mined by all
   the mev-validators integrated by all builders.

##  Quick Start Examples

The example directory provides an example to use the SDK to send a private transaction.

### Config Examples

```toml
[Sender]
ChainURL = "http"
BlockInterval = "3s"
BundleLifeNumber = 21

[[Bundler.Builders]]
Brand = "nodereal"
URL = "https://bsc-mainnet-builder-us.nodereal.io"

[[Bundler.Builders]]
Brand = "puissant"
URL = "https://puissant-builder.48.club"

[[Bundler.Builders]]
Brand = "txboost"
URL = "https://fastbundle-us.blocksmith.org"
Key = "Basic xxxxx"
```

### Run Examples
The steps to run example are as follows

```shell
make example
cd example
./example --config config.toml --privatekey 1bb2....7ca7
```
