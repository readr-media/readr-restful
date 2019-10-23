package poll

import (
	"time"

	"github.com/readr-media/readr-restful/internal/rrsql"
)

// Validate exploit functions to check and modify the content of Poll
func (p *Poll) Validate(modifiers ...func(*Poll)) {

	for _, modifier := range modifiers {
		modifier(p)
	}
}

// ValidatePollInsertID remove the id field if necessary
func ValidatePollInsertID(p *Poll) {
	if p.ID != 0 {
		p.ID = 0
	}
}

// ValidatePollCreatedAt set created_at of Poll to current time
func ValidatePollCreatedAt(p *Poll) {
	if !p.CreatedAt.Valid {
		p.CreatedAt = rrsql.NullTime{Time: time.Now(), Valid: true}
	}
}

// ValidatePollUpdatedAt set updated_at of Poll to now
func ValidatePollUpdatedAt(p *Poll) {
	if p.UpdatedAt.Valid {
		p.UpdatedAt = rrsql.NullTime{Time: time.Now(), Valid: true}
	}
}

// ValidateChoiceInsertID remove the id field if necessary
func ValidateChoiceInsertID(c *Choice) {
	// If assign insert will cause repeat primary key error in MySQL
	// Set it to 0 to enable auto_increment
	if c.ID != 0 {
		c.ID = 0
	}
}

// ValidateChoiceCreatedAt set created_at of Choice to current time
func ValidateChoiceCreatedAt(c *Choice) {

	if !c.CreatedAt.Valid {
		c.CreatedAt = rrsql.NullTime{Time: time.Now(), Valid: true}
	}
}

// ValidateChoiceUpdatedAt set updated_at of Choice to now
func ValidateChoiceUpdatedAt(c *Choice) {
	if !c.UpdatedAt.Valid {
		c.UpdatedAt = rrsql.NullTime{Time: time.Now(), Valid: true}
	}
}

// ValidateChoicePollID check the poll_id field, and set it to pollID
// if PollID is not valid, or PollID is valid but not equal to pollID
func ValidateChoicePollID(pollID int64) func(c *Choice) {
	return func(c *Choice) {
		if !c.PollID.Valid || (c.PollID.Valid && c.PollID.Int != pollID) {
			c.PollID.Int = pollID
			c.PollID.Valid = true
		}
	}
}

// Validate takes a series modifiers to check a single Choice
func (c *Choice) Validate(modifiers ...func(*Choice)) {

	for _, modifier := range modifiers {
		modifier(c)
	}
}
