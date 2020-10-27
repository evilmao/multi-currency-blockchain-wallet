package main

import (
	"encoding/hex"
	"fmt"
	"strings"

	"upex-wallet/wallet-tools/cmd/genwallet/keypair"
	_ "upex-wallet/wallet-tools/cmd/genwallet/keypair/builder"
	"upex-wallet/wallet-tools/cmd/genwallet/keypair/generator"
	"upex-wallet/wallet-tools/cmd/genwallet/keypair/storer"
	wallet "upex-wallet/wallet-tools/cmd/genwallet/wallet/v2"

	"upex-wallet/wallet-base/cmd"
)

const version = "v2"

var (
	RandomGenerator  = "random"
	DeriveGenerator  = "derive"
	ConvertGenerator = "convert"

	AllGeneratorClasses = []string{RandomGenerator, DeriveGenerator, ConvertGenerator}
)

var (
	showVersion bool

	password    string
	setPassword bool
	newPassword string

	dataPath        string
	inputFileName   string
	outputFileName  string
	withAddressFile bool

	builderClass   string
	generatorClass string
	forceMode      bool

	number     uint
	appendMode bool

	safeListMode bool
	listExtMode  bool
	listAllMode  bool

	fromDataV1         bool
	fromDataPreV1      bool
	fromDataCoinlibXLM bool
	fromDataCoinlibV1  bool
	fromDataIotaV1     bool
	fromDataPyBTC      string
)

func main() {
	c := cmd.New(
		"walletctl",
		"read/generate/convert wallet.dat.",
		"./walletctl -c BTC -n 1",
		run,
	)
	c.Flags().BoolVarP(&showVersion, "version", "v", false, "show version of walletctl")

	c.Flags().StringVarP(&password, "password", "p", "123456", "the password of the wallet, for inputting in security mode, provide an empty string")
	c.Flags().BoolVarP(&setPassword, "setpassword", "", false, "set a new password for the wallet")
	c.Flags().StringVarP(&newPassword, "newpassword", "", "", "the new password for the wallet, for inputting in security mode, provide an empty string")

	c.Flags().StringVarP(&dataPath, "datapath", "d", "./", "the data path of the wallet")
	c.Flags().StringVarP(&inputFileName, "inputfile", "f", "wallet.dat", "the input file name of the wallet")
	c.Flags().StringVarP(&outputFileName, "outputfile", "o", "wallet.dat", "the output file name of the wallet")
	c.Flags().BoolVarP(&withAddressFile, "withaddress", "", false, "generate the address file.")

	classes := keypair.AllBuilderClasses()
	c.Flags().StringVarP(&builderClass, "builderclass", "c", "", "the keypair builder class: "+strings.Join(classes, ","))

	c.Flags().StringVarP(&generatorClass, "generatorclass", "g", RandomGenerator, "the keypair generator class: "+strings.Join(AllGeneratorClasses, ","))
	c.Flags().BoolVarP(&forceMode, "force", "", false, "force convert when generatorclass(-g) is 'convert'")

	c.Flags().UintVarP(&number, "number", "n", 0, "the number of keypairs to generate or list")
	c.Flags().BoolVarP(&appendMode, "append", "a", false, "use append mode to generate keypairs")

	c.Flags().BoolVarP(&safeListMode, "list", "l", false, "list keypairs info except of private key and ext-data")
	c.Flags().BoolVarP(&listExtMode, "listext", "L", false, "list keypairs info except of private key")
	c.Flags().BoolVarP(&listAllMode, "listall", "", false, "list keypairs info include private key")

	c.Flags().BoolVarP(&fromDataV1, "fromdatav1", "", false, "read wallet.dat of version v1")
	c.Flags().BoolVarP(&fromDataPreV1, "fromdataprev1", "", false, "read wallet.dat of version pre-v1")
	c.Flags().BoolVarP(&fromDataCoinlibXLM, "fromdatacoinlibxlm", "", false, "read wallet.dat of coinlib's xlm format")
	c.Flags().BoolVarP(&fromDataCoinlibV1, "fromdatacoinlibv1", "", false, "read wallet.dat of coinlib's v1(eos/xrp/eth) format")
	c.Flags().BoolVarP(&fromDataIotaV1, "fromdataiotav1", "", false, "read wallet.dat of iota-v1 format")
	c.Flags().StringVarP(&fromDataPyBTC, "fromdatapybtc", "", "", "read wallet.dat of python-exwallet format")
	c.Execute()
}

func run(c *cmd.Command) error {
	if showVersion {
		fmt.Println("version:", version)
		return nil
	}

	err := initPassword(c)
	if err != nil {
		return err
	}

	builderClass = strings.ToUpper(builderClass)

	// List modes.
	if safeListMode || listExtMode || listAllMode {
		return listWallet()
	}

	// Change password.
	if setPassword {
		return changePassword(c)
	}

	g, err := createGenerator()
	if err != nil {
		if err, ok := err.(*ErrNoNeedConvert); ok {
			if withAddressFile {
				err.w.SetExtStorer(storer.NewAddressFile())
			}
			return err.w.Store(password)
		}

		return err
	}

	if number <= 0 {
		return nil
	}

	var st keypair.Storer
	if withAddressFile {
		st = storer.NewAddressFile()
	}
	w := wallet.New(dataPath, inputFileName, outputFileName, g, st)

	if appendMode {
		err = prepareAppend(w, g)
		if err != nil {
			return err
		}
	}

	err = w.Generate(password, number)
	if err != nil {
		return err
	}

	return w.Store(password)
}

func initPassword(c *cmd.Command) error {
	if setPassword ||
		appendMode ||
		safeListMode || listExtMode || listAllMode ||
		generatorClass == ConvertGenerator {

		return c.Password("password")
	}

	return c.PasswordWithConfirm("password", true)
}

func loadWallet(w *wallet.Wallet) error {
	if len(fromDataPyBTC) > 0 {
		return w.LoadPyBTC(password, fromDataPyBTC)
	}

	if fromDataIotaV1 {
		return w.LoadIotaV1(password, number)
	}

	if fromDataCoinlibXLM {
		return w.LoadCoinlibXLM(password)
	}

	if fromDataPreV1 {
		return w.LoadPreV1(password)
	}

	if fromDataV1 {
		return w.LoadV1(password)
	}

	if fromDataCoinlibV1 {
		return w.LoadCoinlibV1(password, builderClass)
	}

	return w.Load()
}

func listWallet() error {
	w := wallet.New(dataPath, inputFileName, outputFileName, nil, nil)
	err := loadWallet(w)
	if err != nil {
		return err
	}

	max := int(number)
	if max <= 0 || w.Len() < max {
		max = w.Len()
	}

	printInfo := &WalletPrintInfo{}
	for i := 0; i < max; i++ {
		kp, err := w.KeyPairAtIndex(password, i)
		if err != nil {
			return fmt.Errorf("get keypair at index %d failed, %v", i, err)
		}

		if i == 0 {
			printInfo.Class = kp.Class()
			printInfo.Cryptography = string(kp.Cryptography())
			printInfo.Total = w.Len()
		}

		item := NewWalletPrintInfoItem()
		item.ID = i + 1
		if listAllMode {
			item.PrivateKey = hex.EncodeToString(kp.PrivateKey())
		}

		item.PublicKey = hex.EncodeToString(kp.PublicKey())
		item.Address = kp.AddressString()

		if listExtMode || listAllMode {
			if kp, ok := kp.(keypair.WithExtData); ok {
				for k, v := range kp.ExtData() {
					item.ExtData[k] = v.String()
				}
			}
		}

		printInfo.Add(item)
	}

	fmt.Println(printInfo)
	return nil
}

func changePassword(c *cmd.Command) error {
	w := wallet.New(dataPath, inputFileName, outputFileName, nil, nil)
	err := loadWallet(w)
	if err != nil {
		return err
	}

	err = c.PasswordWithConfirm("newpassword", true)
	if err != nil {
		return err
	}

	err = w.ChangePassword(password, newPassword)
	if err != nil {
		return err
	}

	return w.Store(newPassword)
}

type ErrNoNeedConvert struct {
	w *wallet.Wallet
}

func (e *ErrNoNeedConvert) Error() string {
	return "no need to convert"
}

func createGenerator() (keypair.Generator, error) {
	switch generatorClass {
	case RandomGenerator:
		return generator.NewRandom(builderClass), nil
	case DeriveGenerator:
		return generator.NewDerive(builderClass), nil
	case ConvertGenerator:
		appendMode = false

		w := wallet.New(dataPath, inputFileName, outputFileName, nil, nil)
		err := loadWallet(w)
		if err != nil {
			return nil, err
		}

		if w.Len() <= 0 {
			return nil, fmt.Errorf("can't convert keypair from empty wallet")
		}

		if len(builderClass) == 0 || strings.EqualFold(builderClass, w.Class()) {
			return nil, &ErrNoNeedConvert{w}
		}

		builder, ok := keypair.FindBuilder(builderClass)
		if !ok {
			return nil, fmt.Errorf("invalid keypair class: %s", builderClass)
		}

		kp := builder.Build()
		if !forceMode {
			if walletCryptography := w.Cryptography(); kp.Cryptography() != walletCryptography {
				return nil, fmt.Errorf("can't convert %s wallet to %s wallet", walletCryptography, kp.Cryptography())
			}
		}

		if number <= 0 || uint(w.Len()) < number {
			number = uint(w.Len())
		}

		return generator.NewConvert(generator.NewFromWalletV2(password, w), builderClass), nil
	default:
		return nil, fmt.Errorf("unsupported generator class %s", generatorClass)
	}
}

func prepareAppend(w *wallet.Wallet, g keypair.Generator) error {
	if len(builderClass) == 0 {
		return fmt.Errorf("builder class can't be empty")
	}

	err := loadWallet(w)
	if err != nil {
		return err
	}

	if w.Len() <= 0 {
		return fmt.Errorf("can't append keypair into empty wallet")
	}

	lastKP, err := w.LastKeyPair(password)
	if err != nil {
		return fmt.Errorf("get last keypair failed, %v", err)
	}

	if builderClass != lastKP.Class() {
		return fmt.Errorf("can't append %s keypair into %s wallet", builderClass, lastKP.Class())
	}

	if generatorClass == DeriveGenerator {
		deriveG := g.(*generator.Derive)
		err := deriveG.SetOrigin(lastKP)
		if err != nil {
			return err
		}
	}
	return nil
}
