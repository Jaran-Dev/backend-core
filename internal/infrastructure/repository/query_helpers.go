package repository

import (
	"time"

	"gorm.io/gorm"
)

func ApplyPagination(query *gorm.DB, page, pageSize int) *gorm.DB {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize
	return query.Offset(offset).Limit(pageSize)
}

func DerefString(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}

func StringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func DerefTime(t *time.Time) time.Time {
	if t != nil {
		return *t
	}
	return time.Time{}
}

func TimePtr(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

func DerefInt(i *int) int {
	if i != nil {
		return *i
	}
	return 0
}

func IntPtr(i int) *int {
	if i == 0 {
		return nil
	}
	return &i
}

func DerefBool(b *bool) bool {
	if b != nil {
		return *b
	}
	return false
}

func BoolPtr(b bool) *bool {
	if !b {
		return nil
	}
	return &b
}
