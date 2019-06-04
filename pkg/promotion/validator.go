package promotion

import (
	"errors"
	"time"
)

// validate exploit input functions to check and modify the content of Promotion
// It is used in Post and Put http method before using data layer to modify db
func (p *Promotion) validate(modifiers ...func(*Promotion) error) (err error) {

	for _, modifier := range modifiers {
		if err := modifier(p); err != nil {
			return err
		}
	}
	return nil
}

// validateNullBody checks if incoming request body is empty
func validateNullBody(p *Promotion) error {

	if *p == (Promotion{}) {
		return errors.New("Null Promotion Body")
	}
	return nil
}

func validateID(p *Promotion) error {

	if p.ID == 0 {
		return errors.New("Invalid Promotion ID")
	}
	return nil
}

// validateTitle checks if title is given
func validateTitle(p *Promotion) error {

	if p.Title == "" {
		return errors.New("Invalid Title")
	}
	return nil
}

// setCreatedAtNow set created_at to now
func setCreatedAtNow(p *Promotion) error {

	p.CreatedAt = time.Now()

	return nil
}

// setUpdatedAtNow set the updated_at to now
func setUpdatedAtNow(p *Promotion) error {

	p.UpdatedAt.Time = time.Now()
	p.UpdatedAt.Valid = true

	return nil
}
