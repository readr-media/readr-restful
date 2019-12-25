package invoice

import "github.com/readr-media/readr-restful/pkg/invoice/ezpay"

// Provider is the interface each invoice service has to implement
type Provider interface {
	Create() (resp map[string]interface{}, err error)
	Validate() error
}

// NewInvoiceProvider returns a provider based on the service name
func NewInvoiceProvider(name string, data map[string]interface{}) (p Provider, err error) {
	switch name {
	case "ezpay":
		p = &ezpay.InvoiceClient{Payload: data}
		err = nil
	default:
		p = &ezpay.InvoiceClient{Payload: data}
		err = nil
	}
	return p, err
}
