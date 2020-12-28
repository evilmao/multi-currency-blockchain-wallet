package gtrx

const (
	// Precision is the precision of TRX.
	Precision = 6

	// TRX is the unit of TRX, 1 TRX = 10^6 SUN.
	TRX = 1000000

	// NormalTransferFee is 0.1 TRX.
	NormalTransferFee = 0.1 * TRX

	// smart contract
	MaxFeeLimit = 1000
	ParamLen    = 64
)
