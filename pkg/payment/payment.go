package payment

import "github.com/readr-media/readr-restful/pkg/payment/tappay"

type Provider interface {
	Pay(payment map[string]interface{}) (resp map[string]interface{}, token map[string]interface{}, err error)
}

func NewOnetimeProvider(name string) (p Provider, err error) {
	switch name {
	case "tappay":
		p = tappay.OnetimeClient{}
		err = nil
	default:
		p = tappay.OnetimeClient{}
		err = nil
	}
	return p, err
}

func NewRecurringProvider(name string) (p Provider, err error) {
	switch name {
	case "tappay":
		p = tappay.RecurringClient{}
		err = nil
	default:
		// default using tappay
		p = tappay.RecurringClient{}
		err = nil
	}
	return p, err
}
