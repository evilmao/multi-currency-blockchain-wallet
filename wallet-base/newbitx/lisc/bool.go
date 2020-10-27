package lisc

import "strings"

type Bool bool

func NewBool(v bool) *Bool {
	b := Bool(v)
	return &b
}

func (b *Bool) Type() ValueType {
	return BoolType
}

func (b *Bool) True() bool {
	return bool(*b)
}

func (b *Bool) False() bool {
	return !b.True()
}

func (b *Bool) Set(v bool) {
	*b = Bool(v)
}

func (b *Bool) String() string {
	if b.True() {
		return "true"
	}
	return "false"
}

func (b *Bool) Format(deep int) string {
	return strings.Repeat(" ", deep*4) + b.String()
}

func (b *Bool) Reset() {
	b.Set(false)
}
