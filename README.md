# Setting up Blobstream X for SEQ.

This is a computationally intensive task:
AWS instance recomended: r6a.16xlarge or above

clone the patched blobstream X:
```shell
   git clone https://github.com/anomalyFi/blobstreamx.git
```
download artifacts from [here](https://hackmd.io/@succinctlabs/HJE7XRrup#Download-Blobstream-X-Plonky2x-Circuits): @todo remove this and make a single artifact bucket

download the docker image from [patched succinctx repo](https://github.com/AnomalyFi/succinctx/blob/main/client/src/Dockerfile) and build the image with the name `local_prover`
```shell 
docker build . -t local_prover
```
download stark-snark verifer: 
```shell
curl -L https://public-circuits.s3.amazonaws.com/verifier-build13.tar.gz | tar xz
```
Build the circuits locally or download the circuits [from]() @todo upload binaries and add link.
```shell
mkdir -p build && RUST_LOG=debug cargo run --bin header_range_1024 --release build && mv ./target/release/header_range_1024 ./build/header_range_1024

RUST_LOG=debug cargo run --bin next_header --release build && mv ./target/release/next_header ./build/next_header
```
replace header_range_1024 in artifacts/header-range-1024 with the header_range_1024 file in build
replace next_header in artifacts/next-header with next_header file in build
clone succinctx repo and build the verifier or download from [here]() @todo upload the binaries
```shell
   git clone https://github.com/anomalyfi/succinctx.git && cd succinctx/plonky2x/verifier
   go build
```
replace the verifier in artifacts/verifier-build with verifier built now in succinctx folder.
copy Verifier.sol file in artifacts/verifier-build and place it in the verifier-build folder.
complete the .env file

and then run the local prover and local relayer:
```shell
cargo run --bin blobstreamx --release
```
# Blobstream X

![Blobstream X](https://pbs.twimg.com/media/F85boT-bYAAF1hM?format=jpg&name=4096x4096)

Implementation of zero-knowledge proof circuits for [Blobstream](https://docs.celestia.org/developers/blobstream), Celestia's data availability solution for Ethereum.

## Overview

Blobstream X's core contract is `BlobstreamX`, which stores commitments to ranges of data roots from Celestia blocks. Users can query for the validity of a data root of a specific block height via `verifyAttestation`, which proves that the data root is a leaf in the Merkle tree for the block range the specific block height is in.

## Request BlobstreamX Proofs

### Request Proofs from the Succinct Platform

Add env variables to `.env`, following the `.env.example`. You do not need to fill out the local configuration, unless you're planning on doing local proving.

Run `BlobstreamX` script to request updates to the specified light client continuously. For the cadence of requesting updates, update `LOOP_DELAY_MINUTES`.

In `/`, run

```

cargo run --bin blobstreamx --release

```

### [Generate & Relay Proofs Locally](https://hackmd.io/@succinctlabs/HJE7XRrup)

## BlobstreamX Contract Overview

### Contract Deployment

To deploy the `BlobstreamX` contract:

1. Get the genesis parameters for a `BlobstreamX` contract from a specific Celestia block.

   ```shell
   cargo run --bin genesis -- --block <genesis_block>
   ```

2. Add .env variables to `contracts/.env`, following `contracts/.env.example`.
3. Initialize `BlobstreamX` contract with genesis parameters. In `contracts`, run

   ```shell
   forge install

   source .env

   forge script script/Deploy.s.sol --rpc-url $RPC_URL --private-key $PRIVATE_KEY --broadcast --verify --verifier etherscan --etherscan-api-key $ETHERSCAN_API_KEY
   ```

### Succinct Gateway Prover Whitelist

#### Set Whitelist Status

Set the whitelist status of a functionID to Default (0), Custom (1) or Disabled (2).

```shell
cast calldata "setWhitelistStatus(bytes32,uint8)" <YOUR_FUNCTION_ID> <WHITELIST_STATUS>
```

#### Add Custom Prover

Add a custom prover for a specific functionID.

```shell
cast calldata "addCustomProver(bytes32,address)" <FUNCTION_ID> <CUSTOM_PROVER_ADDRESS>
```
