package lisc

import (
	"fmt"
	"strings"
)

type String string

func NewString(v string) *String {
	s := String(v)
	return &s
}

func (s *String) Type() ValueType {
	return StringType
}

func (s *String) String() string {
	return string(*s)
}

func (s *String) Set(v string) {
	*s = String(v)
}

func (s *String) Format(deep int) string {
	temp := strings.Replace(s.String(), `"`, `\"`, -1)
	return strings.Repeat(" ", deep*4) + fmt.Sprintf(`"%s"`, temp)
}

func (s *String) Reset() {
	s.Set("")
}
