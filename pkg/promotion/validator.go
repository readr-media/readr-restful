package promotion

import (
	"errors"
	"time"
)

// Validate exploit input functions to check and modify the content of Promotion
// It is used in Post and Put http method before using data layer to modify db
func (p *Promotion) Validate(modifiers ...func(*Promotion) error) (err error) {

	for _, modifier := range modifiers {
		if err := modifier(p); err != nil {
			return err
		}
	}
	return nil
}

// ValidateNullBody checks if incoming request body is empty
func ValidateNullBody(p *Promotion) error {

	if *p == (Promotion{}) {
		return errors.New("null promotion payload")
	}
	return nil
}

func ValidateID(p *Promotion) error {

	if p.ID == 0 {
		return errors.New("invalid promotion id")
	}
	return nil
}

// ValidateTitle checks if title is given
func ValidateTitle(p *Promotion) error {

	if p.Title == "" {
		return errors.New("invalid title")
	}
	return nil
}

// SetCreatedAtNow set created_at to now
func SetCreatedAtNow(p *Promotion) error {

	if p.CreatedAt == (time.Time{}) {
		p.CreatedAt = time.Now()
	}

	return nil
}

// SetUpdatedAtNow set the updated_at to now
func SetUpdatedAtNow(p *Promotion) error {

	if !p.UpdatedAt.Valid {
		p.UpdatedAt.Time = time.Now()
		p.UpdatedAt.Valid = true
	}

	return nil
}
