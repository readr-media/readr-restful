package models

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"encoding/json"

	"github.com/readr-media/readr-restful/config"
	"github.com/readr-media/readr-restful/utils"
)

type Points struct {
	PointsID   int64      `json:"id" db:"id"`
	MemberID   int64      `json:"member_id" db:"member_id" binding:"required"`
	ObjectType int        `json:"object_type" db:"object_type" binding:"required"`
	ObjectID   int        `json:"object_id" db:"object_id"`
	Points     int        `json:"points" db:"points"`
	Balance    int        `json:"balance" db:"balance"`
	CreatedAt  NullTime   `json:"created_at" db:"created_at"`
	UpdatedBy  NullInt    `json:"updated_by" db:"updated_by"`
	UpdatedAt  NullTime   `json:"updated_at" db:"updated_at"`
	Reason     NullString `json:"reason" db:"reason"`
}

// PointsToken is made to solve problem if Token is added to Points struct
// Since in Insert method getStructDBTags is used,
// and *string seems going to be asserted as string and get an empty database field,
// resulting in insert NamedExec fail.
// I have to use embedded struct here. Might have to reform getStructDBTags
type PointsToken struct {
	Points
	Token       *string `json:"token,omitempty"`
	MemberName  *string `json:"member_name,omitempty"`
	MemberPhone *string `json:"member_phone,omitempty"`
	MemberMail  *string `json:"member_mail,omitempty"`
}

type pointsAPI struct{}

var PointsAPI PointsInterface = new(pointsAPI)

type PointsInterface interface {
	Get(args *PointsArgs) (result []PointsProject, err error)
	Insert(pts PointsToken) (result int, id int, err error)
}

type PointsArgs struct {
	ID         int64
	ObjectType *int64
	ObjectIDs  []int

	MaxResult uint8  `form:"max_result"`
	Page      uint16 `form:"page"`
	OrderBy   string `form:"sort"`
	PayType   string `form:"pay_type"`

	OSQL
}

type OSQL struct {
	query      string
	conditions []string
	joinstr    []string
	limits     []string
	args       []interface{}
	printargs  []interface{}
}

func (a *PointsArgs) get(query string) (result *PointsArgs) {
	a.OSQL.query = query
	return a
}

func (a *PointsArgs) join(jointype, table, on string) (query string, args []interface{}, err error) {
	a.query = fmt.Sprintf("%s %s JOIN %s ON %s", a.query, jointype, table, on)
	a.build()
	return a.query, a.args, nil
}

func (a *PointsArgs) build() {
	// Parse WHERE conditions
	if a.ID != 0 {
		a.conditions = append(a.conditions, "pts.member_id = ?")
		a.args = append(a.args, a.ID)
	}
	if a.ObjectType != nil {
		a.conditions = append(a.conditions, "pts.object_type = ?")
		a.args = append(a.args, int(*a.ObjectType))
	}
	if a.ObjectIDs != nil {
		ph := make([]string, len(a.ObjectIDs))
		for i := range a.ObjectIDs {
			ph[i] = "?"
			a.args = append(a.args, a.ObjectIDs[i])
		}
		a.conditions = append(a.conditions, fmt.Sprintf("pts.object_id IN (%s)", strings.Join(ph, ",")))
	}
	if a.PayType != "" {
		if a.PayType == "topup" {
			a.conditions = append(a.conditions, "pts.points < 0")
		} else if a.PayType == "consumption" {
			a.conditions = append(a.conditions, "pts.points > 0")
		}
	}
	if len(a.conditions) > 0 {
		a.query = fmt.Sprintf("%s WHERE %s", a.query, strings.Join(a.conditions, " AND "))
	}

	// Parse ORDER BY, LIMIT, OFFSET conditions
	if a.OrderBy != "" {
		a.limits = append(a.limits, fmt.Sprintf("ORDER BY %s", orderByHelper(a.OrderBy)))
	}
	if a.MaxResult != 0 {
		a.limits = append(a.limits, "LIMIT ?")
		a.args = append(a.args, a.MaxResult)
	}
	if a.Page != 0 {
		a.limits = append(a.limits, "OFFSET ?")
		a.args = append(a.args, (a.Page-1)*uint16(a.MaxResult))
	}
	if len(a.limits) > 0 {
		a.query = fmt.Sprintf("%s %s;", a.query, strings.Join(a.limits, " "))
	}
}

func (a *PointsArgs) Set(in map[string]interface{}) {
	for k, v := range in {
		switch k {
		case "max_result":
			a.MaxResult = uint8(v.(int))
		case "page":
			a.Page = uint16(v.(int))
		case "sort":
			a.OrderBy = v.(string)
		}
	}
}

type PointsProject struct {
	Points
	Title NullString `db:"title" json:"object_name"`
}

func (p *pointsAPI) Get(args *PointsArgs) (result []PointsProject, err error) {

	// GET should return point history of certain member_id rather than points id
	query, params, err := args.get("SELECT pts.*, pj.title FROM points AS pts").join("LEFT", "projects AS pj", "pts.object_id = pj.project_id")
	if err != nil {
		return nil, err
	}
	if err = DB.Select(&result, query, params...); err != nil {
		return nil, err
	}
	return result, err
}

func (p *pointsAPI) Insert(pts PointsToken) (result int, id int, err error) {
	tags := getStructDBTags("full", pts.Points)

	if pts.Points.Points < 0 && pts.Points.ObjectType == config.Config.Models.PointType["topup"] {

		if pts.Token != nil {
			// Member Pay with Prime Token
			reqBody, _ := json.Marshal(map[string]interface{}{
				// Token is aquired in frontend
				"prime":       pts.Token,
				"partner_key": config.Config.PaymentService.PartnerKey,
				"merchant_id": config.Config.PaymentService.MerchantID,
				// Real amount for TapPay should be positive
				// 100 would become 1 TWD in TapPay
				"amount":   0 - pts.Points.Points,
				"currency": config.Config.PaymentService.Currency,
				"details":  config.Config.PaymentService.PaymentDescription,
				"cardholder": map[string]string{
					"phone_number": *pts.MemberPhone,
					"name":         *pts.MemberName,
					"email":        *pts.MemberMail,
				},
			})

			_, body, err := utils.HTTPRequest("POST", config.Config.PaymentService.PrimeURL,
				map[string]string{
					"x-api-key": config.Config.PaymentService.PartnerKey,
				}, reqBody)

			if err != nil {
				log.Printf("Charge error:%v\n", err)
				return 0, 0, err
			}

			type PaymentResp struct {
				Status      int    `json:"status"`
				Message     string `json:"msg"`
				BankCode    string `json:"bank_result_code"`
				BankMessage string `json:"bank_result_msg"`
			}
			var paymentResp PaymentResp
			json.Unmarshal(body, &paymentResp)

			if paymentResp.Status != 0 {
				return 0, 0, errors.New(fmt.Sprintf("Payment Error, Code: %d, ErrorMsg: %s, BankSatusCode: %s, BankMsg: %s",
					paymentResp.Status, paymentResp.Message, paymentResp.BankCode, paymentResp.BankMessage))
			}
		}
	}

	tx, err := DB.Beginx()
	if err != nil {
		log.Printf("Fail to get sql connection: %v", err)
		return 0, 0, err
	}
	// Either rollback or commit transaction
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()
	// Choose the latest transaction balance
	if err = tx.Get(&result, `SELECT points FROM members WHERE id = ?`, pts.MemberID); err != nil {
		return 0, 0, err
	}
	// New Balance
	result = result - pts.Points.Points
	pts.Balance = result

	pointsU := fmt.Sprintf(`INSERT INTO points (%s) VALUES (:%s)`,
		strings.Join(tags, ","), strings.Join(tags, ",:"))

	inserted, err := tx.NamedExec(pointsU, pts)
	if err != nil {
		return 0, 0, err
	}
	transactionID, err := inserted.LastInsertId()
	if err != nil {
		return 0, 0, err
	}

	if _, err = tx.Exec(`UPDATE members SET points = ?, updated_at = ? WHERE id = ?`, result, pts.CreatedAt, pts.MemberID); err != nil {
		return 0, 0, err
	}
	return result, int(transactionID), err
}
