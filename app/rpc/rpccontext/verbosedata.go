package rpccontext

import (
	"encoding/hex"
	"math"
	"math/big"

	difficultyPackage "github.com/c4ei/c4exd/util/difficulty"
	"github.com/pkg/errors"

	"github.com/c4ei/c4exd/domain/consensus/utils/hashes"

	"github.com/c4ei/c4exd/domain/consensus/utils/txscript"

	"github.com/c4ei/c4exd/app/appmessage"
	"github.com/c4ei/c4exd/domain/consensus/model/externalapi"
	"github.com/c4ei/c4exd/domain/consensus/utils/consensushashing"
	"github.com/c4ei/c4exd/domain/dagconfig"
)

// ErrBuildBlockVerboseDataInvalidBlock indicates that a block that was given to BuildBlockVerboseData is invalid.
// ErrBuildBlockVerboseDataInvalidBlock은 BuildBlockVerboseData에 제공된 블록이 유효하지 않음을 나타냅니다.
var ErrBuildBlockVerboseDataInvalidBlock = errors.New("ErrBuildBlockVerboseDataInvalidBlock")

// GetDifficultyRatio returns the proof-of-work difficulty as a multiple of the
// minimum difficulty using the passed bits field from the header of a block.
// GetDifficultyRatio는 작업 증명 난이도를 다음의 배수로 반환합니다.
// 블록 헤더에서 전달된 비트 필드를 사용하는 최소 난이도.
func (ctx *Context) GetDifficultyRatio(bits uint32, params *dagconfig.Params) float64 {
	// The minimum difficulty is the max possible proof-of-work limit bits
	// converted back to a number. Note this is not the same as the proof of
	// work limit directly because the block difficulty is encoded in a block
	// with the compact form which loses precision.
	// 최소 난이도는 가능한 최대 작업 증명 제한 비트입니다.
	// 다시 숫자로 변환됩니다. 참고로 이는 증명과 동일하지 않습니다.
	// 블록 난이도가 블록에 인코딩되어 있기 때문에 작업 제한이 직접적으로 적용됩니다.
	// 정밀도가 떨어지는 컴팩트한 형태입니다.
	target := difficultyPackage.CompactToBig(bits)

	difficulty := new(big.Rat).SetFrac(params.PowMax, target)
	diff, _ := difficulty.Float64()

	roundingPrecision := float64(100)
	diff = math.Round(diff*roundingPrecision) / roundingPrecision

	return diff
}

// PopulateBlockWithVerboseData populates the given `block` with verbose
// data from `domainBlockHeader` and optionally from `domainBlock`
// PopulateBlockWithVerboseData는 주어진 `블록`을 자세한 정보로 채웁니다.
// `domainBlockHeader`의 데이터 및 선택적으로 `domainBlock`의 데이터
func (ctx *Context) PopulateBlockWithVerboseData(block *appmessage.RPCBlock, domainBlockHeader externalapi.BlockHeader,
	domainBlock *externalapi.DomainBlock, includeTransactionVerboseData bool) error {

	blockHash := consensushashing.HeaderHash(domainBlockHeader)

	blockInfo, err := ctx.Domain.Consensus().GetBlockInfo(blockHash)
	if err != nil {
		return err
	}

	if blockInfo.BlockStatus == externalapi.StatusInvalid {
		return errors.Wrap(ErrBuildBlockVerboseDataInvalidBlock, "cannot build verbose data for "+
			"invalid block")
	}

	_, childrenHashes, err := ctx.Domain.Consensus().GetBlockRelations(blockHash)
	if err != nil {
		return err
	}

	isChainBlock, err := ctx.Domain.Consensus().IsChainBlock(blockHash)
	if err != nil {
		return err
	}

	block.VerboseData = &appmessage.RPCBlockVerboseData{
		Hash:                blockHash.String(),
		Difficulty:          ctx.GetDifficultyRatio(domainBlockHeader.Bits(), ctx.Config.ActiveNetParams),
		ChildrenHashes:      hashes.ToStrings(childrenHashes),
		IsHeaderOnly:        blockInfo.BlockStatus == externalapi.StatusHeaderOnly,
		BlueScore:           blockInfo.BlueScore,
		MergeSetBluesHashes: hashes.ToStrings(blockInfo.MergeSetBlues),
		MergeSetRedsHashes:  hashes.ToStrings(blockInfo.MergeSetReds),
		IsChainBlock:        isChainBlock,
	}
	// selectedParentHash will be nil in the genesis block
	// selectedParentHash는 제네시스 블록에서 nil이 됩니다.
	if blockInfo.SelectedParent != nil {
		block.VerboseData.SelectedParentHash = blockInfo.SelectedParent.String()
	}

	if blockInfo.BlockStatus == externalapi.StatusHeaderOnly {
		return nil
	}

	// Get the block if we didn't receive it previously
	// 이전에 블록을 받지 못한 경우 블록을 가져옵니다.
	if domainBlock == nil {
		domainBlock, err = ctx.Domain.Consensus().GetBlockEvenIfHeaderOnly(blockHash)
		if err != nil {
			return err
		}
	}

	transactionIDs := make([]string, len(domainBlock.Transactions))
	for i, transaction := range domainBlock.Transactions {
		transactionIDs[i] = consensushashing.TransactionID(transaction).String()
	}
	block.VerboseData.TransactionIDs = transactionIDs

	if includeTransactionVerboseData {
		for _, transaction := range block.Transactions {
			err := ctx.PopulateTransactionWithVerboseData(transaction, domainBlockHeader)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// PopulateTransactionWithVerboseData populates the given `transaction` with
// verbose data from `domainTransaction`
// PopulateTransactionWithVerboseData는 주어진 `트랜잭션`을 다음으로 채웁니다.
// `domainTransaction`의 자세한 데이터
func (ctx *Context) PopulateTransactionWithVerboseData(
	transaction *appmessage.RPCTransaction, domainBlockHeader externalapi.BlockHeader) error {

	domainTransaction, err := appmessage.RPCTransactionToDomainTransaction(transaction)
	if err != nil {
		return err
	}

	ctx.Domain.Consensus().PopulateMass(domainTransaction)

	transaction.VerboseData = &appmessage.RPCTransactionVerboseData{
		TransactionID: consensushashing.TransactionID(domainTransaction).String(),
		Hash:          consensushashing.TransactionHash(domainTransaction).String(),
		Mass:          domainTransaction.Mass,
	}
	if domainBlockHeader != nil {
		transaction.VerboseData.BlockHash = consensushashing.HeaderHash(domainBlockHeader).String()
		transaction.VerboseData.BlockTime = uint64(domainBlockHeader.TimeInMilliseconds())
	}
	for _, input := range transaction.Inputs {
		ctx.populateTransactionInputWithVerboseData(input)
	}
	for _, output := range transaction.Outputs {
		err := ctx.populateTransactionOutputWithVerboseData(output)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ctx *Context) populateTransactionInputWithVerboseData(transactionInput *appmessage.RPCTransactionInput) {
	transactionInput.VerboseData = &appmessage.RPCTransactionInputVerboseData{}
}

func (ctx *Context) populateTransactionOutputWithVerboseData(transactionOutput *appmessage.RPCTransactionOutput) error {
	scriptPublicKey, err := hex.DecodeString(transactionOutput.ScriptPublicKey.Script)
	if err != nil {
		return err
	}
	domainScriptPublicKey := &externalapi.ScriptPublicKey{
		Script:  scriptPublicKey,
		Version: transactionOutput.ScriptPublicKey.Version,
	}

	// Ignore the error here since an error means the script
	// couldn't be parsed and there's no additional information about
	// it anyways
	// 오류는 스크립트를 의미하므로 여기서는 오류를 무시합니다.
	// 구문 분석할 수 없으며 이에 대한 추가 정보가 없습니다.
	// 어쨌든 그렇죠
	scriptPublicKeyType, scriptPublicKeyAddress, _ := txscript.ExtractScriptPubKeyAddress(
		domainScriptPublicKey, ctx.Config.ActiveNetParams)

	var encodedScriptPublicKeyAddress string
	if scriptPublicKeyAddress != nil {
		encodedScriptPublicKeyAddress = scriptPublicKeyAddress.EncodeAddress()
	}
	transactionOutput.VerboseData = &appmessage.RPCTransactionOutputVerboseData{
		ScriptPublicKeyType:    scriptPublicKeyType.String(),
		ScriptPublicKeyAddress: encodedScriptPublicKeyAddress,
	}
	return nil
}
