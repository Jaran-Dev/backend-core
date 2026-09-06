package user

import (
	"time"

	"github.com/Jaran-Dev/backend-core/internal/domain/account"
)

type UpdateParams struct {
	DisplayName *string
	Username    *string
	Bio         *string
	Headline    *string
	Pronouns    *string
	Location    *string
	DateOfBirth *time.Time
}

type User struct {
	accountID   account.AccountID
	displayName string
	username    Username
	bio         Bio
	headline    Headline
	pronouns    Pronouns
	location    Location
	dateOfBirth DateOfBirth
	avatar      ProfileImage
	cover       ProfileImage
	createdAt   time.Time
	updatedAt   time.Time
	version     int
}

func NewUser(accID account.AccountID, displayName string) (*User, error) {
	dn, err := NewDisplayName(displayName)
	if err != nil {
		return nil, err
	}
	return &User{
		accountID:   accID,
		displayName: dn.String(),
		avatar:      NewProfileImage("", ImageStatusPending),
		cover:       NewProfileImage("", ImageStatusPending),
		createdAt:   time.Now(),
		updatedAt:   time.Now(),
		version:     1,
	}, nil
}

func (u *User) AccountID() account.AccountID { return u.accountID }
func (u *User) DisplayName() string          { return u.displayName }
func (u *User) Username() Username           { return u.username }
func (u *User) Bio() Bio                     { return u.bio }
func (u *User) Headline() Headline           { return u.headline }
func (u *User) Pronouns() Pronouns           { return u.pronouns }
func (u *User) Location() Location           { return u.location }
func (u *User) DateOfBirth() DateOfBirth     { return u.dateOfBirth }
func (u *User) Avatar() ProfileImage         { return u.avatar }
func (u *User) Cover() ProfileImage          { return u.cover }
func (u *User) CreatedAt() time.Time         { return u.createdAt }
func (u *User) UpdatedAt() time.Time         { return u.updatedAt }
func (u *User) Version() int                 { return u.version }

func (u *User) UpdateProfile(params UpdateParams) error {
	if params.DisplayName != nil {
		dn, err := NewDisplayName(*params.DisplayName)
		if err != nil {
			return err
		}
		u.displayName = dn.String()
	}

	if params.Username != nil {
		un, err := NewUsername(*params.Username)
		if err != nil {
			return err
		}
		u.username = un
	}

	if params.Bio != nil {
		b, err := NewBio(*params.Bio)
		if err != nil {
			return err
		}
		u.bio = b
	}

	if params.Headline != nil {
		h, err := NewHeadline(*params.Headline)
		if err != nil {
			return err
		}
		u.headline = h
	}

	if params.Pronouns != nil {
		p, err := NewPronouns(*params.Pronouns)
		if err != nil {
			return err
		}
		u.pronouns = p
	}

	if params.Location != nil {
		l, err := NewLocation(*params.Location)
		if err != nil {
			return err
		}
		u.location = l
	}

	if params.DateOfBirth != nil {
		d, err := NewDateOfBirth(*params.DateOfBirth)
		if err != nil {
			return err
		}
		u.dateOfBirth = d
	}

	u.updatedAt = time.Now()
	return nil
}

func (u *User) SetAvatar(url string, status ImageStatus) {
	u.avatar = NewProfileImage(url, status)
	u.updatedAt = time.Now()
}

func (u *User) SetCover(url string, status ImageStatus) {
	u.cover = NewProfileImage(url, status)
	u.updatedAt = time.Now()
}

func UnmarshalUser(
	accID account.AccountID,
	displayName string,
	username *string,
	bio *string,
	headline *string,
	pronouns *string,
	location *string,
	dob *time.Time,
	avatar, cover ProfileImage,
	createdAt, updatedAt time.Time,
	version int,
) (*User, error) {
	dn, err := NewDisplayName(displayName)
	if err != nil {
		return nil, err
	}

	var un Username
	if username != nil {
		un, err = NewUsername(*username)
		if err != nil {
			return nil, err
		}
	}

	var b Bio
	if bio != nil {
		b, err = NewBio(*bio)
		if err != nil {
			return nil, err
		}
	}

	var h Headline
	if headline != nil {
		h, err = NewHeadline(*headline)
		if err != nil {
			return nil, err
		}
	}

	var p Pronouns
	if pronouns != nil {
		p, err = NewPronouns(*pronouns)
		if err != nil {
			return nil, err
		}
	}

	var l Location
	if location != nil {
		l, err = NewLocation(*location)
		if err != nil {
			return nil, err
		}
	}

	var d DateOfBirth
	if dob != nil {
		d, err = NewDateOfBirth(*dob)
		if err != nil {
			return nil, err
		}
	}

	return &User{
		accountID:   accID,
		displayName: dn.String(),
		username:    un,
		bio:         b,
		headline:    h,
		pronouns:    p,
		location:    l,
		dateOfBirth: d,
		avatar:      avatar,
		cover:       cover,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
		version:     version,
	}, nil
}
