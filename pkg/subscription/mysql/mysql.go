package mysql

import (
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/readr-media/readr-restful/internal/rrsql"
	"github.com/readr-media/readr-restful/pkg/invoice"
	"github.com/readr-media/readr-restful/pkg/mail"
	"github.com/readr-media/readr-restful/pkg/payment"
	"github.com/readr-media/readr-restful/pkg/subscription"
)

func GetStructTags(mode string, tagname string, input interface{}, options ...interface{}) []string {

	columns := make([]string, 0)
	value := reflect.ValueOf(input)

	// Originally used to rule out id field when insert
	var skipFields []string
	if options != nil {
		skipFields = options[0].([]string)
	}

FindTags:
	for i := 0; i < value.NumField(); i++ {

		field := value.Type().Field(i)
		fieldType := field.Type
		fieldValue := value.Field(i)

		// Use Type() to get struct tags
		tag := value.Type().Field(i).Tag.Get(tagname)
		// Skip fields if there are denoted
		if len(skipFields) > 0 {

			for _, f := range skipFields {
				if tag == f {
					continue FindTags
				}
			}
		}

		if mode == "full" {
			columns = append(columns, tag)
		} else if mode == "non-null" {
			// Append each tag for non-null field
			switch fieldType.Name() {
			case "string":
				if fieldValue.String() != "" {
					columns = append(columns, tag)
				}
			case "int64", "int":
				if fieldValue.Int() != 0 {
					columns = append(columns, tag)
				}
			case "uint32", "uint64":
				if fieldValue.Uint() != 0 {
					columns = append(columns, tag)
				}
			case "NullString", "NullInt", "NullTime", "NullBool":
				if fieldValue.FieldByName("Valid").Bool() {
					columns = append(columns, tag)
				}
			case "bool":
				columns = append(columns, tag)
			default:
				fmt.Printf("unrecognised format: %s value:%v\n", value.Field(i).Type(), fieldValue)
				// TODO: restrict the judgement to certain kind with Kind(), or it might panic
				if fieldValue.Len() > 0 {
					columns = append(columns, tag)
				}
			}
		}
	}
	return columns
}

// SubscriptionService is the mysql version implementation of Subscriber interface.
type SubscriptionService struct {
	DB *sqlx.DB

	Payment     payment.Provider
	Invoice     invoice.Provider
	MailService mail.Mailer
}

// GetSubscriptions could list user subscriptions
func (s *SubscriptionService) GetSubscriptions(params subscription.ListFilter) (results []subscription.Subscription, err error) {

	query, values, err := params.Select()
	err = s.DB.Select(&results, query, values...)
	if err != nil {
		log.Printf("Failed to get subscriptions from MySQL: %s\n", err.Error())
		return nil, err
	}
	return results, nil
}

// CreateSubscription will make first payment, and store infos for recurring payment in db
func (s *SubscriptionService) CreateSubscription(p subscription.Subscription) (err error) {

	// Setup payment service
	s.Payment, err = payment.NewOnetimeProvider(p.PaymentService)

	// Setup invoice service
	p.InvoiceInfos["amount"] = p.Amount
	s.Invoice, err = invoice.NewInvoiceProvider(p.InvoiceService, p.InvoiceInfos)
	s.MailService = &mail.MailService{}

	if err != nil {
		log.Printf("recurring pay error assigning invoice service:%v\n", err)
		return err
	}
	if err = s.Invoice.Validate(); err != nil {
		log.Printf("invoice validated error:%s\n", err.Error())
		return err
	}

	// Set status to init
	p.Status = subscription.StatusInit
	// Create subscription records
	tags := GetStructTags("full", "db", p)
	query := fmt.Sprintf(`INSERT INTO subscriptions (%s) VALUES (:%s)`, strings.Join(tags, ","), strings.Join(tags, ",:"))

	result, err := s.DB.NamedExec(query, p)
	if err != nil {
		return err
	}
	subID, err := result.LastInsertId()
	if err != nil {
		log.Printf("Fail to get last inserted ID when insert a subscription: %v", err)
		return err
	}

	// Pay for the first time, set remember = true to get card_key & card_token
	p.PaymentInfos["amount"] = p.Amount
	p.PaymentInfos["remember"] = true

	_, p.PaymentInfos, err = s.Payment.Pay(p.PaymentInfos)
	if err != nil {
		s.MailService.Set(mail.Mail{To: []string{p.Email}, Subject: "Huston! We have a problem!", Body: err.Error()})
		if err = s.MailService.Send(); err != nil {
			return err
		}
		return err
	}

	_, err = s.Invoice.Create()
	if err != nil {
		log.Printf("error creating invoice:%v\n", err)
		s.MailService.Set(mail.Mail{To: []string{p.Email}, Subject: "Huston! We have a problem!", Body: err.Error()})
		if err = s.MailService.Send(); err != nil {
			return err
		}
		return err
	}

	// update payment token
	update := subscription.Subscription{ID: uint64(subID), Status: subscription.StatusOK, UpdatedAt: rrsql.NullTime{Time: time.Now(), Valid: true}, LastPaidAt: rrsql.NullTime{Time: time.Now(), Valid: true}, PaymentInfos: p.PaymentInfos}
	err = s.UpdateSubscriptions(update)

	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	// Create invoice
	return nil
}

// UpdateSubscriptions updates infos about a subsciption
func (s *SubscriptionService) UpdateSubscriptions(p subscription.Subscription) error {

	tags := GetStructTags("non-null", "db", p)
	fields := rrsql.MakeFieldString("update", `%s = :%s`, tags)
	query := fmt.Sprintf(`UPDATE subscriptions SET %s WHERE id = :id`, strings.Join(fields, ", "))

	_, err := s.DB.NamedExec(query, p)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	return nil
}

// RoutinePay accepts user subscription infos, request recurring pay API, and updates updated_at, last_pay_at.
func (s *SubscriptionService) RoutinePay(subscribers []subscription.Subscription) (err error) {

	for _, p := range subscribers {
		// Initializing payment service
		s.Payment, err = payment.NewRecurringProvider(p.PaymentService)
		if err != nil {
			return err
		}
		p.PaymentInfos["amount"] = p.Amount
		p.InvoiceInfos["amount"] = p.Amount

		s.Invoice, err = invoice.NewInvoiceProvider(p.InvoiceService, p.InvoiceInfos)
		if err != nil {
			log.Printf("recurring pay error assigning invoice service:%v\n", err)
			return err
		}
		if err = s.Invoice.Validate(); err != nil {
			return err
		}
		s.MailService = &mail.MailService{}

		_, _, err := s.Payment.Pay(p.PaymentInfos)
		if err != nil {
			log.Printf("recurring pay error paying:%v\n", err)
			s.MailService.Set(mail.Mail{To: []string{p.Email}, Subject: "Huston! We have a problem!", Body: err.Error()})
			if err = s.MailService.Send(); err != nil {
				return err
			}
			return err
		}
		update := subscription.Subscription{ID: uint64(p.ID), UpdatedAt: rrsql.NullTime{Time: time.Now(), Valid: true}, LastPaidAt: rrsql.NullTime{Time: time.Now(), Valid: true}}
		err = s.UpdateSubscriptions(update)
		if err != nil {
			return err
		}

		_, err = s.Invoice.Create()
		if err != nil {
			s.MailService.Set(mail.Mail{To: []string{p.Email}, Subject: "Huston! We have a problem!", Body: err.Error()})
			if err = s.MailService.Send(); err != nil {
				return err
			}
			log.Printf("recurring pay error creating invoice:%v\n", err)
			return err
		}

	}
	return nil
}
