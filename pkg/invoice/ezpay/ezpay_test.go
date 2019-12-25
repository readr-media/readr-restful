package ezpay

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidator(t *testing.T) {
	f, err := ioutil.ReadFile(filepath.Join("testdata", t.Name()+".golden"))
	if err != nil {
		t.Fatalf("could not read golden file %s ", t.Name())
	}
	var testdata = make([]map[string]interface{}, 0)
	if err = json.Unmarshal(f, &testdata); err != nil {
		t.Fatalf("%s unmarshal testdata with error:%v", t.Name(), err)
	}

	for _, tc := range testdata {

		t.Run(tc["name"].(string), func(t *testing.T) {

			ezpay := &InvoiceClient{Payload: tc["input"].(map[string]interface{})}
			if _, ok := ezpay.Payload["amount"]; ok {
				ezpay.Payload["amount"] = int(ezpay.Payload["amount"].(float64))
			}

			// Change default JSON float64 to int
			for _, key := range []string{
				"TotalAmt", "TaxRate",
			} {
				if value, ok := tc["expect"].(map[string]interface{})[key]; ok {
					tc["expect"].(map[string]interface{})[key] = int(value.(float64))
				}
			}

			err = ezpay.Validate()
			if err != nil {
				assert.Equal(t, tc["error"], err.Error(), "error message not equal")
			} else {
				// Only test if key TimeStamp and MerchantOrderNo exist, not the value
				if _, ok := ezpay.Payload["TimeStamp"]; !ok {
					t.Errorf("lack key:%s", "TimeStamp")
				} else {
					delete(ezpay.Payload, "TimeStamp")
				}
				if _, ok := ezpay.Payload["MerchantOrderNo"]; !ok {
					t.Errorf("lack key:%s", "MerchantOrderNo")
				} else {
					delete(ezpay.Payload, "MerchantOrderNo")
				}

				assert.Equal(t, tc["expect"], ezpay.Payload)
			}
		})
	}
}
