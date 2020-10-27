package lisc

import (
	"fmt"
	"strings"
	"unicode"
)

func parseComment(data []rune) (int, error) {
	for i := 0; i < len(data); i++ {
		if data[i] == '\n' {
			return i + 1, nil
		}
	}
	return len(data), nil
}

func parsePair(data []rune, pair *Pair) (*Pair, int, error) {
	if pair == nil {
		pair = NewPair("")
	}

	for i := 1; i < len(data); {
		if unicode.IsSpace(data[i]) {
			i++
			continue
		}

		var k int
		var err error
		var value Value
		switch {
		case data[i] == ';':
			k, err = parseComment(data[i:])
		case data[i] == '"':
			value, k, err = parseString(data[i:])
		case data[i] == '(':
			value, k, err = parsePair(data[i:], nil)
		case data[i] == ')':
			return pair, i + 1, nil
		default:
			if pair.HasKey() {
				if data[i] == '.' || unicode.IsDigit(data[i]) {
					value, k, err = parseNumber(data[i:])
				} else if data[i] == 't' && len(data[i:]) >= 4 && string(data[i:i+4]) == "true" {
					value = NewBool(true)
					k = 4
				} else if data[i] == 'f' && len(data[i:]) >= 5 && string(data[i:i+5]) == "false" {
					value = NewBool(false)
					k = 5
				} else {
					return nil, 0, fmt.Errorf("parse pair failed, invalid format")
				}
			} else {
				var key string
				key, k, err = parsePairKey(data[i:])
				if err == nil {
					pair.SetKey(key)
				}
			}
		}

		if err != nil {
			return nil, 0, err
		}

		if value != nil {
			pair.Add(value)
		}

		i += k
	}
	return nil, 0, fmt.Errorf("parse pair failed, can't find end tag")
}

func parsePairKey(data []rune) (string, int, error) {
	for i := 0; i < len(data); i++ {
		switch {
		case unicode.IsSpace(data[i]),
			data[i] == ';',
			data[i] == '"',
			data[i] == '\'',
			data[i] == '(',
			data[i] == ')':
			return string(data[:i]), i, nil
		}
	}
	return "", 0, fmt.Errorf("parse pair key failed, invalid format")
}

func parseNumber(data []rune) (*Number, int, error) {
	var meetDot bool
	var i int
	var s string
L:
	for ; i < len(data); i++ {
		switch {
		case data[i] == '.':
			if meetDot {
				fmt.Println(string(data[:i]))
				fmt.Println(string(data[i:]))
				return nil, 0, fmt.Errorf("parse number failed, too many dot")
			}
			meetDot = true
		case data[i] == '\'':
			continue
		case unicode.IsDigit(data[i]):
			//
		default:
			break L
		}
		s += string(data[i])
	}

	if s == "." {
		return nil, 0, fmt.Errorf("parse number failed, invalid format")
	}

	return NewNumber(string(data[:i])), i, nil
}

func parseString(data []rune) (*String, int, error) {
	for i := 1; i < len(data); i++ {
		switch data[i] {
		case '\\':
			i++
		case '"':
			s := strings.Replace(string(data[1:i]), `\"`, `"`, -1)
			return NewString(s), i + 1, nil
		}
	}
	return nil, 0, fmt.Errorf("parse string failed, can't find end tag")
}
