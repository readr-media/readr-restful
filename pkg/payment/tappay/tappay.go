package tappay

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/readr-media/readr-restful/config"
	"github.com/readr-media/readr-restful/utils"
)

// PaymentResp is the response from Pay By Prime API
type PaymentResp struct {
	Status      int    `json:"status"`
	Message     string `json:"msg"`
	BankCode    string `json:"bank_result_code"`
	BankMessage string `json:"bank_result_msg"`
	TradeID     string `json:"rec_trade_id"`
}

// Pay could send payment body for different url, e.g. ByPrime or ByToken
func Pay(url string, payment map[string]interface{}) (resp map[string]interface{}, err error) {

	reqBody, _ := json.Marshal(payment)
	log.Printf("payment request string: %v\n", string(reqBody))
	_, body, err := utils.HTTPRequest("POST", url,
		map[string]string{
			"x-api-key": config.Config.PaymentService.PartnerKey,
		}, reqBody)

	if err != nil {
		log.Printf("Charge error:%v\n", err)
		return resp, err
	}
	err = json.Unmarshal(body, &resp)

	return resp, err
}

// RecurringClient provides the Pay method for recurring payment
type RecurringClient struct{}

// Pay set necessary infos from config file and pass the payload to Pay()
func (c RecurringClient) Pay(payment map[string]interface{}) (resp map[string]interface{}, token map[string]interface{}, err error) {

	// Inject partner_key & merchant_id
	payment["partner_key"] = config.Config.PaymentService.PartnerKey
	payment["merchant_id"] = config.Config.PaymentService.MerchantID
	payment["three_domain_secure"] = true
	payment["frontend_redirect_url"] = config.Config.PaymentService.FrontendRedirectUrl
	payment["backend_notify_url"] = config.Config.PaymentService.BackendNotifyUrl

	resp, err = Pay(config.Config.PaymentService.TokenURL, payment)
	if err != nil {
		return nil, nil, err
	}
	if _, ok := resp["status"]; !ok {
		return nil, nil, errors.New("Pay by prime error: Unexpected response from tappay")
	}
	if status, ok := resp["status"].(float64); ok {
		if status != 0 && status != 2 {
			jsonStr, jsonErr := json.Marshal(resp)
			if jsonErr != nil {
				return nil, nil, errors.New(fmt.Sprintf("Pay by token error, Status: %f", status))
			} else {
				return nil, nil, errors.New(fmt.Sprintf("Pay by token error, %s", string(jsonStr)))
			}
		}
	} else {
		return nil, nil, errors.New("Pay by prime error: Unexpected response from tappay")
	}
	return resp, token, err
}

// OnetimeClient provides the Pay method for one-shot payment
type OnetimeClient struct{}

// Pay passes the payload with prime url to Pay()
func (c OnetimeClient) Pay(payment map[string]interface{}) (resp map[string]interface{}, token map[string]interface{}, err error) {

	// Retrieve key & id from config
	payment["partner_key"] = config.Config.PaymentService.PartnerKey
	payment["merchant_id"] = config.Config.PaymentService.MerchantID
	payment["three_domain_secure"] = true

	resp, err = Pay(config.Config.PaymentService.PrimeURL, payment)
	if err != nil {
		return nil, nil, err
	}
	if _, ok := resp["status"]; !ok {
		return nil, nil, errors.New("Pay by prime error: Unexpected response from tappay")
	}
	if status, ok := resp["status"].(float64); ok {
		if status != 0 && status != 2 {
			jsonStr, jsonErr := json.Marshal(resp)
			if jsonErr != nil {
				return nil, nil, errors.New(fmt.Sprintf("Pay by prime error, Status: %f", status))
			} else {
				return nil, nil, errors.New(fmt.Sprintf("Pay by prime error, %s", string(jsonStr)))
			}
		}
	} else {
		return nil, nil, errors.New("Pay by prime error: Unexpected response from tappay")
	}
	token = make(map[string]interface{})
	// Parse for pay by token
	if rem, ok := payment["remember"]; ok && rem == true {
		var cardSecret = resp["card_secret"].(map[string]interface{})
		token["card_key"] = cardSecret["card_key"]
		token["card_token"] = cardSecret["card_token"]
		token["currency"] = payment["currency"]
		token["details"] = payment["details"]
	}
	return resp, token, err
}
