package timeutil

import "time"

const DefaultLayout = "2006-01-02 15:04:05"

func Now() time.Time {
	return time.Now()
}

func NowPtr() *time.Time {
	t := time.Now()
	return &t
}

func Parse(s string) (time.Time, error) {
	return time.Parse(DefaultLayout, s)
}

func Format(t time.Time) string {
	return t.Format(DefaultLayout)
}

func FormatPtr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(DefaultLayout)
}
