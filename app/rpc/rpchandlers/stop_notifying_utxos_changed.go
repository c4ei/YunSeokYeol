package rpchandlers

import (
	"github.com/c4ei/c4exd/app/appmessage"
	"github.com/c4ei/c4exd/app/rpc/rpccontext"
	"github.com/c4ei/c4exd/infrastructure/network/netadapter/router"
)

// HandleStopNotifyingUTXOsChanged handles the respectively named RPC command
func HandleStopNotifyingUTXOsChanged(context *rpccontext.Context, router *router.Router, request appmessage.Message) (appmessage.Message, error) {
	if !context.Config.UTXOIndex {
		errorMessage := appmessage.NewStopNotifyingUTXOsChangedResponseMessage()
		errorMessage.Error = appmessage.RPCErrorf("Method unavailable when c4exd is run without --utxoindex")
		return errorMessage, nil
	}

	stopNotifyingUTXOsChangedRequest := request.(*appmessage.StopNotifyingUTXOsChangedRequestMessage)
	addresses, err := context.ConvertAddressStringsToUTXOsChangedNotificationAddresses(stopNotifyingUTXOsChangedRequest.Addresses)
	if err != nil {
		errorMessage := appmessage.NewNotifyUTXOsChangedResponseMessage()
		errorMessage.Error = appmessage.RPCErrorf("Parsing error: %s", err)
		return errorMessage, nil
	}

	listener, err := context.NotificationManager.Listener(router)
	if err != nil {
		return nil, err
	}
	context.NotificationManager.StopPropagatingUTXOsChangedNotifications(listener, addresses)

	response := appmessage.NewStopNotifyingUTXOsChangedResponseMessage()
	return response, nil
}
