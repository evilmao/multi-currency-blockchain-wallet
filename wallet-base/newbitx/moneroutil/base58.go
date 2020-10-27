package moneroutil

import (
	"fmt"
	"math/big"
	"regexp"
	"strings"
)

const (
	BASE58 = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
)

var base58Lookup = map[string]int{
	"1": 0, "2": 1, "3": 2, "4": 3, "5": 4, "6": 5, "7": 6, "8": 7,
	"9": 8, "A": 9, "B": 10, "C": 11, "D": 12, "E": 13, "F": 14, "G": 15,
	"H": 16, "J": 17, "K": 18, "L": 19, "M": 20, "N": 21, "P": 22, "Q": 23,
	"R": 24, "S": 25, "T": 26, "U": 27, "V": 28, "W": 29, "X": 30, "Y": 31,
	"Z": 32, "a": 33, "b": 34, "c": 35, "d": 36, "e": 37, "f": 38, "g": 39,
	"h": 40, "i": 41, "j": 42, "k": 43, "m": 44, "n": 45, "o": 46, "p": 47,
	"q": 48, "r": 49, "s": 50, "t": 51, "u": 52, "v": 53, "w": 54, "x": 55,
	"y": 56, "z": 57,
}
var (
	bigBase   = big.NewInt(58)
	converter = newBlockSizeConverter()
)

// Monero base58 chucks every 11 bytes.
type blockSizeConverter struct {
	encodedBlockTable []int
	decodedBlockTable map[int]int
}

func newBlockSizeConverter() *blockSizeConverter {
	s := &blockSizeConverter{}
	s.encodedBlockTable = []int{0, 2, 3, 5, 6, 7, 9, 10, 11}
	s.decodedBlockTable = map[int]int{
		0: 0, 2: 1, 3: 2, 5: 3, 6: 4, 7: 5, 9: 6, 10: 7, 11: 8}
	return s
}

func (s *blockSizeConverter) EncodedSize(n int) int {
	if n < 0 || len(s.encodedBlockTable) <= n {
		return -1
	}

	return s.encodedBlockTable[n]
}

func (s *blockSizeConverter) DecodedSize(n int) int {
	ret, ok := s.decodedBlockTable[n]
	if !ok {
		return -1
	}

	return ret
}

func encodeChunk(raw []byte) (result string) {
	padding := converter.EncodedSize(len(raw))
	if padding < 0 {
		fmt.Println("WTF", len(raw), padding)
		return ""
	}
	remainder := new(big.Int)
	remainder.SetBytes(raw)
	bigZero := new(big.Int)
	for remainder.Cmp(bigZero) > 0 {
		current := new(big.Int)
		remainder.DivMod(remainder, bigBase, current)
		result = string(BASE58[current.Int64()]) + result
	}
	if len(result) < padding {
		result = strings.Repeat("1", (padding-len(result))) + result
	}
	return
}

func decodeChunk(encoded string) (result []byte) {
	decodedSize := converter.DecodedSize(len(encoded))
	if decodedSize < 0 {
		return nil
	}
	result = make([]byte, decodedSize)

	bigResult := big.NewInt(0)
	currentMultiplier := big.NewInt(1)
	tmp := new(big.Int)
	for i := len(encoded) - 1; i >= 0; i-- {
		tmp.SetInt64(int64(base58Lookup[string(encoded[i])]))
		tmp.Mul(currentMultiplier, tmp)
		bigResult.Add(bigResult, tmp)
		currentMultiplier.Mul(currentMultiplier, bigBase)
	}
	raw := bigResult.Bytes()
	copy(result[decodedSize-len(raw):], raw)
	return
}

func EncodeMoneroBase58(data ...[]byte) (result string) {
	var combined []byte
	for _, item := range data {
		combined = append(combined, item...)
	}
	length := len(combined)
	rounds := length / 8
	for i := 0; i < rounds; i++ {
		result += encodeChunk(combined[i*8 : (i+1)*8])
	}
	if length%8 > 0 {
		result += encodeChunk(combined[rounds*8:])
	}
	return
}

func DecodeMoneroBase58(data string) (result []byte) {
	if len(data) == 0 {
		return nil
	}

	regex := regexp.MustCompile(fmt.Sprintf("[^%s]+", BASE58))
	if regex.MatchString(data) {
		return nil
	}

	length := len(data)
	fullBlockCount := length / 11
	lastBlockSize := length % 11
	lastBlockDecodedSize := converter.DecodedSize(lastBlockSize)
	if lastBlockDecodedSize < 0 {
		return nil
	}

	for i := 0; i < fullBlockCount; i++ {
		result = append(result, decodeChunk(data[i*11:(i+1)*11])...)
	}

	if lastBlockSize > 0 {
		result = append(result, decodeChunk(data[fullBlockCount*11:])...)
	}
	return
}
