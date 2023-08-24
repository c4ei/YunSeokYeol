// Copyright (c) 2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package bech32

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

const charset = "qpzry9x8gf2tvdw0s3jn54khce6mua7l"
const checksumLength = 8

// For use in convertBits. Represents a number of bits to convert to or from and whether
// to add padding.
type conversionType struct {
	fromBits uint8
	toBits   uint8
	pad      bool
}

// Conversion types to use in convertBits.
// ConvertBits에서 사용할 변환 유형입니다.
var fiveToEightBits = conversionType{fromBits: 5, toBits: 8, pad: false}
var eightToFiveBits = conversionType{fromBits: 8, toBits: 5, pad: true}

var generator = []int{0x98f2bc8e61, 0x79b76d99e2, 0xf33e5fb3c4, 0xae2eabe2a8, 0x1e4f43e470}

// Encode prepends the version byte, converts to uint5, and encodes to Bech32.
// 인코딩은 버전 바이트 앞에 추가하고 uint5로 변환한 후 Bech32로 인코딩합니다.
func Encode(prefix string, payload []byte, version byte) string {
	data := make([]byte, len(payload)+1)
	data[0] = version
	copy(data[1:], payload)

	converted := convertBits(data, eightToFiveBits)

	return encode(prefix, converted)
}

// Decode decodes a string that was encoded with Encode.
// Decode는 Encode로 인코딩된 문자열을 디코딩합니다.
func Decode(encoded string) (string, []byte, byte, error) {
	prefix, decoded, err := decode(encoded)
	if err != nil {
		return "", nil, 0, err
	}

	converted := convertBits(decoded, fiveToEightBits)
	version := converted[0]
	payload := converted[1:]

	return prefix, payload, version, nil
}

// Decode decodes a Bech32 encoded string, returning the prefix
// and the data part excluding the checksum.
// Bech32로 인코딩된 문자열을 디코딩하여 접두사를 반환합니다.
// 그리고 체크섬을 제외한 데이터 부분입니다.
func decode(encoded string) (string, []byte, error) {
	// The minimum allowed length for a Bech32 string is 10 characters,
	// since it needs a non-empty prefix, a separator, and an 8 character
	// checksum.
	// Bech32 문자열에 허용되는 최소 길이는 10자입니다.
	// 비어 있지 않은 접두사, 구분 기호 및 8자가 필요하기 때문입니다.
	// 체크섬.
	if len(encoded) < checksumLength+2 {
		return "", nil, errors.Errorf("invalid bech32 string length %d",
			len(encoded))
	}
	// Only	ASCII characters between 33 and 126 are allowed.
	// 33~126 사이의 ASCII 문자만 허용됩니다.
	for i := 0; i < len(encoded); i++ {
		if encoded[i] < 33 || encoded[i] > 126 {
			return "", nil, errors.Errorf("invalid character in "+
				"string: '%c'", encoded[i])
		}
	}

	// The characters must be either all lowercase or all uppercase.
	// 문자는 모두 소문자이거나 모두 대문자여야 합니다.
	lower := strings.ToLower(encoded)
	upper := strings.ToUpper(encoded)
	if encoded != lower && encoded != upper {
		return "", nil, errors.Errorf("string not all lowercase or all " +
			"uppercase")
	}

	// We'll work with the lowercase string from now on.
	encoded = lower

	// The string is invalid if the last ':' is non-existent, it is the
	// first character of the string (no human-readable part) or one of the
	// last 8 characters of the string (since checksum cannot contain ':'),
	// or if the string is more than 90 characters in total.
	// 마지막 ':'이 존재하지 않으면 문자열은 유효하지 않습니다.
	// 문자열의 첫 번째 문자(사람이 읽을 수 있는 부분 없음) 또는 다음 중 하나
	// 문자열의 마지막 8자(체크섬에는 ':'이 포함될 수 없으므로),
	// 또는 문자열이 총 90자를 초과하는 경우.
	colonIndex := strings.LastIndexByte(encoded, ':')
	if colonIndex < 1 || colonIndex+checksumLength+1 > len(encoded) {
		return "", nil, errors.Errorf("invalid index of ':'")
	}

	// The prefix part is everything before the last ':'.
	prefix := encoded[:colonIndex]
	data := encoded[colonIndex+1:]

	// Each character corresponds to the byte with value of the index in
	// 'charset'.
	decoded, err := decodeFromBase32(data)
	if err != nil {
		return "", nil, errors.Errorf("failed converting data to bytes: "+
			"%s", err)
	}

	if !verifyChecksum(prefix, decoded) {
		checksum := encoded[len(encoded)-checksumLength:]
		expected := encodeToBase32(calculateChecksum(prefix,
			decoded[:len(decoded)-checksumLength]))

		return "", nil, errors.Errorf("checksum failed. Expected %s, got %s",
			expected, checksum)
	}

	// We exclude the last 8 bytes, which is the checksum.
	return prefix, decoded[:len(decoded)-checksumLength], nil
}

// Encode encodes a byte slice into a bech32 string with the
// prefix. Note that the bytes must each encode 5 bits (base32).
// 인코딩은 바이트 슬라이스를 bech32 문자열로 인코딩합니다.
// 접두사. 바이트는 각각 5비트(base32)를 인코딩해야 합니다.
func encode(prefix string, data []byte) string {
	// Calculate the checksum of the data and append it at the end.
	checksum := calculateChecksum(prefix, data)
	combined := append(data, checksum...)

	// The resulting bech32 string is the concatenation of the prefix, the
	// separator ':', data and checksum. Everything after the separator is
	// represented using the specified charset.
	// 결과 bech32 문자열은 접두사를 연결한 것입니다.
	// 구분 기호 ':', 데이터 및 체크섬. 구분 기호 뒤의 모든 내용은
	// 지정된 문자셋을 사용하여 표현됩니다.
	base32String := encodeToBase32(combined)

	return fmt.Sprintf("%s:%s", prefix, base32String)
}

// decodeFromBase32 converts each character in the string 'chars' to the value of the
// index of the correspoding character in 'charset'.
// decodeFromBase32는 문자열 'chars'의 각 문자를
// 'charset'에서 해당 문자의 인덱스입니다.
func decodeFromBase32(base32String string) ([]byte, error) {
	decoded := make([]byte, 0, len(base32String))
	for i := 0; i < len(base32String); i++ {
		index := strings.IndexByte(charset, base32String[i])
		if index < 0 {
			return nil, errors.Errorf("invalid character not part of "+
				"charset: %c", base32String[i])
		}
		decoded = append(decoded, byte(index))
	}
	return decoded, nil
}

// Converts the byte slice 'data' to a string where each byte in 'data'
// encodes the index of a character in 'charset'.
// IMPORTANT: this function expects the data to be in uint5 format.
// CAUTION: for legacy reasons, in case of an error this function returns
// an empty string instead of an error.
// 바이트 슬라이스 'data'를 'data'의 각 바이트가 포함된 문자열로 변환합니다.
// 'charset'에 있는 문자의 인덱스를 인코딩합니다.
// 중요: 이 함수는 데이터가 uint5 형식일 것으로 예상합니다.
// 주의: 기존 이유로 인해 오류가 발생하면 이 함수는 다음을 반환합니다.
// 오류 대신 빈 문자열입니다.
func encodeToBase32(data []byte) string {
	result := make([]byte, 0, len(data))
	for _, b := range data {
		if int(b) >= len(charset) {
			return ""
		}
		result = append(result, charset[b])
	}
	return string(result)
}

// convertBits converts a byte slice where each byte is encoding fromBits bits,
// to a byte slice where each byte is encoding toBits bits.
// ConvertBits는 각 바이트가 Bits 비트를 인코딩하는 바이트 슬라이스를 변환합니다.
// 각 바이트가 Bits 비트로 인코딩되는 바이트 슬라이스로.
func convertBits(data []byte, conversionType conversionType) []byte {
	// The final bytes, each byte encoding toBits bits.
	var regrouped []byte

	// Keep track of the next byte we create and how many bits we have
	// added to it out of the toBits goal.
	// 우리가 생성하는 다음 바이트와 우리가 가지고 있는 비트 수를 추적합니다.
	// toBits 목표에 추가되었습니다.
	nextByte := byte(0)
	filledBits := uint8(0)

	for _, b := range data {
		// Discard unused bits.
		b = b << (8 - conversionType.fromBits)

		// How many bits remaining to extract from the input data.
		remainingFromBits := conversionType.fromBits
		for remainingFromBits > 0 {
			// How many bits remaining to be added to the next byte.
			remainingToBits := conversionType.toBits - filledBits

			// The number of bytes to next extract is the minimum of
			// remainingFromBits and remainingToBits.
			toExtract := remainingFromBits
			if remainingToBits < toExtract {
				toExtract = remainingToBits
			}

			// Add the next bits to nextByte, shifting the already
			// added bits to the left.
			nextByte = (nextByte << toExtract) | (b >> (8 - toExtract))

			// Discard the bits we just extracted and get ready for
			// next iteration.
			b = b << toExtract
			remainingFromBits -= toExtract
			filledBits += toExtract

			// If the nextByte is completely filled, we add it to
			// our regrouped bytes and start on the next byte.
			if filledBits == conversionType.toBits {
				regrouped = append(regrouped, nextByte)
				filledBits = 0
				nextByte = 0
			}
		}
	}

	// We pad any unfinished group if specified.
	if conversionType.pad && filledBits > 0 {
		nextByte = nextByte << (conversionType.toBits - filledBits)
		regrouped = append(regrouped, nextByte)
		filledBits = 0
		nextByte = 0
	}

	return regrouped
}

// The checksum is a 40 bits BCH codes defined over GF(2^5).
// It ensures the detection of up to 6 errors in the address and 8 in a row.
// Combined with the length check, this provides very strong guarantee against errors.
// For more details please refer to the Bech32 Address Serialization section
// of the spec.
// 체크섬은 GF(2^5)에 정의된 40비트 BCH 코드입니다.
// 주소에서 최대 6개, 연속 8개 오류 감지를 보장합니다.
// 길이 확인과 결합되어 오류에 대한 매우 강력한 보장을 제공합니다.
// 자세한 내용은 Bech32 주소 직렬화 섹션을 참조하세요.
// 스펙의
func calculateChecksum(prefix string, payload []byte) []byte {
	prefixLower5Bits := prefixToUint5Array(prefix)
	payloadInts := ints(payload)
	templateZeroes := []int{0, 0, 0, 0, 0, 0, 0, 0}

	// prefixLower5Bits + 0 + payloadInts + templateZeroes
	concat := append(prefixLower5Bits, 0)
	concat = append(concat, payloadInts...)
	concat = append(concat, templateZeroes...)

	polyModResult := polyMod(concat)
	var res []byte
	for i := 0; i < checksumLength; i++ {
		res = append(res, byte((polyModResult>>uint(5*(checksumLength-1-i)))&31))
	}

	return res
}

// For more details please refer to the Bech32 Address Serialization section
// of the spec.
// 자세한 내용은 Bech32 주소 직렬화 섹션을 참조하세요.
// 스펙의
func verifyChecksum(prefix string, payload []byte) bool {
	prefixLower5Bits := prefixToUint5Array(prefix)
	payloadInts := ints(payload)

	// prefixLower5Bits + 0 + payloadInts
	dataToVerify := append(prefixLower5Bits, 0)
	dataToVerify = append(dataToVerify, payloadInts...)

	return polyMod(dataToVerify) == 0
}

func prefixToUint5Array(prefix string) []int {
	prefixLower5Bits := make([]int, len(prefix))
	for i := 0; i < len(prefix); i++ {
		char := prefix[i]
		charLower5Bits := int(char & 31)
		prefixLower5Bits[i] = charLower5Bits
	}

	return prefixLower5Bits
}

func ints(payload []byte) []int {
	payloadInts := make([]int, len(payload))
	for i, b := range payload {
		payloadInts[i] = int(b)
	}

	return payloadInts
}

// For more details please refer to the Bech32 Address Serialization section
// of the spec.
func polyMod(values []int) int {
	checksum := 1
	for _, value := range values {
		topBits := checksum >> 35
		checksum = ((checksum & 0x07ffffffff) << 5) ^ value
		for i := 0; i < len(generator); i++ {
			if ((topBits >> uint(i)) & 1) == 1 {
				checksum ^= generator[i]
			}
		}
	}

	return checksum ^ 1
}
