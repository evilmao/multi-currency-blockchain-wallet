package lisc

import (
	"fmt"
	"strings"
)

type Pair struct {
	key    string
	values []Value
	index  map[string]Value
}

func NewPair(key string) *Pair {
	return &Pair{
		key:   key,
		index: make(map[string]Value),
	}
}

func (p *Pair) Type() ValueType {
	return PairType
}

func (p *Pair) Key() string {
	return p.key
}

func (p *Pair) HasKey() bool {
	return len(p.Key()) > 0
}

func (p *Pair) SetKey(key string) bool {
	if p.HasKey() {
		return false
	}

	p.key = key
	return true
}

func (p *Pair) Add(value Value) *Pair {
	if value == nil {
		return p
	}

	p.values = append(p.values, value)
	if value.Type() == PairType {
		if value := value.(*Pair); value.HasKey() {
			p.index[value.Key()] = value
		}
	}
	return p
}

func (p *Pair) AddNumber(v string) *Pair {
	return p.Add(NewNumber(v))
}

func (p *Pair) AddString(v string) *Pair {
	return p.Add(NewString(v))
}

func (p *Pair) AddBool(v bool) *Pair {
	return p.Add(NewBool(v))
}

func (p *Pair) ValueCount() int {
	return len(p.values)
}

func (p *Pair) Value(index int) (Value, bool) {
	if index < 0 || p.ValueCount() <= index {
		return nil, false
	}

	return p.values[index], true
}

func (p *Pair) ValueByKey(key string) (Value, bool) {
	value, ok := p.index[key]
	return value, ok
}

func (p *Pair) Find(paths ...interface{}) (Value, bool) {
	var value Value = p
	for _, path := range paths {
		if value.Type() != PairType {
			return nil, false
		}

		v := value.(*Pair)
		var ok bool
		switch path := path.(type) {
		case int:
			value, ok = v.Value(path)
		case uint:
			value, ok = v.Value(int(path))
		case string:
			value, ok = v.ValueByKey(path)
		default:
			return nil, false
		}

		if !ok {
			return nil, false
		}
	}
	return value, true
}

func (p *Pair) FindByType(valueType ValueType, paths ...interface{}) (Value, error) {
	value, ok := p.Find(paths...)
	if !ok {
		return nil, fmt.Errorf("can't find")
	}

	if value.Type() == valueType {
		return value, nil
	}

	if value.Type() == PairType {
		value := value.(*Pair)
		if value.ValueCount() > 0 {
			first, _ := value.Value(0)
			if first.Type() == valueType {
				return first, nil
			}
		}
	}

	return nil, fmt.Errorf("not %s, is %s", valueType, value.Type())
}

func (p *Pair) Int64(defaultValue int64, paths ...interface{}) (int64, error) {
	value, err := p.FindByType(NumberType, paths...)
	if err != nil {
		return defaultValue, err
	}

	d, err := value.(*Number).Int64()
	if err != nil {
		return defaultValue, fmt.Errorf("type convert failed, %v", err)
	}

	return d, nil
}

func (p *Pair) SetInt64(v int64, paths ...interface{}) error {
	value, err := p.FindByType(NumberType, paths...)
	if err != nil {
		return err
	}

	value.(*Number).SetInt64(v)
	return nil
}

func (p *Pair) Float64(defaultValue float64, paths ...interface{}) (float64, error) {
	value, err := p.FindByType(NumberType, paths...)
	if err != nil {
		return defaultValue, err
	}

	d, err := value.(*Number).Float64()
	if err != nil {
		return defaultValue, fmt.Errorf("type convert failed, %v", err)
	}

	return d, nil
}

func (p *Pair) SetFloat64(v float64, paths ...interface{}) error {
	value, err := p.FindByType(NumberType, paths...)
	if err != nil {
		return err
	}

	value.(*Number).SetFloat64(v)
	return nil
}

func (p *Pair) String(defaultValue string, paths ...interface{}) (string, error) {
	value, err := p.FindByType(StringType, paths...)
	if err != nil {
		return defaultValue, err
	}

	return value.(*String).String(), nil
}

func (p *Pair) SetString(v string, paths ...interface{}) error {
	value, err := p.FindByType(StringType, paths...)
	if err != nil {
		return err
	}

	value.(*String).Set(v)
	return nil
}

func (p *Pair) Bool(defaultValue bool, paths ...interface{}) (bool, error) {
	value, err := p.FindByType(BoolType, paths...)
	if err != nil {
		return defaultValue, err
	}

	return value.(*Bool).True(), nil
}

func (p *Pair) SetBool(v bool, paths ...interface{}) error {
	value, err := p.FindByType(BoolType, paths...)
	if err != nil {
		return err
	}

	value.(*Bool).Set(v)
	return nil
}

func (p *Pair) Pair(paths ...interface{}) (*Pair, error) {
	value, err := p.FindByType(PairType, paths...)
	if err != nil {
		return nil, err
	}

	return value.(*Pair), nil
}

func (p *Pair) Format(deep int) string {
	s := strings.Repeat(" ", deep*4) + "("
	if p.HasKey() {
		s += p.Key()
	}
	s += "\n"

	vn := p.ValueCount()
	for i, v := range p.values {
		s += v.Format(deep + 1)
		if i < vn-1 {
			s += "\n"
		}
	}
	return s + ")"
}

func (p *Pair) Reset() {
	p.values = nil
	p.index = make(map[string]Value)
}
