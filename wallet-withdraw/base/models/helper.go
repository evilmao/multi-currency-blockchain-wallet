package models

import (
	"sort"

	"upex-wallet/wallet-base/models"

	"github.com/shopspring/decimal"
)

func SelectAccount(accounts []*models.Account, amount decimal.Decimal) ([]*models.Account, bool) {
	if len(accounts) == 0 {
		return nil, false
	}

	// a1 >= a2 >= a3
	sort.Slice(accounts, func(i, j int) bool {
		return accounts[i].Balance.GreaterThan(*accounts[j].Balance)
	})

	total := decimal.Zero
	for i, acc := range accounts {
		total = total.Add(*acc.Balance)
		if total.GreaterThanOrEqual(amount) {
			return accounts[:i+1], true
		}
	}
	return accounts, false
}

// selectUTXO selects sorted utxos to match the amount (if amount > 0) and in the limit length.
func selectUTXO(utxos []*models.UTXO, amount decimal.Decimal, limitLen int, bigOrder bool) ([]*models.UTXO, decimal.Decimal, bool) {

	total := decimal.Zero
	if len(utxos) == 0 {
		return nil, total, false
	}

	if bigOrder {
		// u1 >= u2 >= u3
		sort.Slice(utxos, func(i, j int) bool {
			return utxos[i].Amount.GreaterThan(utxos[j].Amount)
		})
	} else {
		// u1 <= u2 <= u3
		sort.Slice(utxos, func(i, j int) bool {
			return utxos[i].Amount.LessThan(utxos[j].Amount)
		})
	}

	withAmount := amount.GreaterThan(decimal.Zero)

	for i, u := range utxos {
		total = total.Add(u.Amount)
		// 当满足账户下的 资金之和大于或等于要交易的金额即可,返回前几个utxo信息即可
		if withAmount && total.GreaterThanOrEqual(amount) {
			return utxos[:i+1], total, true
		}

		if i+1 >= limitLen {
			return utxos[:i+1], total, !withAmount
		}
	}
	return utxos, total, !withAmount
}

// SelectUTXO selects utxos to match the amount (if amount > 0) and in the limit length.
func SelectUTXO(address, symbol string, amount decimal.Decimal, limitLen int) ([]*models.UTXO, decimal.Decimal, bool) {
	utxos := models.GetUTXOsByAddress(address, symbol)
	return selectUTXO(utxos, amount, limitLen, true)
}

// SelectSmallUTXO selects small-utxos in [limitLen/3, limitLen] length.
func SelectSmallUTXO(symbol, address string, maxAmount decimal.Decimal, limitLen int) ([]*models.UTXO, decimal.Decimal, bool) {
	if maxAmount.LessThanOrEqual(decimal.Zero) {
		return nil, decimal.Zero, false
	}

	minLen := limitLen / 3
	utxos := models.GetSmallUTXOsByAddress(symbol, address, maxAmount)

	if len(utxos) < minLen {
		return nil, decimal.Zero, false
	}
	return selectUTXO(utxos, decimal.Zero, limitLen, false)
}

// SortAccountsByBalance a1 >= a2 >= a3
// account 中按balance 从大到小排序
func SortAccountsByBalance(accounts []*models.Account) ([]*models.Account, bool) {
	if len(accounts) == 0 {
		return nil, false
	}
	sort.Slice(accounts, func(i, j int) bool {
		return accounts[i].Balance.GreaterThan(*accounts[j].Balance)
	})
	return accounts, true
}
