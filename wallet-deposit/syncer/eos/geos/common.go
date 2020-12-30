package geos

import (
	"strings"

	"github.com/buger/jsonparser"
)

func CheckActionName(actionName, currency string) bool {
	if strings.ToLower(currency) == "abbc" {
		// blockchain abbc support transfer and transferx
		if actionName == "transfer" || actionName == "transferx" {
			return true
		}
	} else if strings.ToLower(currency) == "dccy" {
		if actionName == "transfer" || actionName == "extransfer" {
			return true
		}
	} else {
		if actionName == "transfer" {
			return true
		}
	}

	return false
}

func GetTransferInfo(data []byte, name string) ([]string, string, error) {
	if name == "extransfer" {
		quantity, err := jsonparser.GetString(data, "act", "data", "quantity", "quantity")
		if err != nil {
			return nil, "", err
		}
		address, err := jsonparser.GetString(data, "act", "data", "quantity", "contract")
		if err != nil {
			return nil, "", err
		}
		amountParts := strings.Split(quantity, " ")
		return amountParts, address, nil
	}
	quantity, err := jsonparser.GetString(data, "act", "data", "quantity")
	if err != nil {
		return nil, "", err
	}
	amountParts := strings.Split(quantity, " ")
	return amountParts, "", nil
}
