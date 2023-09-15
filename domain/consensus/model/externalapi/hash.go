package externalapi

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/pkg/errors"
)

// DomainHashSize of array used to store hashes.
// 해시를 저장하는 데 사용되는 배열의 DomainHashSize입니다.
const DomainHashSize = 32

// DomainHash is the domain representation of a Hash
// NewDomainHashFromByteArray는 새로운 DomainHash를 구성하는 바이트 배열입니다.
type DomainHash struct {
	hashArray [DomainHashSize]byte
}

// NewZeroHash returns a DomainHash that represents the zero value (0x000000...000)
// NewZeroHash는 0 값(0x000000...000)을 나타내는 DomainHash를 반환합니다.
func NewZeroHash() *DomainHash {
	return &DomainHash{hashArray: [32]byte{}}
}

// NewDomainHashFromByteArray constructs a new DomainHash out of a byte array
// NewDomainHashFromByteArray는 바이트 배열에서 새로운 DomainHash를 구성합니다.
func NewDomainHashFromByteArray(hashBytes *[DomainHashSize]byte) *DomainHash {
	return &DomainHash{
		hashArray: *hashBytes,
	}
}

// NewDomainHashFromByteSlice constructs a new DomainHash out of a byte slice.
// Returns an error if the length of the byte slice is not exactly `DomainHashSize`
// NewDomainHashFromByteSlice는 바이트 슬라이스에서 새로운 DomainHash를 구성합니다.
// 바이트 슬라이스의 길이가 정확히 `DomainHashSize`가 아닌 경우 오류를 반환합니다.
func NewDomainHashFromByteSlice(hashBytes []byte) (*DomainHash, error) {
	if len(hashBytes) != DomainHashSize {
		return nil, errors.Errorf("invalid hash size. Want: %d, got: %d",
			DomainHashSize, len(hashBytes))
	}
	domainHash := DomainHash{
		hashArray: [DomainHashSize]byte{},
	}
	copy(domainHash.hashArray[:], hashBytes)
	return &domainHash, nil
}

// NewDomainHashFromString constructs a new DomainHash out of a hex-encoded string.
// Returns an error if the length of the string is not exactly `DomainHashSize * 2`
// NewDomainHashFromString은 16진수로 인코딩된 문자열에서 새로운 DomainHash를 구성합니다.
// 문자열의 길이가 정확히 `DomainHashSize * 2`가 아닌 경우 오류를 반환합니다.
func NewDomainHashFromString(hashString string) (*DomainHash, error) {
	expectedLength := DomainHashSize * 2
	// Return error if hash string is too long.
	if len(hashString) != expectedLength {
		return nil, errors.Errorf("hash string length is %d, while it should be be %d",
			len(hashString), expectedLength)
	}

	hashBytes, err := hex.DecodeString(hashString)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	fmt.Printf("##### 66 /externalapi/hash.go NewDomainHashFromString() hashString:%+v\n", hashString)
	return NewDomainHashFromByteSlice(hashBytes)
}

// String returns the Hash as the hexadecimal string of the hash.
func (hash DomainHash) String() string {
	return hex.EncodeToString(hash.hashArray[:])
}

// ByteArray returns the bytes in this hash represented as a byte array.
// The hash bytes are cloned, therefore it is safe to modify the resulting array.
// ByteArray는 바이트 배열로 표시된 이 해시의 바이트를 반환합니다.
// 해시 바이트가 복제되었으므로 결과 배열을 수정해도 안전합니다.
func (hash *DomainHash) ByteArray() *[DomainHashSize]byte {
	arrayClone := hash.hashArray
	return &arrayClone
}

// ByteSlice returns the bytes in this hash represented as a byte slice.
// The hash bytes are cloned, therefore it is safe to modify the resulting slice.
// ByteSlice는 바이트 슬라이스로 표시된 이 해시의 바이트를 반환합니다.
// 해시 바이트가 복제되었으므로 결과 슬라이스를 수정해도 안전합니다.
func (hash *DomainHash) ByteSlice() []byte {
	return hash.ByteArray()[:]
}

// If this doesn't compile, it means the type definition has been changed, so it's
// an indication to update Equal and Clone accordingly.
// 컴파일되지 않으면 유형 정의가 변경된 것이므로
// 이에 따라 Equal 및 Clone을 업데이트하라는 표시입니다.
var _ DomainHash = DomainHash{hashArray: [DomainHashSize]byte{}}

// Equal returns whether hash equals to other
func (hash *DomainHash) Equal(other *DomainHash) bool {
	if hash == nil || other == nil {
		return hash == other
	}

	return hash.hashArray == other.hashArray
}

// Less returns true if hash is less than other
func (hash *DomainHash) Less(other *DomainHash) bool {
	return bytes.Compare(hash.hashArray[:], other.hashArray[:]) < 0
}

// LessOrEqual returns true if hash is smaller or equal to other
func (hash *DomainHash) LessOrEqual(other *DomainHash) bool {
	return bytes.Compare(hash.hashArray[:], other.hashArray[:]) <= 0
}

// CloneHashes returns a clone of the given hashes slice.
// Note: since DomainHash is a read-only type, the clone is shallow
// CloneHashes는 주어진 해시 슬라이스의 복제본을 반환합니다.
// 참고: DomainHash는 읽기 전용 유형이므로 복제본이 얕습니다.
func CloneHashes(hashes []*DomainHash) []*DomainHash {
	clone := make([]*DomainHash, len(hashes))
	copy(clone, hashes)
	return clone
}

// HashesEqual returns whether the given hash slices are equal.
func HashesEqual(a, b []*DomainHash) bool {
	if len(a) != len(b) {
		return false
	}

	for i, hash := range a {
		if !hash.Equal(b[i]) {
			return false
		}
	}
	return true
}
