package util

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
)

// ReadBytes reads n bytes from reader.
func ReadBytes(reader io.Reader, n int) ([]byte, error) {
	if reader == nil || n <= 0 {
		return nil, fmt.Errorf("invalid args (%v, %v)", reader, n)
	}

	buf := make([]byte, n)
	pos := 0
	for {
		m, err := reader.Read(buf[pos:])
		if err != nil {
			if err == io.EOF {
				pos += m
				if pos < n {
					return buf[:pos], err
				}
				return buf, nil
			}
			return nil, err
		}

		pos += m
		if pos >= n {
			return buf, nil
		}
	}
}

// ForeachLine reads string from reader line-by-line.
func ForeachLine(reader io.Reader, fun func(line string) error) error {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		err := fun(scanner.Text())
		if err != nil {
			return err
		}
	}
	return scanner.Err()
}

// Transformer transforms bytes.
type Transformer interface {
	Transform([]byte) ([]byte, error)
}

// CombineTransformer combines multi transformers.
type CombineTransformer []Transformer

func (ts CombineTransformer) Transform(data []byte) ([]byte, error) {
	var err error
	for i, t := range ts {
		if t == nil {
			continue
		}

		data, err = t.Transform(data)
		if err != nil {
			return nil, fmt.Errorf("perform transformer at index %d failed, %v", i, err)
		}
	}
	return data, nil
}

// The TransformFunc type is an adapter to allow the use of
// ordinary functions as Transformer.
type TransformFunc func([]byte) ([]byte, error)

func (f TransformFunc) Transform(data []byte) ([]byte, error) {
	return f(data)
}

// StringReplacer replaces string data.
type StringReplacer map[string]string

func (r StringReplacer) Transform(data []byte) ([]byte, error) {
	for old, new := range r {
		data = bytes.Replace(data, []byte(old), []byte(new), -1)
	}
	return data, nil
}
