package storer

import (
	"fmt"
	"sort"
	"time"
)

// Section represents a section of [start, end).
type Section struct {
	Start    int
	End      int
	Time     time.Time
	IsSystem bool
	Tag      string
}

func (s *Section) Verify() error {
	if s.Start >= s.End {
		return fmt.Errorf("invalid section: %s", s.RangeString())
	}
	return nil
}

func (s *Section) RangeString() string {
	return fmt.Sprintf("[%d, %d)", s.Start, s.End)
}

type Sections []*Section

func (secs Sections) Verify() error {
	if len(secs) == 0 {
		return nil
	}

	sort.Slice(secs, func(i, j int) bool {
		return secs[i].Start < secs[j].Start
	})

	for i, s := range secs {
		if err := s.Verify(); err != nil {
			return err
		}

		if i > 0 {
			if s.Start != secs[i-1].End {
				return fmt.Errorf("invalid nearby sections: %s, %s",
					secs[i-1].RangeString(), s.RangeString())
			}
		}
	}

	return nil
}
