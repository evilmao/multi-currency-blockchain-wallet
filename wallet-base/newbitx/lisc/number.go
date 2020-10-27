package lisc

import (
	"fmt"
	"strings"
)

type Number struct {
	raw string
	num string
}

func NewNumber(v string) *Number {
	return &Number{
		raw: v,
		num: strings.Replace(v, "'", "", -1),
	}
}

func (n *Number) Type() ValueType {
	return NumberType
}

func (n *Number) Int64() (int64, error) {
	var d int64
	_, err := fmt.Sscanf(n.num, "%d", &d)
	if err != nil {
		return 0, err
	}
	return d, nil
}

func (n *Number) Float64() (float64, error) {
	var d float64
	_, err := fmt.Sscanf(n.num, "%f", &d)
	if err != nil {
		return 0, err
	}
	return d, nil
}

func (n *Number) SetInt64(v int64) {
	n.raw = fmt.Sprintf("%d", v)
	n.num = n.raw
}

func (n *Number) SetFloat64(v float64) {
	n.raw = fmt.Sprintf("%f", v)
	n.num = n.raw
}

func (n *Number) String() string {
	return n.num
}

func (n *Number) Format(deep int) string {
	return strings.Repeat(" ", deep*4) + n.raw
}

func (n *Number) Reset() {
	n.SetInt64(0)
}
