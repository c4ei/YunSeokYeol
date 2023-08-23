package protowire

import (
	"github.com/c4ei/c4exd/app/appmessage"
	"github.com/pkg/errors"
)

func (x *C4exdMessage_Verack) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "C4exdMessage_Verack is nil")
	}
	return &appmessage.MsgVerAck{}, nil
}

func (x *C4exdMessage_Verack) fromAppMessage(_ *appmessage.MsgVerAck) error {
	return nil
}
