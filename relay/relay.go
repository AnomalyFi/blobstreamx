package main

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"strconv"
	"strings"

	"github.com/AnomalyFi/hypersdk/crypto/ed25519"
	"github.com/AnomalyFi/hypersdk/rpc"
	"github.com/AnomalyFi/nodekit-seq/actions"
	"github.com/AnomalyFi/nodekit-seq/auth"
	wrpc "github.com/AnomalyFi/nodekit-seq/rpc"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"
	"github.com/joho/godotenv"
)

var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

type CommitHeaderRangeInput struct {
	TargetBlock uint64
	Input       []byte
	Output      []byte
	Proof       []byte
}

type Output struct {
	Calldata   string `json:"calldata"`
	ChainID    uint64 `json:"chain_id"`
	FunctionID string `json:"function_id"`
	Input      string `json:"input"`
	Output     string `json:"output"`
	To         string `json:"to"`
}

var MainMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"targetBlock\",\"type\":\"uint64\"},{\"internalType\":\"bytes\",\"name\":\"input\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"output\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"proof\",\"type\":\"bytes\"}],\"internalType\":\"structinputs.CommitHeaderRangeInput\",\"name\":\"_c\",\"type\":\"tuple\"}],\"name\":\"dummyCommitHeaderRangeInput\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"height\",\"type\":\"uint64\"},{\"internalType\":\"bytes32\",\"name\":\"header\",\"type\":\"bytes32\"}],\"internalType\":\"structinputs.InitializerInput\",\"name\":\"_i\",\"type\":\"tuple\"}],\"name\":\"dummyInitializerInput\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"targetHeader\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"dataCommitment\",\"type\":\"bytes32\"}],\"internalType\":\"structinputs.OutputBreaker\",\"name\":\"_o\",\"type\":\"tuple\"}],\"name\":\"dummyOutputBreaker\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"_tupleRootNonce\",\"type\":\"uint256\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"height\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"dataRoot\",\"type\":\"bytes32\"}],\"internalType\":\"structinputs.DataRootTuple\",\"name\":\"_tuple\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"bytes32[]\",\"name\":\"sideNodes\",\"type\":\"bytes32[]\"},{\"internalType\":\"uint256\",\"name\":\"key\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"numLeaves\",\"type\":\"uint256\"}],\"internalType\":\"structinputs.BinaryMerkleProof\",\"name\":\"_proof\",\"type\":\"tuple\"}],\"internalType\":\"structinputs.VerifyAttestationInput\",\"name\":\"_v\",\"type\":\"tuple\"}],\"name\":\"dummyVerifyAttestationInput\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}
var MainABI, _ = MainMetaData.GetAbi()

func main() {
	ctx := context.Background()
	err := godotenv.Load()
	if err != nil {
		panic(fmt.Errorf("cant load env: %s", err))
	}
	uri := os.Getenv("RELAY_RPC_URL")
	cli := rpc.NewJSONRPCClient(uri)
	contractAddress, err := ids.FromString(os.Getenv("CONTRACT_ADDRESS"))
	if err != nil {
		panic(fmt.Errorf("invalid contract address: %s", err))
	}
	networkID, err := strconv.Atoi(os.Getenv("NETWORK_ID"))
	if err != nil {
		panic(fmt.Errorf("invalid network id: %s", err))
	}
	chainID, err := ids.FromString(os.Getenv("CHAIN_ID"))
	if err != nil {
		panic(fmt.Errorf("invalid chain id: %s", err))
	}
	p := common.FromHex(os.Getenv("PRIVATE_KEY"))
	factory := auth.NewED25519Factory(ed25519.PrivateKey(p))
	outputRaw, err := os.Open("./relay/output.json")
	if err != nil {
		panic(err)
	}
	o, err := ioutil.ReadAll(outputRaw)
	var output Output
	err = json.Unmarshal(o, &output)
	if err != nil {
		panic(err)
	}
	proofRaw, err := os.Open("./relay/proof.bin")
	if err != nil {
		panic(err)
	}
	inputF := common.Hex2Bytes(output.Input)
	l := len(inputF)
	targetBlock := binary.BigEndian.Uint64(inputF[l-8:])
	pr, err := ioutil.ReadAll(proofRaw)
	// encode the input
	input := CommitHeaderRangeInput{
		TargetBlock: targetBlock,
		Input:       common.Hex2Bytes(output.Input),
		Output:      common.Hex2Bytes(output.Output),
		Proof:       pr,
	}
	packed, err := abi.ABI.Pack(*MainABI, "dummyCommitHeaderRangeInput", input)
	if err != nil {
		fmt.Print(err)
	}
	packed = packed[4:] // remove the function id
	transact := actions.Transact{
		FunctionName:    "commit_header_range",
		ContractAddress: contractAddress,
		Input:           packed,
	}

	wcli := wrpc.NewJSONRPCClient("", uint32(networkID), chainID)
	parser, err := wcli.Parser(ctx)
	if err != nil {
		panic(fmt.Errorf("can't get parser: %s", err))
	}
	f, tx, _, err := cli.GenerateTransaction(ctx, parser, nil, &transact, factory)
	if err != nil {
		panic(fmt.Errorf("cant generate tx: %s", err))
	}
	err = f(ctx)
	if err != nil {
		panic(fmt.Errorf("cant send tx: %s", err))
	}
	log.Info("successfully submitted tx, txID: ", tx.ID())
}
