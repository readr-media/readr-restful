package ezpay

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/readr-media/readr-restful/config"
)

const DefaultCommentLength = 71

// InvoiceClient holds the infos to create invoice with ezPay
type InvoiceClient struct {
	Payload map[string]interface{}
}

// PKCS7Padding will add paddings to input bytearray
func PKCS7Padding(b []byte, blocksize int) ([]byte, error) {

	if blocksize <= 0 {
		return nil, errors.New("invalid blocksize")
	}
	if b == nil || len(b) == 0 {
		return nil, errors.New("invalid PKCS7 data (empty or not padded)")
	}
	n := blocksize - (len(b) % blocksize)
	pb := make([]byte, len(b)+n)
	copy(pb, b)
	copy(pb[len(b):], bytes.Repeat([]byte{byte(n)}, n))
	return pb, nil
}

// Create makes an invoice API call to ezpay
func (c *InvoiceClient) Create() (resp map[string]interface{}, err error) {

	dataURL := url.Values{}
	for k, v := range c.Payload {
		dataURL.Set(k, fmt.Sprintf("%v", v))
	}
	postdata := []byte(dataURL.Encode())
	key := []byte(config.Config.InvoiceService.Key)
	iv := []byte(config.Config.InvoiceService.IV)

	// encrypt PostData_ first with AES-CBC
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("error creating new cipher for %s when create invoice:%v", key, err.Error())
	}
	// Add PKCS7Padding
	data, err := PKCS7Padding(postdata, aes.BlockSize)
	if err != nil {
		return nil, err
	}
	encrypter := cipher.NewCBCEncrypter(block, iv)
	encrypter.CryptBlocks(data, data)

	postURL := url.Values{}
	postURL.Set("MerchantID_", config.Config.InvoiceService.MerchantID)
	postURL.Set("PostData_", hex.EncodeToString(data))

	req, err := http.NewRequest("POST", config.Config.InvoiceService.URL, bytes.NewBufferString(postURL.Encode()))
	if err != nil {
		return nil, fmt.Errorf("creating http request for ezPay error:%s", err.Error())
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{}
	r, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("requesting ezPay API error: %s", err.Error())
	}
	defer r.Body.Close()
	// Parse response
	respBody, _ := ioutil.ReadAll(r.Body)
	err = json.Unmarshal(respBody, &resp)
	if err != nil {
		return nil, fmt.Errorf("parsing response from ezPay error:%s", err.Error())
	}

	if status, ok := resp["Status"]; ok && status != "SUCCESS" {
		return nil, fmt.Errorf("create invoice error:%s", resp["Message"])
	}
	return resp, nil
}

func get(target map[string]interface{}, key string, defaultValue interface{}) (result interface{}) {
	if value, ok := target[key]; ok {
		switch value.(type) {
		case string:
			if value.(string) != "" {
				return value
			}
		}
		return value
	}
	return defaultValue
}

// iinterfaceSliceToa converts a []interface{} containing strings to a pure []string
func interfaceSliceToa(i []interface{}) (result []string) {
	for _, v := range i {
		result = append(result, v.(string))
	}
	return result
}

// interfaceSliceItoa converts each integer in []interface{} to a and put them into a []string
func interfaceSliceItoa(i []interface{}) (result []string) {
	for _, v := range i {
		result = append(result, strconv.Itoa(int(v.(float64))))
	}
	return result
}

// Validate check the data for InvoiceClient, fix missing fields
func (c *InvoiceClient) Validate() (err error) {

	var result = make(map[string]interface{}, 0)

	result["RespondType"] = get(c.Payload, "response_type", "JSON")
	result["TimeStamp"] = time.Now().Unix()
	result["MerchantOrderNo"] = time.Now().Format("20060102")
	result["Status"] = get(c.Payload, "status", "0")
	result["TaxType"] = get(c.Payload, "tax_type", "1")
	result["Category"] = get(c.Payload, "category", "B2C")
	result["LoveCode"] = get(c.Payload, "love_code", "")
	result["CarrierType"] = get(c.Payload, "carrier_type", "")
	result["CarrierNum"] = get(c.Payload, "carrier_num", "")
	result["BuyerName"] = get(c.Payload, "buyer_name", "")
	result["BuyerEmail"] = get(c.Payload, "buyer_email", "")

	result["ItemName"] = get(c.Payload, "item_name", []interface{}{})
	result["ItemCount"] = get(c.Payload, "item_count", []interface{}{})
	result["ItemPrice"] = get(c.Payload, "item_price", []interface{}{})
	result["ItemUnit"] = get(c.Payload, "item_unit", []interface{}{})

	result["TotalAmt"] = get(c.Payload, "amount", nil)
	if result["TotalAmt"] == nil {
		return errors.New("invalid amount")
	}

	if config.Config.InvoiceService.APIVersion == "" {
		result["Version"] = "1.4"
	} else {
		result["Version"] = config.Config.InvoiceService.APIVersion
	}

	switch result["TaxType"].(string) {
	case "2":
		result["TaxRate"] = 0
		result["TaxAmt"] = 0
		result["Amt"] = result["TotalAmt"]
		result["CustomsClearance"] = "1"
	case "3":
		result["TaxRate"] = 0
		result["TaxAmt"] = 0
		result["Amt"] = result["TotalAmt"]
	case "9":
		// TODO: validation
		// Temporarily fallthrough
		fallthrough
	case "1":
		fallthrough
	default:
		result["TaxRate"] = 5
		result["TaxAmt"] = math.Round(float64(result["TotalAmt"].(int)) * (float64(result["TaxRate"].(int)) / 100))
		result["Amt"] = float64(result["TotalAmt"].(int)) - result["TaxAmt"].(float64)
	}

	if result["Category"] == "B2B" {

		result["PrintFlag"] = "Y"
		result["BuyerUBN"] = get(c.Payload, "buyer_ubn", "-")
		result["BuyerAddress"] = get(c.Payload, "buyer_address", "-")

		var taxfreePrice = []int{}
		for _, v := range result["ItemPrice"].([]interface{}) {
			price := int(math.Round(v.(float64) / float64(1+result["TaxRate"].(int)/100)))
			taxfreePrice = append(taxfreePrice, price)
		}
		result["ItemPrice"] = taxfreePrice
		delete(result, "CarrierType")

	} else if result["Category"] == "B2C" {

		if result["LoveCode"] != "" {
			// check if LoveCode is a 3~7 digits int string
			if match, _ := regexp.MatchString("^[0-9]{3,7}$", result["LoveCode"].(string)); !match {
				delete(result, "CarrierType")
				result["PrintFlag"] = "Y"
			} else {
				result["CarrierType"] = ""
			}
		} else {

			switch result["CarrierType"] {
			case "0":
				fallthrough
			case "1":
				var checkString string
				if result["CarrierType"] == "0" {
					checkString = "^/[A-Z0-9+-.]{7}$"
				} else if result["CarrierType"] == "1" {
					checkString = "^/[A-Z0-9+-.]{7}$"
				}
				if match, _ := regexp.MatchString(checkString, result["CarrierNum"].(string)); !match {
					delete(result, "CarrierType")
					result["PrintFlag"] = "Y"
					result["Comment"] = fmt.Sprintf("Incorrect carrier num: %s", result["CarrierNum"])
				} else {
					result["CarrierNum"] = strings.TrimSpace(result["CarrierNum"].(string))
				}
			case "2":
				if buyerEmail, ok := result["BuyerEmail"]; buyerEmail == "" || !ok {
					return errors.New("empty buyer_email when carrier_type = 2")
				}
				result["CarrierNum"] = result["BuyerEmail"]
			default:
				delete(result, "CarrierType")
				result["PrintFlag"] = "Y"
			}
		}
	}
	if len(result["ItemCount"].([]interface{})) == len(result["ItemPrice"].([]interface{})) {
		var itemAmt = []int{}
		for i := range result["ItemCount"].([]interface{}) {
			count := int(result["ItemCount"].([]interface{})[i].(float64))
			price := int(result["ItemPrice"].([]interface{})[i].(float64))
			itemAmt = append(itemAmt, count*price)
		}
		result["ItemAmt"] = itemAmt
	}

	result["ItemName"] = strings.Join(interfaceSliceToa(result["ItemName"].([]interface{})), "|")
	result["ItemUnit"] = strings.Join(interfaceSliceToa(result["ItemUnit"].([]interface{})), "|")
	result["ItemPrice"] = strings.Join(interfaceSliceItoa(result["ItemPrice"].([]interface{})), "|")
	result["ItemCount"] = strings.Join(interfaceSliceItoa(result["ItemCount"].([]interface{})), "|")
	result["ItemAmt"] = strings.Join(func(i []int) (r []string) {
		for _, v := range i {
			r = append(r, strconv.Itoa(v))
		}
		return r
	}(result["ItemAmt"].([]int)), "|")

	// Trim comment messeage to allowed length
	if comment, ok := result["Comment"].(string); ok && len(comment) > DefaultCommentLength {
		result["Comment"] = func(s string, l int) string {
			result := []rune(s)
			if len(result) > l {
				result = result[:l]
			}
			return string(result)
		}(comment, DefaultCommentLength)
	}
	c.Payload = result
	return nil
}
