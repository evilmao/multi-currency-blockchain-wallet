package storer

import (
	"fmt"
	"os"

	"upex-wallet/wallet-tools/cmd/genwallet/keypair"

	"upex-wallet/wallet-base/util"
)

type Writer interface {
	Write(keypair.PublicKey) error
	Close() error
}

type FileStorer struct {
	dataPath string
	writer   Writer
}

func (s *FileStorer) Open(dataPath string) {
	s.dataPath = dataPath
}

func (s *FileStorer) Append(pubKey keypair.PublicKey) error {
	if s.writer == nil {
		return fmt.Errorf("writer is nil")
	}

	return s.writer.Write(pubKey)
}

func (s *FileStorer) Close() error {
	return s.writer.Close()
}

type SectionFileStorer struct {
	FileStorer
	sections  Sections
	secIdx    int
	pubKeyIdx int
}

func NewSectionFileStorer(sections []*Section) *SectionFileStorer {
	return &SectionFileStorer{
		sections:  sections,
		secIdx:    -1,
		pubKeyIdx: -1,
	}
}

func (s *SectionFileStorer) Next() (*Section, bool, error) {
	s.pubKeyIdx++

	var newSection bool
	if s.secIdx < 0 || s.pubKeyIdx == s.sections[s.secIdx].End {
		newSection = true
		s.secIdx++
		if s.secIdx >= len(s.sections) {
			return nil, newSection, fmt.Errorf("pubKey at index %d in no section", s.pubKeyIdx)
		}
	}

	return s.sections[s.secIdx], newSection, nil
}

func OpenFileWriter(fileName string) (*util.FileWriter, error) {
	w := new(util.FileWriter)
	return w, w.Open(fileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
}
