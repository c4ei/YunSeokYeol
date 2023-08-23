package testapi

import (
	"github.com/c4ei/c4exd/domain/consensus/model"
	"github.com/c4ei/c4exd/domain/consensus/utils/txscript"
)

// TestTransactionValidator adds to the main TransactionValidator methods required by tests
type TestTransactionValidator interface {
	model.TransactionValidator
	SigCache() *txscript.SigCache
	SetSigCache(sigCache *txscript.SigCache)
}
