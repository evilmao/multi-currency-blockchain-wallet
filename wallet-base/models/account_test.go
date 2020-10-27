package models_test

import (
	"fmt"
	"strings"
	"testing"

	"upex-wallet/wallet-base/db"
	"upex-wallet/wallet-base/models"

	"github.com/shopspring/decimal"
)

type DBDecimalTestCase struct {
	Amount decimal.Decimal
	Add    bool
}

type DBDecimalTestCases []DBDecimalTestCase

func (c DBDecimalTestCases) Add(amount decimal.Decimal, repeat int) DBDecimalTestCases {
	if repeat < 1 {
		repeat = 1
	}

	for i := 0; i < repeat; i++ {
		c = append(c, DBDecimalTestCase{
			Amount: amount,
			Add:    true,
		})
	}
	return c
}

func (c DBDecimalTestCases) Sub(amount decimal.Decimal, repeat int) DBDecimalTestCases {
	if repeat < 1 {
		repeat = 1
	}

	for i := 0; i < repeat; i++ {
		c = append(c, DBDecimalTestCase{
			Amount: amount,
			Add:    false,
		})
	}
	return c
}

func testDBDecimal(t *testing.T, dsn string, testCases DBDecimalTestCases) {
	dbInstance, err := db.New(dsn, "")
	if err != nil {
		t.Fatal(err)
	}
	defer dbInstance.Close()

	dbInstance.AutoMigrate(&models.Account{})

	testAddress := strings.Repeat("0", 10)
	dbInstance.Where("address=?", testAddress).Delete(&models.Account{})

	acc := models.Account{
		Address:  testAddress,
		SymbolID: 0,
	}
	acc.Insert()
	defer func() {
		dbInstance.Where("address=?", testAddress).Delete(&models.Account{})
	}()

	var total decimal.Decimal
	for i, c := range testCases {
		data := map[string]interface{}{
			"balance": c.Amount,
		}
		if c.Add {
			data["op"] = "add"
			total = total.Add(c.Amount)
		} else {
			data["op"] = "sub"
			total = total.Sub(c.Amount)
		}

		dbInstance.Where("address=?", testAddress).First(&acc)

		err = acc.ForUpdate(data)
		if err != nil {
			t.Fatal(fmt.Sprintf("update account at index %d failed, %v", i, err))
		}

		dbInstance.Where("address=?", testAddress).First(&acc)

		fmt.Printf("balance: %s, after case %d, correct: %t\n", acc.Balance, i, acc.Balance.Equal(total))

		if !acc.Balance.Equal(total) {
			t.Fatal(fmt.Sprintf("balance not correct, need: %s, got: %s.", total, acc.Balance))
		}
	}

	fmt.Println("*** all test cases are ok ***")
}

func TestAccountDBDecimal(t *testing.T) {
	const dsn = "root:123qwe!@#QWE@tcp(127.0.0.1:3306)/sandbox_xxx?charset=utf8&parseTime=True&loc=Local"

	testDBDecimal(t, dsn, DBDecimalTestCases{}.
		Add(decimal.New(999999999999999999, -20), 11).
		Sub(decimal.New(999999999999999999, -20), 11).
		Add(decimal.New(99999999999, 0), 10).
		Sub(decimal.New(99999999999, 0), 10))
}
