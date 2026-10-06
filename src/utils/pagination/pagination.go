package pagination

import (
	"fmt"
	"strconv"
)

func Page(text string, size int) (int, error) {
	page, err := strconv.Atoi(text)
	maxInt := int(^uint(0) >> 1)
	if err != nil || page < 1 || size < 1 || page-1 > maxInt/size {
		return 0, fmt.Errorf("invalid page")
	}
	return page, nil
}

func Size(text string, maximum int) (int, error) {
	size, err := strconv.Atoi(text)
	if err != nil || size < 1 || size > maximum {
		return 0, fmt.Errorf("page size must be between 1 and %d", maximum)
	}
	return size, nil
}
