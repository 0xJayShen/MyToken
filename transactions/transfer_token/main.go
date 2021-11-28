package main

import (
	"context"
	"fmt"

	"Mytoken"
	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"google.golang.org/grpc"

	"github.com/onflow/flow-go-sdk/client"
)

var transferCode string = fmt.Sprintf(`
import MyToken from %s

import FungibleToken from %s

transaction(amount: UFix64, to: Address) {

    // The Vault resource that holds the tokens that are being transferred
    let sentVault: @FungibleToken.Vault

    prepare(signer: AuthAccount) {
        // Get a reference to the signer's stored vault
        let vaultRef = signer.borrow<&MyToken.Vault>(from: /storage/mytokenVault)
            ?? panic("Could not borrow reference to the owner's Vault!")

        // Withdraw tokens from the signer's stored vault
        self.sentVault <- vaultRef.withdraw(amount: amount)
    }

    execute {
        // Get the recipient's public account object
        let recipient = getAccount(to)

        // Get a reference to the recipient's Receiver
        let receiverRef = recipient.getCapability(/public/mytokenReceiver)!.borrow<&{FungibleToken.Receiver}>()
            ?? panic("Could not borrow receiver reference to the recipient's Vault")

        // Deposit the withdrawn tokens in the recipient's receiver
        receiverRef.deposit(from: <-self.sentVault)
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
		SetScript([]byte(transferCode)).
		SetGasLimit(100).
		SetProposalKey(acctAddress, acctKey.Index, acctKey.SequenceNumber).
		SetReferenceBlockID(referenceBlock.ID).
		SetPayer(acctAddress).
		AddAuthorizer(acctAddress)

	tx.AddArgument(cadence.UFix64(uint64(10000)))

	tx.AddArgument(cadence.NewAddress(flow.HexToAddress("MyToken.Config.SingerAddress")))
	if err := tx.SignEnvelope(acctAddress, acctKey.Index, signer); err != nil {
		panic(err)
	}

	if err := flowClient.SendTransaction(ctx, *tx); err != nil {
		panic(err)
	}

	MyToken.WaitForSeal(ctx, flowClient, tx.ID())
}
