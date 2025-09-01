package prompt

import (
	"fmt"
	"time"
)

func ParseDate(input string) (string, error) {
	today := time.Now()
	var d time.Time

	switch input {
	case ".":
		d = today
	case ".y":
		d = today.AddDate(0, 0, -1)
	case ".t":
		d = today.AddDate(0, 0, 1)
	default:
		parsed, err := time.Parse("2006-01-02", input)
		if err != nil {
			return "", fmt.Errorf("invalid date format, please use YYYY-MM-DD or . for today, .y for yesterday, .t for tomorrow")
		}
		d = parsed
	}
	return d.Format("2006-01-02"), nil
}