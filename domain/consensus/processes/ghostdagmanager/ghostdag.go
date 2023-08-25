package ghostdagmanager

import (
	"fmt"
	"math/big"

	"github.com/c4ei/c4exd/domain/consensus/model"
	"github.com/c4ei/c4exd/domain/consensus/model/externalapi"
	"github.com/c4ei/c4exd/util/difficulty"
	"github.com/pkg/errors"
)

type blockGHOSTDAGData struct {
	blueScore          uint64
	blueWork           *big.Int
	selectedParent     *externalapi.DomainHash
	mergeSetBlues      []*externalapi.DomainHash
	mergeSetReds       []*externalapi.DomainHash
	bluesAnticoneSizes map[externalapi.DomainHash]externalapi.KType
}

func (bg *blockGHOSTDAGData) toModel() *externalapi.BlockGHOSTDAGData {
	fmt.Printf("@@@@ 23 line ghostdag.go toModel() bg.blueScore:%v, bg.blueWork:%v, bg.selectedParent:%v, bg.mergeSetBlues:%v, bg.mergeSetReds:%v, bg.bluesAnticoneSizes:%v\n", bg.blueScore, bg.blueWork, bg.selectedParent, bg.mergeSetBlues, bg.mergeSetReds, bg.bluesAnticoneSizes)
	return externalapi.NewBlockGHOSTDAGData(bg.blueScore, bg.blueWork, bg.selectedParent, bg.mergeSetBlues, bg.mergeSetReds, bg.bluesAnticoneSizes)
}

// GHOSTDAG runs the GHOSTDAG protocol and calculates the block BlockGHOSTDAGData by the given parents.
// The function calculates MergeSetBlues by iterating over the blocks in
// the anticone of the new block selected parent (which is the parent with the
// highest blue score) and adds any block to newNode.blues if by adding
// it to MergeSetBlues these conditions will not be violated:
//
// 1) |anticone-of-candidate-block ∩ blue-set-of-newBlock| ≤ K
//
//  2. For every blue block in blue-set-of-newBlock:
//     |(anticone-of-blue-block ∩ blue-set-newBlock) ∪ {candidate-block}| ≤ K.
//     We validate this condition by maintaining a map BluesAnticoneSizes for
//     each block which holds all the blue anticone sizes that were affected by
//     the new added blue blocks.
//     So to find out what is |anticone-of-blue ∩ blue-set-of-newBlock| we just iterate in
//     the selected parent chain of the new block until we find an existing entry in
//     BluesAnticoneSizes.
//
// For further details see the article https://eprint.iacr.org/2018/104.pdf
// GHOSTDAG는 GHOSTDAG 프로토콜을 실행하고 주어진 부모에 의해 블록 BlockGHOSTDAGData를 계산합니다.
// 함수는 다음의 블록을 반복하여 MergeSetBlues를 계산합니다.
// 선택된 새로운 블록의 안티콘
// 가장 높은 파란색 점수) 추가하는 경우 newNode.blues에 블록을 추가합니다.
// MergeSetBlues에는 다음 조건이 위반되지 않습니다.
//
// 1) |후보 블록 안티콘 ∩ newBlock 블루 세트| ≤ K
//
// 2. blue-set-of-newBlock의 모든 파란색 블록에 대해:
// |(파란색 블록의 안티콘 ∩ 파란색 세트-newBlock) ∪ {후보 블록}| ≤ K.
// BluesAnticoneSizes 맵을 유지하여 이 조건을 검증합니다.
// 영향을 받은 모든 파란색 안티콘 크기를 보유하는 각 블록
// 새로 추가된 파란색 블록입니다.
// 따라서 |anticone-of-blue ∩ blue-set-of-newBlock|이 무엇인지 알아보려면 우리는 단지 반복
// 기존 항목을 찾을 때까지 새 블록의 선택된 상위 체인
// BluesAnticoneSizes.
//
// 자세한 내용은 https://eprint.iacr.org/2018/104.pdf 기사를 참조하세요.
func (gm *ghostdagManager) GHOSTDAG(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) error {
	newBlockData := &blockGHOSTDAGData{
		blueWork:           new(big.Int),
		mergeSetBlues:      make([]*externalapi.DomainHash, 0),
		mergeSetReds:       make([]*externalapi.DomainHash, 0),
		bluesAnticoneSizes: make(map[externalapi.DomainHash]externalapi.KType),
	}

	blockParents, err := gm.dagTopologyManager.Parents(stagingArea, blockHash)
	if err != nil {
		return err
	}

	isGenesis := len(blockParents) == 0
	if !isGenesis {
		selectedParent, err := gm.findSelectedParent(stagingArea, blockParents)
		if err != nil {
			return err
		}

		newBlockData.selectedParent = selectedParent
		newBlockData.mergeSetBlues = append(newBlockData.mergeSetBlues, selectedParent)
		newBlockData.bluesAnticoneSizes[*selectedParent] = 0
	}

	mergeSetWithoutSelectedParent, err := gm.mergeSetWithoutSelectedParent(
		stagingArea, newBlockData.selectedParent, blockParents)
	if err != nil {
		return err
	}

	for _, blueCandidate := range mergeSetWithoutSelectedParent {
		isBlue, candidateAnticoneSize, candidateBluesAnticoneSizes, err := gm.checkBlueCandidate(
			stagingArea, newBlockData.toModel(), blueCandidate)
		if err != nil {
			return err
		}

		if isBlue {
			// No k-cluster violation found, we can now set the candidate block as blue
			newBlockData.mergeSetBlues = append(newBlockData.mergeSetBlues, blueCandidate)
			newBlockData.bluesAnticoneSizes[*blueCandidate] = candidateAnticoneSize
			for blue, blueAnticoneSize := range candidateBluesAnticoneSizes {
				newBlockData.bluesAnticoneSizes[blue] = blueAnticoneSize + 1
			}
		} else {
			newBlockData.mergeSetReds = append(newBlockData.mergeSetReds, blueCandidate)
		}
	}

	if !isGenesis {
		selectedParentGHOSTDAGData, err := gm.ghostdagDataStore.Get(gm.databaseContext, stagingArea, newBlockData.selectedParent, false)
		if err != nil {
			return err
		}
		newBlockData.blueScore = selectedParentGHOSTDAGData.BlueScore() + uint64(len(newBlockData.mergeSetBlues))
		// We inherit the bluework from the selected parent
		newBlockData.blueWork.Set(selectedParentGHOSTDAGData.BlueWork())
		// Then we add up all the *work*(not blueWork) that all of newBlock merge set blues and selected parent did
		for _, blue := range newBlockData.mergeSetBlues {
			// We don't count the work of the virtual genesis
			if blue.Equal(model.VirtualGenesisBlockHash) {
				continue
			}

			header, err := gm.headerStore.BlockHeader(gm.databaseContext, stagingArea, blue)
			if err != nil {
				return err
			}
			newBlockData.blueWork.Add(newBlockData.blueWork, difficulty.CalcWork(header.Bits()))
		}
	} else {
		// Genesis's blue score is defined to be 0.
		newBlockData.blueScore = 0
		newBlockData.blueWork.SetUint64(0)
	}

	gm.ghostdagDataStore.Stage(stagingArea, blockHash, newBlockData.toModel(), false)

	return nil
}

type chainBlockData struct {
	hash      *externalapi.DomainHash
	blockData *externalapi.BlockGHOSTDAGData
}

func (gm *ghostdagManager) checkBlueCandidate(stagingArea *model.StagingArea, newBlockData *externalapi.BlockGHOSTDAGData,
	blueCandidate *externalapi.DomainHash) (isBlue bool, candidateAnticoneSize externalapi.KType,
	candidateBluesAnticoneSizes map[externalapi.DomainHash]externalapi.KType, err error) {

	// The maximum length of node.blues can be K+1 because
	// it contains the selected parent.
	if externalapi.KType(len(newBlockData.MergeSetBlues())) == gm.k+1 {
		return false, 0, nil, nil
	}

	candidateBluesAnticoneSizes = make(map[externalapi.DomainHash]externalapi.KType, gm.k)

	// Iterate over all blocks in the blue set of newNode that are not in the past
	// of blueCandidate, and check for each one of them if blueCandidate potentially
	// enlarges their blue anticone to be over K, or that they enlarge the blue anticone
	// of blueCandidate to be over K.
	// 과거에 있지 않은 파란색 newNode 세트의 모든 블록을 반복합니다.
	// blueCandidate의 경우 blueCandidate가 잠재적으로 있는지 각각 확인합니다.
	// 파란색 안티콘을 K보다 크게 확대하거나 파란색 안티콘을 확대합니다.
	// blueCandidate의 K가 초과됩니다.
	chainBlock := chainBlockData{
		blockData: newBlockData,
	}

	for {
		// 체인 블록을 사용하는 블루 후보 확인
		isBlue, isRed, err := gm.checkBlueCandidateWithChainBlock(stagingArea, newBlockData, chainBlock, blueCandidate,
			candidateBluesAnticoneSizes, &candidateAnticoneSize)
		if err != nil {
			return false, 0, nil, err
		}

		if isBlue {
			break
		}

		if isRed {
			return false, 0, nil, nil
		}

		selectedParentGHOSTDAGData, err := gm.ghostdagDataStore.Get(gm.databaseContext, stagingArea, chainBlock.blockData.SelectedParent(), false)
		if err != nil {
			return false, 0, nil, err
		}

		chainBlock = chainBlockData{hash: chainBlock.blockData.SelectedParent(),
			blockData: selectedParentGHOSTDAGData,
		}
	}

	return true, candidateAnticoneSize, candidateBluesAnticoneSizes, nil
}

func (gm *ghostdagManager) checkBlueCandidateWithChainBlock(stagingArea *model.StagingArea,
	newBlockData *externalapi.BlockGHOSTDAGData, chainBlock chainBlockData, blueCandidate *externalapi.DomainHash,
	candidateBluesAnticoneSizes map[externalapi.DomainHash]externalapi.KType,
	candidateAnticoneSize *externalapi.KType) (isBlue, isRed bool, err error) {

	// If blueCandidate is in the future of chainBlock, it means
	// that all remaining blues are in the past of chainBlock and thus
	// in the past of blueCandidate. In this case we know for sure that
	// the anticone of blueCandidate will not exceed K, and we can mark
	// it as blue.
	//
	// The new block is always in the future of blueCandidate, so there's
	// no point in checking it.

	// We check if chainBlock is not the new block by checking if it has a hash.
	// blueCandidate가 chainBlock의 미래에 있다면 이는 다음을 의미합니다.
	// 나머지 모든 파란색은 chainBlock의 과거에 있으므로
	// blueCandidate의 과거. 이 경우 우리는 확실히 알고 있습니다
	// blueCandidate의 안티콘은 K를 초과하지 않으며 표시할 수 있습니다.
	// 파란색으로 표시됩니다.
	//
	// 새 블록은 항상 blueCandidate의 미래에 있으므로
	// 확인해봐도 의미가 없습니다.

	// 해시가 있는지 확인하여 chainBlock이 새 블록이 아닌지 확인합니다.
	if chainBlock.hash != nil {
		isAncestorOfBlueCandidate, err := gm.dagTopologyManager.IsAncestorOf(stagingArea, chainBlock.hash, blueCandidate)
		if err != nil {
			return false, false, err
		}
		if isAncestorOfBlueCandidate {
			return true, false, nil
		}
	}

	for _, block := range chainBlock.blockData.MergeSetBlues() {
		// Skip blocks that exist in the past of blueCandidate.
		isAncestorOfBlueCandidate, err := gm.dagTopologyManager.IsAncestorOf(stagingArea, block, blueCandidate)
		if err != nil {
			return false, false, err
		}

		if isAncestorOfBlueCandidate {
			continue
		}

		candidateBluesAnticoneSizes[*block], err = gm.blueAnticoneSize(stagingArea, block, newBlockData)
		if err != nil {
			return false, false, err
		}
		*candidateAnticoneSize++

		if *candidateAnticoneSize > gm.k {
			// k-cluster violation: The candidate's blue anticone exceeded k
			return false, true, nil
		}

		if candidateBluesAnticoneSizes[*block] == gm.k {
			// k-cluster violation: A block in candidate's blue anticone already
			// has k blue blocks in its own anticone
			return false, true, nil
		}

		// This is a sanity check that validates that a blue
		// block's blue anticone is not already larger than K.
		// 이것은 파란색이 올바른지 확인하는 온전성 검사입니다.
		// 블록의 파란색 안티콘은 아직 K보다 크지 않습니다.
		if candidateBluesAnticoneSizes[*block] > gm.k {
			return false, false, errors.New("found blue anticone size larger than k")
		}
	}

	return false, false, nil
}

// blueAnticoneSize returns the blue anticone size of 'block' from the worldview of 'context'.
// Expects 'block' to be in the blue set of 'context'
// blueAnticoneSize는 'context'의 세계관에서 'block'의 파란색 안티콘 크기를 반환합니다.
// '블록'이 파란색 '컨텍스트' 세트에 있을 것으로 예상합니다.
func (gm *ghostdagManager) blueAnticoneSize(stagingArea *model.StagingArea,
	block *externalapi.DomainHash, context *externalapi.BlockGHOSTDAGData) (externalapi.KType, error) {

	isTrustedData := false
	for current := context; current != nil; {
		if blueAnticoneSize, ok := current.BluesAnticoneSizes()[*block]; ok {
			return blueAnticoneSize, nil
		}
		if current.SelectedParent().Equal(gm.genesisHash) {
			break
		}

		var err error
		current, err = gm.ghostdagDataStore.Get(gm.databaseContext, stagingArea, current.SelectedParent(), isTrustedData)
		if err != nil {
			return 0, err
		}
		if current.SelectedParent().Equal(model.VirtualGenesisBlockHash) {
			isTrustedData = true
			current, err = gm.ghostdagDataStore.Get(gm.databaseContext, stagingArea, current.SelectedParent(), isTrustedData)
			if err != nil {
				return 0, err
			}
		}
	}
	return 0, errors.Errorf("block %s is not in blue set of the given context", block)
}
