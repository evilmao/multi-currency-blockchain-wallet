package lisc

import (
	"fmt"
	"io/ioutil"
)

type Config struct {
	*Pair
}

func New() *Config {
	return &Config{
		Pair: NewPair("root"),
	}
}

func (c *Config) Load(filePath string) error {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read config file failed, %v", err)
	}

	err = c.Parse(string(data))
	if err != nil {
		return err
	}

	return nil
}

func (c *Config) Parse(raw string) error {
	data := []rune("(" + raw + ")")
	_, _, err := parsePair(data, c.Pair)
	if err != nil {
		return err
	}
	return nil
}

func (c *Config) Format() string {
	var s string
	vn := c.ValueCount()
	for i, v := range c.values {
		s += v.Format(0)
		if i < vn-1 {
			s += "\n"
		}
	}
	return s
}
