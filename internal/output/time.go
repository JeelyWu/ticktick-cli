package output

import "time"

func FormatTime(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.Local().Format("2006-01-02 15:04")
}
