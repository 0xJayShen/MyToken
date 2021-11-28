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

var depositMinterCode  = fmt.Sprintf(`
import MyToken from %s

transaction(minterAddress: Address) {

    let resourceStoragePath: StoragePath
    let capabilityPrivatePath: CapabilityPath
    let minterCapability: Capability<&MyToken.Minter>

    prepare(adminAccount: AuthAccount) {

        // These paths must be unique within the MyToken contract account's storage
        self.resourceStoragePath = /storage/minter_01    // e.g. /storage/minter_01
        self.capabilityPrivatePath = /private/minter_01 // e.g. private/minter_01

        // Create a reference to the admin resource in storage.
        let tokenAdmin = adminAccount.borrow<&MyToken.Administrator>(from: MyToken.AdminStoragePath)
            ?? panic("Could not borrow a reference to the admin resource")

        // Create a new minter resource and a private link to a capability for it in the admin's storage.
        let minter <- tokenAdmin.createNewMinter()
        adminAccount.save(<- minter, to: self.resourceStoragePath)
        self.minterCapability = adminAccount.link<&MyToken.Minter>(
            self.capabilityPrivatePath,
            target: self.resourceStoragePath
        ) ?? panic("Could not link minter")

    }

    execute {
        // This is the account that the capability will be given to
        let minterAccount = getAccount(minterAddress)

        let capabilityReceiver = minterAccount.getCapability
            <&MyToken.MinterProxy{MyToken.MinterProxyPublic}>
            (MyToken.MinterProxyPublicPath)!
            .borrow() ?? panic("Could not borrow capability receiver reference")

        capabilityReceiver.setMinterCapability(cap: self.minterCapability)
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
		SetScript([]byte(depositMinterCode)).
		SetGasLimit(100).
		SetProposalKey(acctAddress, acctKey.Index, acctKey.SequenceNumber).
		SetReferenceBlockID(referenceBlock.ID).
		SetPayer(acctAddress).
		AddAuthorizer(acctAddress)
	tx.AddArgument(cadence.NewAddress(flow.HexToAddress(MyToken.Config.SingerAddress)))
	if err := tx.SignEnvelope(acctAddress, acctKey.Index, signer); err != nil {
		panic(err)
	}

	if err := flowClient.SendTransaction(ctx, *tx); err != nil {
		panic(err)
	}

	MyToken.WaitForSeal(ctx, flowClient, tx.ID())
}
