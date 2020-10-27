package crypto

import (
	"github.com/dchest/blake256"
)

const CheckSumLen = 4

func CheckSum(input []byte) (cksum [CheckSumLen]byte) {
	h := DoubleSha256(input)
	copy(cksum[:], h[:CheckSumLen])
	return
}

func Blake256CheckSum(input []byte) (cksum [CheckSumLen]byte) {
	h := Sum(Sum(input, blake256.New()), blake256.New())
	copy(cksum[:], h[:CheckSumLen])
	return
}
