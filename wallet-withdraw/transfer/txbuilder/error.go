package txbuilder

import (
	"fmt"

	"github.com/shopspring/decimal"
)

// ErrFeeNotEnough def.
type ErrFeeNotEnough struct {
	Fee     decimal.Decimal
	NeedFee decimal.Decimal
}

func NewErrFeeNotEnough(fee, needFee decimal.Decimal) *ErrFeeNotEnough {
	return &ErrFeeNotEnough{
		Fee:     fee,
		NeedFee: needFee,
	}
}

func (e *ErrFeeNotEnough) Error() string {
	return fmt.Sprintf("tx fee not enough, need: %s, got: %s", e.NeedFee, e.Fee)
}

// ErrBalanceForFeeNotEnough def.
type ErrBalanceForFeeNotEnough struct {
	Address string
	NeedFee decimal.Decimal
}

func NewErrBalanceForFeeNotEnough(address string, needFee decimal.Decimal) *ErrBalanceForFeeNotEnough {
	return &ErrBalanceForFeeNotEnough{
		Address: address,
		NeedFee: needFee,
	}
}

func (e *ErrBalanceForFeeNotEnough) Error() string {
	return fmt.Sprintf("address balance of %s not enough for fee of %s", e.Address, e.NeedFee)
}

// ErrUnsupportedCurrency def.
type ErrUnsupportedCurrency struct {
	Currency string
}

func NewErrUnsupportedCurrency(currency string) *ErrUnsupportedCurrency {
	return &ErrUnsupportedCurrency{
		Currency: currency,
	}
}

func (e *ErrUnsupportedCurrency) Error() string {
	return fmt.Sprintf("unsupported currency of %s", e.Currency)
}

//
