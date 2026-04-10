package domain

import "time"

func ParseUserTime(value string, loc *time.Location) (time.Time, error) {
	if parsed, err := time.ParseInLocation("2006-01-02", value, loc); err == nil {
		return parsed, nil
	}
	return time.Parse(time.RFC3339, value)
}
