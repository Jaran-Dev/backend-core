package shared

type Locale string

const (
	LocaleEnglish         Locale = "en"
	LocaleSoraniKurdish   Locale = "ckb"
	LocaleKurmanjiKurdish Locale = "kmr"
)

func (l Locale) IsValid() bool {
	switch l {
	case LocaleEnglish, LocaleSoraniKurdish, LocaleKurmanjiKurdish:
		return true
	}
	return false
}

func NewLocale(s string) (Locale, error) {
	l := Locale(s)
	if !l.IsValid() {
		return "", ErrInvalidLocale
	}
	return l, nil
}
