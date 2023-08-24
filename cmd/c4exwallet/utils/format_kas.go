package utils

import (
	"fmt"

	"github.com/c4ei/c4exd/domain/consensus/utils/constants"
)

// FormatC4x takes the amount of sompis as uint64, and returns amount of C4X with 8  decimal places
func FormatC4x(amount uint64) string {
	res := "                   "
	if amount > 0 {
		res = fmt.Sprintf("%19.8f", float64(amount)/constants.SompiPerC4ex)
	}
	return res
}
