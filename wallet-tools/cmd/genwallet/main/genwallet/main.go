package main

import (
	"fmt"
	"strings"
	"time"

	"upex-wallet/wallet-base/cmd"

	"upex-wallet/wallet-tools/cmd/genwallet/keypair"
	_ "upex-wallet/wallet-tools/cmd/genwallet/keypair/builder"
	"upex-wallet/wallet-tools/cmd/genwallet/keypair/generator"
	"upex-wallet/wallet-tools/cmd/genwallet/keypair/storer"
	wallet "upex-wallet/wallet-tools/cmd/genwallet/wallet/v2"
)

var (
	RandomGenerator     = "random"
	DeriveGenerator     = "derive"
	AllGeneratorClasses = []string{RandomGenerator, DeriveGenerator}
)

var (
	password       string
	dataPath       string
	inputFileName  string
	outputFileName string

	builderClass   string
	generatorClass string

	normalNumber uint
	systemNumber uint

	rebuildMode bool
	appendMode  bool

	tag string
)

func main() {
	c := cmd.New(
		"genwallet",
		"generate wallet.dat and derived files for deposit & withdraw.",
		"./genwallet -c eth -n 3 -N 2",
		run,
	)
	// 设置外部参数
	c.Flags().StringVarP(&password, "password", "p", "123456", "the password of the wallet, for inputting in security mode, provide an empty string")
	c.Flags().StringVarP(&dataPath, "datapath", "d", "./", "the data path of the wallet")
	c.Flags().StringVarP(&inputFileName, "inputfile", "f", "wallet.dat", "the input file name of the wallet")
	c.Flags().StringVarP(&outputFileName, "outputfile", "o", "wallet.dat", "the output file name of the wallet")

	classes := keypair.AllBuilderClasses()
	c.Flags().StringVarP(&builderClass, "builderclass", "c", "", "the keypair builder class: "+strings.Join(classes, ","))

	c.Flags().StringVarP(&generatorClass, "generatorclass", "g", RandomGenerator, "the keypair generator class: "+strings.Join(AllGeneratorClasses, ","))

	c.Flags().UintVarP(&normalNumber, "normalNumber", "n", 0, "the number of normal address keypairs to generate")
	c.Flags().UintVarP(&systemNumber, "systemNumber", "N", 0, "the number of system address keypairs to generate")

	c.Flags().BoolVarP(&rebuildMode, "rebuild", "", false, "rebuild meta and other derived files")
	c.Flags().BoolVarP(&appendMode, "append", "a", false, "use append mode to generate keypairs")

	c.Flags().StringVarP(&tag, "tag", "t", "", "the tag of the addresses to generate")
	c.Execute()
}

func run(c *cmd.Command) error {
	err := initPassword(c)
	if err != nil {
		return err
	}
	// 货币符号
	builderClass = strings.ToUpper(builderClass)

	// 返回结构体
	meta := NewMeta(dataPath, inputFileName, outputFileName)
	// 默认false, 忽略
	if rebuildMode {
		return rebuild(meta)
	}

	// 系统用户,普通用户数量
	if normalNumber+systemNumber <= 0 {
		return nil
	}

	// 创建生成器, 默认是用RandomGenerator {buliderClass:"BTC"}
	g, err := createGenerator()
	if err != nil {
		return err
	}

	// 新建钱包,返回钱包结构体
	w := wallet.New(dataPath, inputFileName, outputFileName, g, nil)

	// 默认false, 忽略
	if appendMode {
		err = prepareAppend(meta, w, g)
		if err != nil {
			return err
		}
	}

	tm := time.Now()
	if normalNumber > 0 {
		// 使用Add方法,普通账户追加到meta.sections 数组中
		meta.Add(&storer.Section{
			Start:    w.Len(),                     // 标记普通用户的 起始位置
			End:      w.Len() + int(normalNumber), // 系统类型的起始位置
			Time:     tm,
			IsSystem: false,
			Tag:      tag,
		})
	}

	// 系统账户
	if systemNumber > 0 {
		meta.Add(&storer.Section{
			Start:    w.Len() + int(normalNumber),
			End:      w.Len() + int(normalNumber+systemNumber),
			Time:     tm,
			IsSystem: true,
			Tag:      tag,
		})
	}

	// 钱包结构体调用方法
	err = w.Generate(password, normalNumber+systemNumber)
	if err != nil {
		return err
	}

	err = store(meta, w)
	if err != nil {
		return err
	}

	// 输出 系统用户的用户私钥
	return getNormalHexPriKey(w, normalNumber, systemNumber)
}

func initPassword(c *cmd.Command) error {
	if rebuildMode || appendMode {
		return c.Password("password")
	}

	return c.PasswordWithConfirm("password", true)
}

func rebuild(meta *Meta) error {
	meta.Load()

	w := wallet.New(dataPath, inputFileName, outputFileName, nil, nil)
	err := w.Load()
	if err != nil {
		return err
	}

	if len(meta.Sections()) == 0 {
		meta.Add(&storer.Section{
			Start:    0,
			End:      w.Len(),
			Time:     time.Now(),
			IsSystem: false,
		})
	}

	return store(meta, w)
}

func createGenerator() (keypair.Generator, error) {
	switch generatorClass {
	case RandomGenerator:
		return generator.NewRandom(builderClass), nil
	case DeriveGenerator:
		return generator.NewDerive(builderClass), nil
	default:
		return nil, fmt.Errorf("unsupported generator class %s", generatorClass)
	}
}

func prepareAppend(meta *Meta, w *wallet.Wallet, g keypair.Generator) error {
	if len(builderClass) == 0 {
		return fmt.Errorf("builder class can't be empty")
	}

	err := meta.Load()
	if err != nil {
		return err
	}

	err = w.Load()
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

func createStore(meta *Meta) keypair.Storer {
	return keypair.NewCombineStorer(
		storer.NewDepositAddressFile(meta.Sections()),
		storer.NewDepositAddress(meta.Sections()),
		storer.NewWithdrawAddress(meta.Sections()))
}

func store(meta *Meta, w *wallet.Wallet) error {
	st := createStore(meta)
	w.SetExtStorer(st)
	err := w.Store(password)
	if err != nil {
		return err
	}

	return meta.Store()
}

func getNormalHexPriKey(w *wallet.Wallet, n, s uint) error {
	if n <= 0 {
		err := fmt.Errorf("Number of Normal address keypairs can not be empty.")
		return err
	}
	addressWithPriKeys := make(map[string]string, n)
	// var normalhexPrikeys []string
	hexPrivateKeys := w.HexPrivateKeys
	addressArray := w.AddressStrArray

	if len(hexPrivateKeys) > 0 && len(addressArray) > 0 {
		normlAddressArray := addressArray[:n]
		normalhexPrikeys := hexPrivateKeys[:n]
		for i, addr := range normlAddressArray {
			addressWithPriKeys[addr] = normalhexPrikeys[i]
		}
	}
	fmt.Printf("---------------------\nNormal address and private keys:\n%s \n--------------------", addressWithPriKeys)
	return nil
}
