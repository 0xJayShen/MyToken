package main

import (
	"context"
	"fmt"

	"Mytoken"
	"github.com/onflow/flow-go-sdk"
	"google.golang.org/grpc"

	"github.com/onflow/flow-go-sdk/client"
)

var setupMinterCode string = fmt.Sprintf(`
import MyToken from %s

transaction {

    prepare(minter: AuthAccount) {

        let minterProxy <- MyToken.createMinterProxy()

        minter.save(
            <- minterProxy, 
            to: MyToken.MinterProxyStoragePath,
        )
            
        minter.link<&MyToken.MinterProxy{MyToken.MinterProxyPublic}>(
            MyToken.MinterProxyPublicPath,
            target: MyToken.MinterProxyStoragePath
        )
    }
}`, MyToken.Config.MyTokenAddress)

func main() {
	ctx := context.Background()
	flowClient, err := client.New(MyToken.Config.Node, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	referenceBlock, err := flowClient.GetLatestBlock(ctx, false)
	if err != nil {
		panic(err)
	}

	acctAddress, acctKey, signer := MyToken.ServiceAccount(flowClient, MyToken.Config.SingerAddress, MyToken.Config.SingerPriv)
	tx := flow.NewTransaction().
		SetScript([]byte(setupMinterCode)).
		SetGasLimit(100).
		SetProposalKey(acctAddress, acctKey.Index, acctKey.SequenceNumber).
		SetReferenceBlockID(referenceBlock.ID).
		SetPayer(acctAddress).
		AddAuthorizer(acctAddress)

	if err := tx.SignEnvelope(acctAddress, acctKey.Index, signer); err != nil {
		panic(err)
	}

	if err := flowClient.SendTransaction(ctx, *tx); err != nil {
		panic(err)
	}

	MyToken.WaitForSeal(ctx, flowClient, tx.ID())
}
