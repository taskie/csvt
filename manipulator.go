package csvt

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
)

type Transposer struct {
	LengthChecked bool
}

func (transposer *Transposer) Transpose(records [][]string) ([][]string, error) {
	oldRowN := len(records)
	oldColumnN := -1
	for _, record := range records {
		columnN := len(record)
		if oldColumnN != columnN {
			if transposer.LengthChecked && oldColumnN != -1 {
				return nil, fmt.Errorf("the number of the columns is invalid")
			}
			if columnN > oldColumnN {
				oldColumnN = columnN
			}
		}
	}

	newRowN := oldColumnN
	newColumnN := oldRowN
	newRecords := make([][]string, newRowN)
	for i := 0; i < newRowN; i++ {
		newRecords[i] = make([]string, newColumnN)
		for j := 0; j < newColumnN; j++ {
			if j < len(records) {
				oldRecord := records[j]
				if i < len(oldRecord) {
					oldCell := records[j][i]
					newRecords[i][j] = oldCell
				}
			}
		}
	}
	return newRecords, nil
}

type Mapper struct {
	Header        []string
	LengthChecked bool
}

func (mapper *Mapper) Map(record []string) (item map[string]string, isBody bool, error error) {
	if mapper.Header == nil {
		mapper.Header = record
		return nil, false, nil
	}
	item = make(map[string]string)
	if mapper.LengthChecked && len(mapper.Header) != len(record) {
		return nil, true, fmt.Errorf("the number of the columns is invalid")
	}
	for i, title := range mapper.Header {
		if i < len(record) {
			item[title] = record[i]
		}
	}
	return item, true, nil
}

func (mapper *Mapper) MapAll(records [][]string) ([]map[string]string, error) {
	items := make([]map[string]string, 0)
	for _, record := range records {
		item, isBody, err := mapper.Map(record)
		if err != nil {
			return nil, err
		}
		if isBody {
			items = append(items, item)
		}
	}
	return items, nil
}

type Unmapper struct {
	Header     []string
	KeyChecked bool
}

func (unmapper *Unmapper) MakeHeader(items []map[string]string) ([]string, error) {
	first := true
	keySet := make(map[string]struct{})
	for _, item := range items {
		for k := range item {
			if _, ok := keySet[k]; !ok {
				if unmapper.KeyChecked && !first {
					return nil, fmt.Errorf("invalid key: %s", k)
				}
				keySet[k] = struct{}{}
			}
		}
		first = false
	}
	keys := make([]string, 0)
	for k := range keySet {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys, nil
}

func (unmapper *Unmapper) PrepareHeader(items []map[string]string) error {
	header, err := unmapper.MakeHeader(items)
	if err != nil {
		return err
	}
	unmapper.Header = header
	return nil
}

func (unmapper *Unmapper) Unmap(item map[string]string) ([]string, error) {
	if unmapper.Header == nil {
		err := unmapper.PrepareHeader([]map[string]string{item})
		if err != nil {
			return nil, err
		}
	}
	if unmapper.KeyChecked && len(unmapper.Header) != len(item) {
		for k := range item {
			found := false
			for _, hk := range unmapper.Header {
				if k == hk {
					found = true
					break
				}
			}
			if !found {
				return nil, fmt.Errorf("invalid key: %s", k)
			}
		}
	}
	record := make([]string, 0)
	for _, k := range unmapper.Header {
		if v, ok := item[k]; ok {
			record = append(record, v)
		} else {
			if unmapper.KeyChecked {
				return nil, fmt.Errorf("entry is not found: %s", k)
			}
			record = append(record, "")
		}
	}
	return record, nil
}

func (unmapper *Unmapper) UnmapAll(items []map[string]string) ([][]string, error) {
	if unmapper.Header == nil {
		err := unmapper.PrepareHeader(items)
		if err != nil {
			return nil, err
		}
	}
	records := make([][]string, 1)
	records[0] = make([]string, len(unmapper.Header))
	for i, cell := range unmapper.Header {
		records[0][i] = cell
	}
	for _, item := range items {
		record, err := unmapper.Unmap(item)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}

type Range struct {
	Begin int
	End   int
}

func (r Range) Size() int {
	return r.End - r.Begin
}

func (r Range) Includes(i int) bool {
	return r.Begin <= i && i < r.End
}

func NewRange(begin, end int) Range {
	return Range{Begin: begin, End: end}
}

func NewRangeTo(end int) Range {
	return NewRange(0, end)
}

func NewRangeFrom(begin int) Range {
	return NewRange(begin, math.MaxInt32)
}

func normalizeRanges(rs Ranges, max int) error {
	for i, r := range rs {
		if r.Begin < 0 {
			if r.End == 0 {
				r.End = max // workaround
			}
			r.Begin = max + r.Begin
		}
		if r.End < 0 {
			r.End = max + r.End
		}
		rs[i] = r
	}
	sort.SliceStable(rs, func(i, j int) bool {
		return rs[i].End > rs[j].End
	})
	sort.SliceStable(rs, func(i, j int) bool {
		return rs[i].Begin < rs[j].Begin
	})
	normalized := make([]Range, 0)
	prev := NewRange(0, 0)
	for _, r := range rs {
		if r.Begin > r.End {
			return fmt.Errorf("invalid range: %d:%d", r.Begin, r.End)
		}
		if prev.End < r.Begin {
			if prev.Size() > 0 {
				normalized = append(normalized, prev)
			}
			prev = r
		} else {
			if prev.End < r.End {
				prev.End = r.End
			}
		}
	}
	if prev.Size() > 0 {
		normalized = append(normalized, prev)
	}
	copy(rs, normalized)
	return nil
}

func ParseRanges(s string) ([]Range, error) {
	if s == "" {
		return nil, nil
	}
	rs := make([]Range, 0)
	for _, sr := range strings.Split(s, ",") {
		sr = strings.Trim(sr, " ")
		scs := strings.SplitN(sr, ":", 2)
		if len(scs) == 1 {
			i, err := strconv.Atoi(scs[0])
			if err != nil {
				return nil, err
			}
			rs = append(rs, NewRange(i, i+1))
		} else {
			var i, j int
			var err error
			if scs[0] != "" {
				i, err = strconv.Atoi(scs[0])
				if err != nil {
					return nil, err
				}
			} else {
				i = 0
			}
			if scs[1] != "" {
				j, err = strconv.Atoi(scs[1])
				if err != nil {
					return nil, err
				}
			} else {
				j = math.MaxInt32
			}
			rs = append(rs, NewRange(i, j))
		}
	}
	return rs, nil
}

type Ranges []Range

func (rs Ranges) Includes(i int) bool {
	for _, r := range rs {
		if r.Includes(i) {
			return true
		}
	}
	return false
}

type Slicer struct {
	RowRanges Ranges
	ColRanges Ranges
}

func (s *Slicer) Slice(records [][]string) ([][]string, error) {
	err := normalizeRanges(s.RowRanges, len(records))
	if err != nil {
		return nil, err
	}
	maxCols := -1
	for _, record := range records {
		cols := len(record)
		if maxCols < cols {
			maxCols = cols
		}
	}
	err = normalizeRanges(s.ColRanges, maxCols)
	if err != nil {
		return nil, err
	}
	newRecords := make([][]string, 0)
	for i, record := range records {
		if s.RowRanges != nil && !s.RowRanges.Includes(i) {
			continue
		}
		newRecord := make([]string, 0)
		for j, cell := range record {
			if s.ColRanges != nil && !s.ColRanges.Includes(j) {
				continue
			}
			newRecord = append(newRecord, cell)
		}
		newRecords = append(newRecords, newRecord)
	}
	return newRecords, nil
}
