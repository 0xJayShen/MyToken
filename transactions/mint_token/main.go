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

var mintCode = fmt.Sprintf(`
import MyToken from %s
import FungibleToken from %s

transaction(amount: UFix64, to: Address) {

    let tokenMinter: &MyToken.MinterProxy
    let tokenReceiver: &{FungibleToken.Receiver}

    prepare(minterAccount: AuthAccount) {
        self.tokenMinter = minterAccount
            .borrow<&MyToken.MinterProxy>(from: MyToken.MinterProxyStoragePath)
            ?? panic("No minter available")

        self.tokenReceiver = getAccount(to)
            .getCapability(/public/mytokenReceiver)!
            .borrow<&{FungibleToken.Receiver}>()
            ?? panic("Unable to borrow receiver reference")
    }

    execute {
        let mintedVault <- self.tokenMinter.mintTokens(amount: amount)
        self.tokenReceiver.deposit(from: <-mintedVault)
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
		SetScript([]byte(mintCode)).
		SetGasLimit(100).
		SetProposalKey(acctAddress, acctKey.Index, acctKey.SequenceNumber).
		SetReferenceBlockID(referenceBlock.ID).
		SetPayer(acctAddress).
		AddAuthorizer(acctAddress)
	tx.AddArgument(cadence.UFix64(uint64(10000)))
	tx.AddArgument(cadence.NewAddress(flow.HexToAddress(MyToken.Config.SingerAddress)))

	if err := tx.SignEnvelope(acctAddress, acctKey.Index, signer); err != nil {
		panic(err)
	}

	if err := flowClient.SendTransaction(ctx, *tx); err != nil {
		panic(err)
	}

	MyToken.WaitForSeal(ctx, flowClient, tx.ID())
}
