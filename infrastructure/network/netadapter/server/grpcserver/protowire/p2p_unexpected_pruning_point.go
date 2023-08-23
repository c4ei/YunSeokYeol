package protowire

import "github.com/c4ei/c4exd/app/appmessage"

func (x *C4exdMessage_UnexpectedPruningPoint) toAppMessage() (appmessage.Message, error) {
	return &appmessage.MsgUnexpectedPruningPoint{}, nil
}

func (x *C4exdMessage_UnexpectedPruningPoint) fromAppMessage(_ *appmessage.MsgUnexpectedPruningPoint) error {
	return nil
}
