package deposit

import (
	"encoding/hex"
	"fmt"
	"math"
	"regexp"
	"strings"
	"unicode/utf8"

	"upex-wallet/wallet-base/models"
	"upex-wallet/wallet-tools/base/crypto"

	"github.com/asaskevich/govalidator"
)

var (
	// NormalTxHash valid tx hash format.
	NormalTxHash = regexp.MustCompile(`^[a-zA-Z0-9_+/=]+$`).MatchString
	txTagLength  = 100
)

// TxString formats tx to string.
func TxString(tx *models.Tx) string {
	if tx == nil {
		return ""
	}

	return fmt.Sprintf("symbol: %s, hash: %s, address: %s, amount: %s, confirm: %d, tag: %s",
		tx.Symbol, tx.Hash, tx.Address, tx.Amount, tx.Confirm, tx.Extra)
}

// GenSequenceID generates a hash string with size of 32.
func GenSequenceID(datas ...[]byte) string {
	var buf []byte
	for _, data := range datas {
		buf = append(buf, data...)
	}
	return hex.EncodeToString(crypto.Hash160(buf))[:32]
}

// ValidTxTag valid a tx's tag.
func ValidTxTag(tag, currency string) bool {
	tag = strings.TrimSpace(tag)
	if govalidator.IsNull(tag) {
		return false
	}

	switch {
	case strings.EqualFold(currency, "xmr"):
		data, err := hex.DecodeString(tag)
		return err == nil && len(data) == 32
	default:
		return govalidator.IsNumeric(tag)
	}
}

// TruncateTxTag truncate tx tag to length of txTagLength.
func TruncateTxTag(tag string) string {
	tag = strings.TrimSpace(tag)
	if !ValidUTF8MB3(tag) {
		return "not utf8mb3 string"
	}

	if utf8.RuneCountInString(tag) > txTagLength {
		uft8Str := []rune(tag)
		tag = string(uft8Str[0:txTagLength])
	}
	return tag
}

// ValidUTF8MB3 returns whether s is a valid [utf8mb3](https://dev.mysql.com/doc/refman/8.0/en/charset-unicode-utf8mb3.html) string.
func ValidUTF8MB3(s string) bool {
	for _, c := range s {
		if len(string(c)) > 3 {
			return false
		}
	}
	return true
}

// CalculateConfirm calculates tx confirmations.
func CalculateConfirm(txHeight, currentHeight int64) uint16 {
	if txHeight <= 0 || currentHeight <= 0 || currentHeight < txHeight {
		return 0
	}

	confirm := currentHeight - txHeight + 1
	if confirm > int64(math.MaxUint16) {
		return math.MaxUint16
	}

	return uint16(confirm)
}
