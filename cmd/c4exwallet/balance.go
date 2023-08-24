package main

import (
	"context"
	"fmt"

	"github.com/c4ei/c4exd/cmd/c4exwallet/daemon/client"
	"github.com/c4ei/c4exd/cmd/c4exwallet/daemon/pb"
	"github.com/c4ei/c4exd/cmd/c4exwallet/utils"
)

func balance(conf *balanceConfig) error {
	daemonClient, tearDown, err := client.Connect(conf.DaemonAddress)
	if err != nil {
		return err
	}
	defer tearDown()

	ctx, cancel := context.WithTimeout(context.Background(), daemonTimeout)
	defer cancel()
	response, err := daemonClient.GetBalance(ctx, &pb.GetBalanceRequest{})
	if err != nil {
		return err
	}

	pendingSuffix := ""
	if response.Pending > 0 {
		pendingSuffix = " (pending)"
	}
	if conf.Verbose {
		pendingSuffix = ""
		println("Address                                                                       Available             Pending")
		println("-----------------------------------------------------------------------------------------------------------")
		for _, addressBalance := range response.AddressBalances {
			fmt.Printf("%s %s %s\n", addressBalance.Address, utils.FormatC4x(addressBalance.Available), utils.FormatC4x(addressBalance.Pending))
		}
		println("-----------------------------------------------------------------------------------------------------------")
		print("                                                 ")
	}
	fmt.Printf("Total balance, C4X %s %s%s\n", utils.FormatC4x(response.Available), utils.FormatC4x(response.Pending), pendingSuffix)

	return nil
}
