package course

import (
	"time"

	"github.com/Jaran-Dev/backend-core/internal/domain/admin"

	"github.com/google/uuid"
)

type Course struct {
	id        CourseID
	slug      Slug
	status    CourseStatus
	coverURL  string
	createdBy admin.AdminID
	createdAt time.Time
	updatedAt time.Time
}

func NewCourse(slug Slug, coverURL string, creatorID admin.AdminID) *Course {
	now := time.Now()
	return &Course{
		id:        CourseID(uuid.New()),
		slug:      slug,
		status:    CourseStatusDraft,
		coverURL:  coverURL,
		createdBy: creatorID,
		createdAt: now,
		updatedAt: now,
	}
}

func (c *Course) ID() CourseID {
	return c.id
}

func (c *Course) Slug() Slug {
	return c.slug
}

func (c *Course) Status() CourseStatus {
	return c.status
}

func (c *Course) CoverURL() string {
	return c.coverURL
}

func (c *Course) CreatedBy() admin.AdminID {
	return c.createdBy
}

func (c *Course) CreatedAt() time.Time {
	return c.createdAt
}

func (c *Course) UpdatedAt() time.Time {
	return c.updatedAt
}

func (c *Course) Archive() {
	c.status = CourseStatusArchived
	c.updatedAt = time.Now()
}
