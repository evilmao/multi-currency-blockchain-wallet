package txbuilder

import (
	"fmt"
)

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
