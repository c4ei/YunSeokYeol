package transactionid

import (
	"github.com/c4ei/c4exd/domain/consensus/model/externalapi"
)

// FromBytes creates a DomainTransactionID from the given byte slice
func FromBytes(transactionIDBytes []byte) (*externalapi.DomainTransactionID, error) {
	hash, err := externalapi.NewDomainHashFromByteSlice(transactionIDBytes)
	if err != nil {
		return nil, err
	}
	return (*externalapi.DomainTransactionID)(hash), nil
}
