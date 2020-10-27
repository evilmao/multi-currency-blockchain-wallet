package storer

import (
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strings"

	"upex-wallet/wallet-tools/cmd/genwallet/keypair"

	"upex-wallet/wallet-base/util"
)

const (
	addressSQLHeaderFormat = "SET NAMES utf8 ;\n" +
		"SET character_set_client = utf8 ;\n" +
		"CREATE TABLE IF NOT EXISTS `address` (\n" +
		"`id` int(10) unsigned NOT NULL AUTO_INCREMENT,\n" +
		"`address` varchar(%d) DEFAULT NULL,\n" +
		"`type` tinyint(4) DEFAULT NULL,\n" +
		"`version` varchar(8) DEFAULT NULL,\n" +
		"`pub_key` varchar(512) DEFAULT NULL,\n" +
		"PRIMARY KEY (`id`),\n" +
		"UNIQUE KEY `addr_ver` (`address`,`version`),\n" +
		"KEY `idx_address_address` (`address`)\n" +
		") ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8;\n\n"
)

func AddressSQLHeader(addrLen int) string {
	return fmt.Sprintf(addressSQLHeaderFormat, addrLen+10)
}

type DepositAddress struct {
	*SectionFileStorer
	secCounter int
}

func NewDepositAddress(sections []*Section) *DepositAddress {
	return &DepositAddress{
		NewSectionFileStorer(sections),
		0,
	}
}

func (s *DepositAddress) Append(pubKey keypair.PublicKey) error {
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
			fileName = fmt.Sprintf("%s-deposit-address-%d-%s.sql", strings.ToLower(pubKey.Class()), s.secCounter, sec.Tag)
		} else {
			fileName = fmt.Sprintf("%s-deposit-address-%d.sql", strings.ToLower(pubKey.Class()), s.secCounter)
		}
		fileName = filepath.Join(s.dataPath, fileName)
		w, err := OpenDepositAddressSQLWriter(fileName)
		if err != nil {
			return err
		}
		s.writer = w
	}
	return s.SectionFileStorer.Append(pubKey)
}

type DepositAddressSQLWriter struct {
	*util.FileWriter
	counter int
	head    string
	sql     string
}

func OpenDepositAddressSQLWriter(fileName string) (*DepositAddressSQLWriter, error) {
	w, err := OpenFileWriter(fileName)
	if err != nil {
		return nil, err
	}
	return &DepositAddressSQLWriter{
		FileWriter: w,
		counter:    0,
		head:       "INSERT INTO address (address, type, version) VALUES ",
		sql:        "",
	}, nil
}

func (w *DepositAddressSQLWriter) Write(pubKey keypair.PublicKey) error {
	if w.counter == 0 {
		addressSQLHeader := AddressSQLHeader(len(pubKey.AddressString()))
		_, err := w.WriteString(addressSQLHeader)
		if err != nil {
			return fmt.Errorf("store deposit address sql header failed, %v", err)
		}

		w.sql += w.head
	} else {
		if w.counter%100 == 0 {
			w.sql += ";\n" + w.head
		} else {
			w.sql += ",\n"
		}
	}

	w.sql += fmt.Sprintf("('%s', 1, 1)", pubKey.AddressString())

	if w.counter > 0 && w.counter%100 == 0 {
		_, err := w.WriteString(w.sql)
		if err != nil {
			return fmt.Errorf("store deposit address sql into %s failed, %v", w.FileName(), err)
		}

		w.sql = ""
	}

	w.counter++
	return nil
}

func (w *DepositAddressSQLWriter) Close() error {
	if len(w.sql) > 0 {
		w.sql += ";\n"
		_, err := w.WriteString(w.sql)
		if err != nil {
			return fmt.Errorf("store deposit address sql at end failed, %v", err)
		}

		w.sql = ""
	}
	return w.FileWriter.Close()
}

type WithdrawAddress struct {
	*SectionFileStorer
}

func NewWithdrawAddress(sections []*Section) *WithdrawAddress {
	return &WithdrawAddress{
		NewSectionFileStorer(sections),
	}
}

func (s *WithdrawAddress) Append(pubKey keypair.PublicKey) error {
	sec, isNew, err := s.Next()
	if err != nil {
		return err
	}

	if isNew {
		if s.writer != nil {
			s.writer.Close()
		}

		var fileName string
		if len(sec.Tag) > 0 {
			fileName = fmt.Sprintf("%s-withdraw-address-%d-%s.sql", strings.ToLower(pubKey.Class()), s.secIdx+1, sec.Tag)
		} else {
			fileName = fmt.Sprintf("%s-withdraw-address-%d.sql", strings.ToLower(pubKey.Class()), s.secIdx+1)
		}
		fileName = filepath.Join(s.dataPath, fileName)
		w, err := OpenWithdrawAddressSQLWriter(fileName, sec.IsSystem)
		if err != nil {
			return err
		}
		s.writer = w
	}
	return s.SectionFileStorer.Append(pubKey)
}

type WithdrawAddressSQLWriter struct {
	*util.FileWriter
	isSystem bool
	counter  int
	head     string
	sql      string
}

func OpenWithdrawAddressSQLWriter(fileName string, isSystem bool) (*WithdrawAddressSQLWriter, error) {
	w, err := OpenFileWriter(fileName)
	if err != nil {
		return nil, err
	}
	return &WithdrawAddressSQLWriter{
		FileWriter: w,
		isSystem:   isSystem,
		counter:    0,
		head:       "INSERT INTO address (address, pub_key, type, version) VALUES ",
		sql:        "",
	}, nil
}

func (w *WithdrawAddressSQLWriter) Write(pubKey keypair.PublicKey) error {
	if w.counter == 0 {
		addressSQLHeader := AddressSQLHeader(len(pubKey.AddressString()))
		_, err := w.WriteString(addressSQLHeader)
		if err != nil {
			return fmt.Errorf("store withdraw address sql header failed, %v", err)
		}

		w.sql += w.head
	} else {
		if w.counter%100 == 0 {
			w.sql += ";\n" + w.head
		} else {
			w.sql += ",\n"
		}
	}

	if w.isSystem {
		w.sql += fmt.Sprintf("('%s', '%s', 0, 1)", pubKey.AddressString(), hex.EncodeToString(pubKey.PublicKey()))
	} else {
		w.sql += fmt.Sprintf("('%s', '%s', 1, 1)", pubKey.AddressString(), hex.EncodeToString(pubKey.PublicKey()))
	}

	if w.counter > 0 && w.counter%100 == 0 {
		_, err := w.WriteString(w.sql)
		if err != nil {
			return fmt.Errorf("store withdraw address sql into %s failed, %v", w.FileName(), err)
		}

		w.sql = ""
	}

	w.counter++
	return nil
}

func (w *WithdrawAddressSQLWriter) Close() error {
	if len(w.sql) > 0 {
		w.sql += ";\n"
		_, err := w.WriteString(w.sql)
		if err != nil {
			return fmt.Errorf("store withdraw address sql at end failed, %v", err)
		}

		w.sql = ""
	}
	return w.FileWriter.Close()
}
