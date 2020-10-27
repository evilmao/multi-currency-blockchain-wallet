package storer

import (
	"fmt"
	"path/filepath"
	"strings"

	"upex-wallet/wallet-tools/cmd/genwallet/keypair"

	"upex-wallet/wallet-base/util"
)

type AddressFile struct {
	FileStorer
}

func NewAddressFile() *AddressFile {
	return &AddressFile{}
}

func (s *AddressFile) Append(pubKey keypair.PublicKey) error {
	if s.writer == nil {
		fileName := strings.ToLower(pubKey.Class()) + "-addrs.txt"
		fileName = filepath.Join(s.dataPath, fileName)
		w, err := OpenAddressFileWriter(fileName)
		if err != nil {
			return err
		}
		s.writer = w
	}
	return s.FileStorer.Append(pubKey)
}

type DepositAddressFile struct {
	*SectionFileStorer
	secCounter int
}

func NewDepositAddressFile(sections []*Section) *DepositAddressFile {
	return &DepositAddressFile{
		NewSectionFileStorer(sections),
		0,
	}
}

func (s *DepositAddressFile) Append(pubKey keypair.PublicKey) error {
	sec, isNew, err := s.Next()
	if err != nil {
		return err
	}

	if sec.IsSystem {
		return nil
	}

	if isNew {
		if s.writer != nil {
			s.writer.Close()
		}

		s.secCounter++
		var fileName string
		if len(sec.Tag) > 0 {
			fileName = fmt.Sprintf("%s-deposit-addrs-%d-%s.txt", strings.ToLower(pubKey.Class()), s.secCounter, sec.Tag)
		} else {
			fileName = fmt.Sprintf("%s-deposit-addrs-%d.txt", strings.ToLower(pubKey.Class()), s.secCounter)
		}

		fileName = filepath.Join(s.dataPath, fileName)
		w, err := OpenAddressFileWriter(fileName)
		if err != nil {
			return err
		}
		s.writer = w
	}
	return s.SectionFileStorer.Append(pubKey)
}

type AddressFileWriter struct {
	*util.FileWriter
}

func OpenAddressFileWriter(fileName string) (*AddressFileWriter, error) {
	w, err := OpenFileWriter(fileName)
	if err != nil {
		return nil, err
	}
	return &AddressFileWriter{w}, nil
}

func (w *AddressFileWriter) Write(pubKey keypair.PublicKey) error {
	_, err := w.WriteString(pubKey.AddressString() + "\n")
	if err != nil {
		return fmt.Errorf("store address %s into %s failed, %v",
			pubKey.AddressString(), w.FileName(), err)
	}
	return nil
}
