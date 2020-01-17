package mysql

import (
	"testing"
	"time"

	"github.com/readr-media/readr-restful/internal/rrsql"
	"github.com/readr-media/readr-restful/pkg/subscription"
	"github.com/stretchr/testify/assert"
)

func TestGetStructTags(t *testing.T) {
	for _, tc := range []struct {
		name    string
		mode    string
		tagname string
		input   subscription.Subscription
		expect  []string
	}{
		{"all-tags", "full", "db", subscription.Subscription{}, []string{"id", "member_id", "email", "amount", "created_at", "updated_at", "last_paid_at", "payment_service", "invoice_service", "status", "payment_infos", "invoice_infos"}},
		{"partial-empty", "non-null", "db", subscription.Subscription{}, []string{}},
		{"built-in-types", "non-null", "db", subscription.Subscription{ID: 1, MemberID: rrsql.NullInt{Int: 648, Valid: true}, Amount: 100, PaymentService: "tappay", Status: 1}, []string{"id", "member_id", "amount", "payment_service", "status"}},
		{"NullTime", "non-null", "db", subscription.Subscription{CreatedAt: rrsql.NullTime{Time: time.Now(), Valid: true}}, []string{"created_at"}},
		{"partial-map", "non-null", "db", subscription.Subscription{PaymentInfos: map[string]interface{}{"foo": "bar"}}, []string{"payment_infos"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result := GetStructTags(tc.mode, tc.tagname, tc.input)
			assert.Equal(t, tc.expect, result)
		})
	}
}
