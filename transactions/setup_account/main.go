package main

import (
	"context"
	"fmt"

	"Mytoken"
	"github.com/onflow/flow-go-sdk"
	"google.golang.org/grpc"

	"github.com/onflow/flow-go-sdk/client"
)

var setupCode string = fmt.Sprintf(`
import MyToken from %s

import FungibleToken from %s

transaction {

    prepare(signer: AuthAccount) {

        // It's OK if the account already has a Vault, but we don't want to replace it
        if(signer.borrow<&MyToken.Vault>(from: /storage/mytokenVault) != nil) {
            return
        }
        
        // Create a new MyToken Vault and put it in storage
        signer.save(<-MyToken.createEmptyVault(), to: /storage/mytokenVault)

        // Create a public capability to the Vault that only exposes
        // the deposit function through the Receiver interface
        signer.link<&MyToken.Vault{FungibleToken.Receiver}>(
            /public/mytokenReceiver,
            target: /storage/mytokenVault
        )

        // Create a public capability to the Vault that only exposes
        // the balance field through the Balance interface
        signer.link<&MyToken.Vault{FungibleToken.Balance}>(
            /public/mytokenBalance,
            target: /storage/mytokenVault
        )
    }
}`, MyToken.Config.MyTokenAddress, MyToken.Config.FungibleTokenAddress)

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
		SetScript([]byte(setupCode)).
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
