package trx

import (
	"fmt"
	"strconv"
	"strings"

	"upex-wallet/wallet-base/currency"
	"upex-wallet/wallet-base/models"
	"upex-wallet/wallet-config/withdraw/transfer/config"
	"upex-wallet/wallet-withdraw/transfer/txbuilder/trx/gtrx"
)

const (
	NORMAL = "normal"
	TRC20  = "trc20"
)

var (
	supportAssets = map[string]*AssetInfo{}
)

type AssetInfo struct {
	ID              int
	Precision       int
	Type            string
	ContractAddress string
}

func InitSupportAssets(cfg *config.Config) error {
	supportAssets[strings.ToUpper(cfg.Currency)] = &AssetInfo{
		ID:        0,
		Precision: gtrx.Precision,
		Type:      NORMAL,
	}

	symbols := models.GetCurrencies()

	for _, s := range symbols {
		c := s.Symbol
		detail := currency.CurrencyDetail(c)
		if detail == nil {
			return fmt.Errorf("trx init support assets, can't find currency detail of %s", c)
		}

		if detail.IsToken() && detail.ChainBelongTo(cfg.Currency) {
			if strings.HasPrefix(strings.ToUpper(detail.Address), "T") {
				// 合约地址T开头是TRC20币种
				supportAssets[c] = &AssetInfo{
					Precision:       detail.Decimal,
					Type:            TRC20,
					ContractAddress: detail.Address,
				}
			} else {
				// 合约为数字ID类型是TRC10币种
				assetID, err := strconv.Atoi(detail.Address)
				if err != nil {
					return fmt.Errorf("strconv asset id of %s failed, %v", c, err)
				}

				supportAssets[c] = &AssetInfo{
					ID:        assetID,
					Precision: detail.Decimal,
					Type:      NORMAL,
				}
			}
		}
		return nil
	}
	return nil
}

func SupportAssetInfo(currency string) (*AssetInfo, bool) {
	currency = strings.ToUpper(currency)
	info, ok := supportAssets[currency]
	return info, ok
}
