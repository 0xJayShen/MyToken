package main

import (
	"context"
	"fmt"

	MyToken "Mytoken"
	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"google.golang.org/grpc"
)

var getBalanceCode string = fmt.Sprintf(`
import MyToken from %s

import FungibleToken from %s

pub fun main(address: Address): UFix64 {
    let account = getAccount(address)

    let vaultRef = account.getCapability(/public/mytokenBalance)!
        .borrow<&MyToken.Vault{FungibleToken.Balance}>()
        ?? panic("Could not borrow Balance reference to the Vault")

    return vaultRef.balance
}`, MyToken.Config.MyTokenAddress, MyToken.Config.FungibleTokenAddress)

var (
	searchAddress = ""
)

func main() {
	flowClient, err := client.New(MyToken.Config.Node, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	result, err := flowClient.ExecuteScriptAtLatestBlock(ctx, []byte(getBalanceCode), []cadence.Value{cadence.NewAddress(flow.HexToAddress(searchAddress))})
	if err != nil {
		panic(err)
	}

	fmt.Println(MyToken.CadenceValueToJsonString(result))
}
