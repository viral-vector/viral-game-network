package pagination

import (
	"strconv"
	"testing"
)

func TestPaginationBoundsAndOffsetOverflow(t *testing.T) {
	for _, text := range []string{"-1", "0", "abc", "101"} {
		if _, err := Size(text, 100); err == nil {
			t.Errorf("invalid size accepted: %s", text)
		}
	}
	for _, text := range []string{"-1", "0", "abc", strconv.FormatUint(uint64(^uint(0)>>1), 10)} {
		if _, err := Page(text, 100); err == nil {
			t.Errorf("invalid page accepted: %s", text)
		}
	}
	if page, err := Page("3", 15); err != nil || page != 3 {
		t.Fatal("valid page rejected", page, err)
	}
	if size, err := Size("100", 100); err != nil || size != 100 {
		t.Fatal("maximum page size rejected", size, err)
	}
	if _, err := Page("1", 0); err == nil {
		t.Fatal("zero page size accepted")
	}
}
