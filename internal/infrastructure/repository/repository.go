package repository

import "time"

func NullableString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func DerefString(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}

func NullableTime(t time.Time) interface{} {
	if t.IsZero() {
		return nil
	}
	return t
}
