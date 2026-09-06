package user

import (
	"net/mail"
	"regexp"
	"time"
)

type Email string

func NewEmail(e string) (Email, error) {
	_, err := mail.ParseAddress(e)
	if err != nil {
		return "", ErrInvalidEmail
	}
	return Email(e), nil
}

func (e Email) String() string { return string(e) }

type DisplayName string

func NewDisplayName(s string) (DisplayName, error) {
	if len(s) == 0 || len(s) > 50 {
		return "", ErrInvalidDisplayName
	}
	return DisplayName(s), nil
}

func (d DisplayName) String() string { return string(d) }

type Username string

var usernameRegex = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func NewUsername(s string) (Username, error) {
	if s == "" {
		return "", nil // Optional field
	}
	if len(s) < 3 || len(s) > 30 {
		return "", ErrInvalidUsername
	}
	if !usernameRegex.MatchString(s) {
		return "", ErrInvalidUsername
	}
	return Username(s), nil
}

func (u Username) String() string { return string(u) }

type Bio string

func NewBio(s string) (Bio, error) {
	if len(s) > 500 {
		return "", ErrInvalidBio
	}
	return Bio(s), nil
}

func (b Bio) String() string { return string(b) }

type Headline string

func NewHeadline(s string) (Headline, error) {
	if len(s) > 100 {
		return "", ErrInvalidHeadline
	}
	return Headline(s), nil
}

func (h Headline) String() string { return string(h) }

type Pronouns string

func NewPronouns(s string) (Pronouns, error) {
	if len(s) > 20 {
		return "", ErrInvalidPronouns
	}
	return Pronouns(s), nil
}

func (p Pronouns) String() string { return string(p) }

type Location string

func NewLocation(s string) (Location, error) {
	if len(s) > 100 {
		return "", ErrInvalidLocation
	}
	return Location(s), nil
}

func (l Location) String() string { return string(l) }

type DateOfBirth time.Time

func NewDateOfBirth(t time.Time) (DateOfBirth, error) {
	if t.IsZero() {
		return DateOfBirth(time.Time{}), nil
	}
	// Minimum age 13 requirement
	thirteenYearsAgo := time.Now().AddDate(-13, 0, 0)
	if t.After(thirteenYearsAgo) {
		return DateOfBirth{}, ErrInvalidDateOfBirth
	}
	return DateOfBirth(t), nil
}

func (d DateOfBirth) Time() time.Time { return time.Time(d) }

type ImageStatus string

const (
	ImageStatusPending ImageStatus = "pending"
	ImageStatusReady   ImageStatus = "ready"
	ImageStatusError   ImageStatus = "error"
)

type ProfileImage struct {
	url    string
	status ImageStatus
}

func NewProfileImage(url string, status ImageStatus) ProfileImage {
	return ProfileImage{url: url, status: status}
}

func (p ProfileImage) URL() string         { return p.url }
func (p ProfileImage) Status() ImageStatus { return p.status }
