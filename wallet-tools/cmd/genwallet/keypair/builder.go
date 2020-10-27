package keypair

import (
	"fmt"
	"sort"
	"strings"
)

// 构造器 接口, 用于构造各个钱包
// Class 方法 string 类型
// Build 方法 KeyPair(接口) 类型
type Builder interface {
	Class() string
	Build() KeyPair
}

var (
	// 定义builder变量为map类型并初始化
	builders = make(map[string]Builder)
)

// 注册构造器函数
func RegisterBuilder(builder Builder) error {
	if builder == nil {
		return nil
	}

	// string 类型
	class := builder.Class()
	if len(class) == 0 {
		return fmt.Errorf("keypair class can't be empty")
	}
	// string 转大写
	class = strings.ToUpper(class)
	if _, ok := builders[class]; ok {
		return fmt.Errorf("duplicated keypair class: %s", class)
	}

	builders[class] = builder
	return nil
}

func FindBuilder(class string) (Builder, bool) {
	class = strings.ToUpper(class)
	builder, ok := builders[class]
	return builder, ok
}

func AllBuilderClasses() []string {
	if len(builders) == 0 {
		return nil
	}

	classes := make([]string, 0, len(builders))
	for c := range builders {
		classes = append(classes, c)
	}
	sort.Strings(classes)
	return classes
}

func Random(class string) (KeyPair, error) {
	builder, ok := FindBuilder(class)
	if !ok {
		return nil, fmt.Errorf("invalid keypair class: %s", class)
	}

	// 密钥结构体
	kp := builder.Build()
	err := kp.Random()
	if err != nil {
		return nil, err
	}

	return kp, nil
}

func Build(class string, privateKey []byte) (KeyPair, error) {
	builder, ok := FindBuilder(class)
	if !ok {
		return nil, fmt.Errorf("invalid keypair class: %s", class)
	}

	kp := builder.Build()
	err := kp.SetPrivateKey(privateKey)
	if err != nil {
		return nil, err
	}

	return kp, nil
}
