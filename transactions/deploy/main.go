package main

import (
	"context"
	"io/ioutil"

	MyToken "Mytoken"
	"github.com/onflow/flow-go-sdk/client"
	"github.com/onflow/flow-go-sdk/templates"
	"google.golang.org/grpc"
)

func main() {
	ctx := context.Background()
	flowClient, err := client.New(MyToken.Config.Node, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	referenceBlock, err := flowClient.GetLatestBlock(context.Background(), false)
	if err != nil {
		panic(err)
	}

	serviceAcctAddr, serviceAcctKey, singer := MyToken.ServiceAccount(flowClient, MyToken.Config.SingerAddress, MyToken.Config.SingerPriv)
	name := "MyToken"

	contractPath := ""
	code, err := ioutil.ReadFile(contractPath)
	if err != nil {
		panic(err)
	}

	tx := templates.AddAccountContract(serviceAcctAddr, templates.Contract{
		Name:   name,
		Source: string(code),
	})
	tx.SetProposalKey(
		serviceAcctAddr,
		serviceAcctKey.Index,
		serviceAcctKey.SequenceNumber,
	)
	tx.SetReferenceBlockID(referenceBlock.ID)
	tx.SetPayer(serviceAcctAddr)
	tx.SetGasLimit(9999)
	if err := tx.SignEnvelope(serviceAcctAddr, serviceAcctKey.Index, singer); err != nil {
		panic(err)
	}

	if err := flowClient.SendTransaction(ctx, *tx); err != nil {
		panic(err)
	}

	MyToken.WaitForSeal(ctx, flowClient, tx.ID())
}
