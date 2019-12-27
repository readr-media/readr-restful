package subscription

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/readr-media/readr-restful/internal/rrsql"
)

// PaymentStore wraps the necessary data structure for sqlx
// gin could bind map[string]interface{} directly, but not sqlx
type PaymentStore map[string]interface{}

// Value converts the value in PaymentStore to JSON and could be stored to db
func (p PaymentStore) Value() (driver.Value, error) {
	return json.Marshal(p)
}

// Scan the data from database and stores in PaymentStore
func (p *PaymentStore) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), p)
}

// InvoiceStore holds the infos for invoice
type InvoiceStore map[string]interface{}

// Value converts values in InvoiceStore to JSON and could be stored to db
func (i InvoiceStore) Value() (driver.Value, error) {
	return json.Marshal(i)
}

// Scan the data from database and stores in InvoiceStore
func (i *InvoiceStore) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), i)
}

// Subscriber provides the interface for different db backend
//go:generate mockgen -package=mock -destination=test/mock/mock.go github.com/readr-media/readr-restful/pkg/subscription Subscriber
type Subscriber interface {
	GetSubscriptions(f ListFilter) (results []Subscription, err error)
	CreateSubscription(s Subscription) error
	UpdateSubscriptions(s Subscription) error
	RoutinePay(subscribers []Subscription) error
}

// Subscription is the model for unmarshalling JSON, and serialized to database
type Subscription struct {
	ID             uint64         `json:"id,omitempty" db:"id"`                                 // Subscription id
	MemberID       rrsql.NullInt  `json:"member_id,omitempty" db:"member_id"`                   // Member who subscribed
	Email          string         `json:"email,omitempty" db:"email" binding:"required"`        // Email for failure handle, invoice create
	Amount         int            `json:"amount,omitempty" db:"amount" binding:"required,gt=0"` // Amount to pay
	CreatedAt      rrsql.NullTime `json:"created_at,omitempty" db:"created_at"`                 // The time when first created
	UpdatedAt      rrsql.NullTime `json:"updated_at,omitempty" db:"updated_at"`                 // The time when renewal
	LastPaidAt     rrsql.NullTime `json:"last_paid_at,omitempty" db:"last_paid_at"`             // Last time paid
	PaymentService string         `json:"payment_service,omitempty" db:"payment_service"`       // Payment service name
	InvoiceService string         `json:"invoice_service,omitempty" db:"invoice_service"`       // Invoice service name
	Status         int            `json:"status,omitempty" db:"status"`
	PaymentInfos   PaymentStore   `json:"payment_infos,omitempty" db:"payment_infos"`
	InvoiceInfos   InvoiceStore   `json:"invoice_infos,omitempty" db:"invoice_infos"`
}

type ListFilter interface {
	Select() (string, []interface{}, error)
}
