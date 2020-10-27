package main

import (
	"bufio"
	"fmt"
	"path/filepath"
	"strconv"
	"time"

	"upex-wallet/wallet-base/util"

	"upex-wallet/wallet-base/newbitx/lisc"

	"upex-wallet/wallet-tools/cmd/genwallet/keypair/storer"
)

const (
	normalSectionKey = "normal-section"
	systemSectionKey = "system-section"
	timeLayoutFormat = "2006-01-02 15:04:05"
)

type Meta struct {
	dataPath       string
	inputFileName  string
	outputFileName string
	sections       storer.Sections // section 数组
}

// 构造meta结构体
// sections 默认是个空数组
func NewMeta(dataPath, inputFileName, outputFileName string) *Meta {
	return &Meta{
		dataPath:       dataPath,
		inputFileName:  inputFileName + ".meta",
		outputFileName: outputFileName + ".meta",
	}
}

func (m *Meta) Load() error {
	cfg := lisc.New()
	err := cfg.Load(filepath.Join(m.dataPath, m.inputFileName))
	if err != nil {
		return err
	}

	m.sections = nil
	for i := 0; i < cfg.ValueCount(); i++ {
		v, _ := cfg.Value(i)
		if v.Type() != lisc.PairType {
			return fmt.Errorf("invalid meta file format")
		}

		pair := v.(*lisc.Pair)
		if pair.Key() != normalSectionKey && pair.Key() != systemSectionKey {
			return fmt.Errorf("invalid meta file format")
		}

		if pair.ValueCount() < 2 {
			return fmt.Errorf("invalid meta file format")
		}

		start, err := pair.Int64(0, 0)
		if err != nil {
			return fmt.Errorf("read section start failed, %v", err)
		}

		end, err := pair.Int64(0, 1)
		if err != nil {
			return fmt.Errorf("read section end failed, %v", err)
		}

		if start < 0 || end <= start {
			return fmt.Errorf("invalid section start/end, (%d, %d)", start, end)
		}

		var (
			tm  = time.Time{}
			tag string
		)
		if pair.ValueCount() >= 3 {
			ts, err := pair.String("", 2)
			if err != nil {
				return fmt.Errorf("read section time failed, %v", err)
			}

			tm, err = time.Parse(timeLayoutFormat, ts)
			if err != nil {
				return fmt.Errorf("parse section time failed, %v", err)
			}

			if pair.ValueCount() >= 4 {
				tag, err = pair.String("", 3)
				if err != nil {
					return fmt.Errorf("read section tag failed, %v", err)
				}
			}
		}

		m.Add(&storer.Section{
			Start:    int(start),
			End:      int(end),
			Time:     tm,
			IsSystem: pair.Key() == systemSectionKey,
			Tag:      tag,
		})
	}

	return m.sections.Verify()
}

func (m *Meta) Add(systemSection *storer.Section) {
	m.sections = append(m.sections, systemSection)
}

func (m *Meta) Store() error {
	cfg := lisc.New()
	for _, sec := range m.sections {
		sectionKey := normalSectionKey
		if sec.IsSystem {
			sectionKey = systemSectionKey
		}

		cfg.Add(
			lisc.NewPair(sectionKey).
				AddNumber(strconv.Itoa(sec.Start)).
				AddNumber(strconv.Itoa(sec.End)).
				AddString(sec.Time.Format(timeLayoutFormat)).
				AddString(sec.Tag))
	}

	return util.WithWriteFile(filepath.Join(m.dataPath, m.outputFileName), func(w *bufio.Writer) error {
		_, err := w.WriteString(cfg.Format())
		return err
	})
}

func (m *Meta) Sections() storer.Sections {
	return m.sections
}
