package category

import "github.com/Jaran-Dev/backend-core/internal/domain/shared"

type CategoryTranslation struct {
	categoryID  CategoryID
	locale      shared.Locale
	title       string
	description string
}

func NewCategoryTranslation(categoryID CategoryID, locale shared.Locale, title, description string) CategoryTranslation {
	return CategoryTranslation{
		categoryID:  categoryID,
		locale:      locale,
		title:       title,
		description: description,
	}
}

func (t *CategoryTranslation) CategoryID() CategoryID { return t.categoryID }
func (t *CategoryTranslation) Locale() shared.Locale  { return t.locale }
func (t *CategoryTranslation) Title() string          { return t.title }
func (t *CategoryTranslation) Description() string    { return t.description }

func (t *CategoryTranslation) UpdateTitle(title string)      { t.title = title }
func (t *CategoryTranslation) UpdateDescription(desc string) { t.description = desc }

func UnmarshalCategoryTranslation(categoryID CategoryID, locale shared.Locale, title, description string) (CategoryTranslation, error) {
	if !locale.IsValid() {
		return CategoryTranslation{}, ErrInvalidLocale
	}
	return CategoryTranslation{
		categoryID:  categoryID,
		locale:      locale,
		title:       title,
		description: description,
	}, nil
}
