package main

import (
	"context"
	"fmt"

	MyToken "Mytoken"
	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk/client"
	"google.golang.org/grpc"
)

var getTotalSupply string = fmt.Sprintf(`
import MyToken from %s

pub fun main(): UFix64 {
    return MyToken.totalSupply
}`, MyToken.Config.MyTokenAddress)

func main() {
	flowClient, err := client.New(MyToken.Config.Node, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	result, err := flowClient.ExecuteScriptAtLatestBlock(ctx, []byte(getTotalSupply), []cadence.Value{})
	if err != nil {
		panic(err)
	}

	fmt.Println(MyToken.CadenceValueToJsonString(result))
}
