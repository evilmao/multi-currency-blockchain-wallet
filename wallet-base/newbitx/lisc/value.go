package lisc

type ValueType int

func (t ValueType) String() string {
	switch t {
	case NumberType:
		return "NumberType"
	case StringType:
		return "StringType"
	case BoolType:
		return "BoolType"
	case PairType:
		return "PairType"
	default:
		return "UnknownType"
	}
}

// Value types
const (
	NumberType ValueType = 0
	StringType ValueType = 1
	BoolType   ValueType = 2
	PairType   ValueType = 10
)

type Value interface {
	Type() ValueType
	Format(deep int) string
	Reset()
}
