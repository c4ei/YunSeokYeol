package pow

import (
	"github.com/c4ei/c4exd/domain/consensus/model/externalapi"
	"github.com/c4ei/c4exd/domain/consensus/utils/consensushashing"
	"github.com/c4ei/c4exd/domain/consensus/utils/hashes"
	"github.com/c4ei/c4exd/domain/consensus/utils/serialization"
	"github.com/c4ei/c4exd/util/difficulty"

	"math/big"

	"github.com/pkg/errors"
)

// State is an intermediate data structure with pre-computed values to speed up mining.
// 상태는 마이닝 속도를 높이기 위해 미리 계산된 값이 있는 중간 데이터 구조입니다.
type State struct {
	mat        matrix
	Timestamp  int64
	Nonce      uint64
	Target     big.Int
	prePowHash externalapi.DomainHash
}

// NewState creates a new state with pre-computed values to speed up mining
// It takes the target from the Bits field
// NewState는 마이닝 속도를 높이기 위해 미리 계산된 값으로 새 상태를 생성합니다.
// Bits 필드에서 대상을 가져옵니다.
func NewState(header externalapi.MutableBlockHeader) *State {
	target := difficulty.CompactToBig(header.Bits())
	// Zero out the time and nonce.
	timestamp, nonce := header.TimeInMilliseconds(), header.Nonce()
	header.SetTimeInMilliseconds(0)
	header.SetNonce(0)
	prePowHash := consensushashing.HeaderHash(header)
	header.SetTimeInMilliseconds(timestamp)
	header.SetNonce(nonce)

	return &State{
		Target:     *target,
		prePowHash: *prePowHash,
		mat:        *generateMatrix(prePowHash),
		Timestamp:  timestamp,
		Nonce:      nonce,
	}
}

// CalculateProofOfWorkValue hashes the internal header and returns its big.Int value
func (state *State) CalculateProofOfWorkValue() *big.Int {
	// PRE_POW_HASH || TIME || 32 zero byte padding || NONCE
	writer := hashes.NewPoWHashWriter()
	writer.InfallibleWrite(state.prePowHash.ByteSlice())
	err := serialization.WriteElement(writer, state.Timestamp)
	if err != nil {
		panic(errors.Wrap(err, "this should never happen. Hash digest should never return an error"))
	}
	zeroes := [32]byte{}
	writer.InfallibleWrite(zeroes[:])
	err = serialization.WriteElement(writer, state.Nonce)
	if err != nil {
		panic(errors.Wrap(err, "this should never happen. Hash digest should never return an error"))
	}
	powHash := writer.Finalize()
	heavyHash := state.mat.HeavyHash(powHash)
	return toBig(heavyHash)
}

// IncrementNonce the nonce in State by 1
func (state *State) IncrementNonce() {
	state.Nonce++
}

// CheckProofOfWork check's if the block has a valid PoW according to the provided target
// it does not check if the difficulty itself is valid or less than the maximum for the appropriate network
// 제공된 대상에 따라 블록에 유효한 PoW가 있는지 확인하려면 ProofOfWork를 확인하세요.
// 난이도 자체가 유효한지, 해당 네트워크의 최대값보다 작은지는 확인하지 않습니다.
func (state *State) CheckProofOfWork() bool {
	// The block pow must be less than the claimed target
	// 블록 전력은 청구된 목표보다 작아야 합니다.
	powNum := state.CalculateProofOfWorkValue()

	// The block hash must be less or equal than the claimed target.
	// 블록 해시는 청구된 대상보다 작거나 같아야 합니다.
	return powNum.Cmp(&state.Target) <= 0
}

// CheckProofOfWorkByBits check's if the block has a valid PoW according to its Bits field
// it does not check if the difficulty itself is valid or less than the maximum for the appropriate network
// ProofOfWorkByBits를 확인하여 해당 블록에 Bits 필드에 따라 유효한 PoW가 있는지 확인합니다.
// 난이도 자체가 유효한지, 해당 네트워크의 최대값보다 작은지는 확인하지 않습니다.
func CheckProofOfWorkByBits(header externalapi.MutableBlockHeader) bool {
	return NewState(header).CheckProofOfWork()
}

// ToBig converts a externalapi.DomainHash into a big.Int treated as a little endian string.
// ToBig은 externalapi.DomainHash를 리틀 엔디안 문자열로 처리되는 big.Int로 변환합니다.
func toBig(hash *externalapi.DomainHash) *big.Int {
	// We treat the Hash as little-endian for PoW purposes, but the big package wants the bytes in big-endian, so reverse them.
	// ToBig은 externalapi.DomainHash를 리틀 엔디안 문자열로 처리되는 big.Int로 변환합니다.
	buf := hash.ByteSlice()
	blen := len(buf)
	for i := 0; i < blen/2; i++ {
		buf[i], buf[blen-1-i] = buf[blen-1-i], buf[i]
	}

	return new(big.Int).SetBytes(buf)
}

// BlockLevel returns the block level of the given header.
// BlockLevel은 주어진 헤더의 블록 수준을 반환합니다.
func BlockLevel(header externalapi.BlockHeader, maxBlockLevel int) int {
	// Genesis is defined to be the root of all blocks at all levels, so we define it to be the maximal block level.
	// 제네시스는 모든 레벨의 모든 블록의 루트로 정의되므로 최대 블록 레벨로 정의합니다.
	if len(header.DirectParents()) == 0 {
		return maxBlockLevel
	}

	proofOfWorkValue := NewState(header.ToMutable()).CalculateProofOfWorkValue()
	level := maxBlockLevel - proofOfWorkValue.BitLen()
	// If the block has a level lower than genesis make it zero.
	if level < 0 {
		level = 0
	}
	return level
}
