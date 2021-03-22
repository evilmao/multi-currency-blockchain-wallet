package alarm

import (
    "fmt"
    "github.com/shopspring/decimal"
)

// ErrFeeNotEnough
type ErrorBalanceNotEnough struct {
    EmailContent string
}

func NewErrorBalanceNotEnough(balance, amount interface{}) *ErrorBalanceNotEnough {
    return &ErrorBalanceNotEnough{
        EmailContent: fmt.Sprintf("账户余额(%v)不足,交易金额为:[%v]", balance, amount),
    }
}

func (e *ErrorBalanceNotEnough) Error() string {
    return fmt.Sprintf("wallet balance not enough")
}

// ErrFeeNotEnough def.
type ErrorBalanceLessThanFee struct {
    EmailContent string
}

func NewErrorBalanceLessThanFee(fee decimal.Decimal) *ErrorBalanceLessThanFee {
    return &ErrorBalanceLessThanFee{
        EmailContent: fmt.Sprintf("钱包余额不足,低于交易手续费[%v];", fee),
    }
}

func (e *ErrorBalanceLessThanFee) Error() string {
    return fmt.Sprintf("no avaliable accounts for system ")
}

// ErrorBalanceLessCost.
type ErrorBalanceLessCost struct {
    EmailContent string
}

func NewErrorBalanceLessCost(fee, balance, amount decimal.Decimal) *ErrorBalanceLessCost {
    return &ErrorBalanceLessCost{
        EmailContent: fmt.Sprintf("钱包余额(%v)不足: 转账金额为(%v),所需手续费为(%v);", balance, fee, amount),
    }
}

func (e *ErrorBalanceLessCost) Error() string {
    return fmt.Sprintf("wallet balance less than cost ")
}

// ErrorBalanceLessCost.
type NotMatchAccount struct {
    EmailContent string
}

func NewNotMatchAccount(fee, amount, balance decimal.Decimal, address string) *NotMatchAccount {
    difference := amount.Sub(balance)
    return &NotMatchAccount{
        EmailContent: fmt.Sprintf("未匹配到满足当前交易的系统地址; 当前交易金额为: %v ;手续费为: %v ;当前系统地址中最大金额为: %v,地址为:[%s],至少需要再转入金额为:[%v]", fee, balance, amount, address, difference),
    }
}

func (e *NotMatchAccount) Error() string {
    return fmt.Sprintf("don't match the appropriate system address ")
}

type ErrorAccountBalanceNotEnough struct {
    Address      string
    NeedFee      decimal.Decimal
    EmailContent string
}

func NewErrorAccountBalanceNotEnough(address string, symbol string, needFee decimal.Decimal) *ErrorAccountBalanceNotEnough {
    return &ErrorAccountBalanceNotEnough{
        Address:      address,
        NeedFee:      needFee,
        EmailContent: fmt.Sprintf("地址(%s),%s余额不足(低于交易手续费%s)", address, symbol, needFee.String()),
    }
}

func (e *ErrorAccountBalanceNotEnough) Error() string {
    return fmt.Sprintf("address balance not enough for transaction fee ")
}

type WarnCoolWalletBalanceChange struct {
    WarnDetail string
}

func NewWarnCoolWalletBalanceChange(maxRemainBalance, amount decimal.Decimal, coldAddress, txID string) *WarnCoolWalletBalanceChange {
    return &WarnCoolWalletBalanceChange{
        WarnDetail: fmt.Sprintf("钱包资金超过预留最大值: %s, 当前余额为:%s, 已转入冷钱包(%s),交易ID:%s", maxRemainBalance.String(), amount.String(), coldAddress, txID),
    }
}

func (e *WarnCoolWalletBalanceChange) Error() string {
    return fmt.Sprintf("cool wallet balance has changed ")
}

// ErrFeeNotEnough def.
type ErrFeeNotEnough struct {
    Fee          decimal.Decimal
    NeedFee      decimal.Decimal
    EmailContent string
}

func NewErrFeeNotEnough(fee, needFee decimal.Decimal) *ErrFeeNotEnough {
    return &ErrFeeNotEnough{
        Fee:          fee,
        NeedFee:      needFee,
        EmailContent: fmt.Sprintf("手续费不足,当前手续费:[%v],所需手续费:[%v]", fee, needFee),
    }
}

func (e *ErrFeeNotEnough) Error() string {
    return fmt.Sprintf("tx fee not enough, need: %s, got: %s", e.NeedFee, e.Fee)
}
