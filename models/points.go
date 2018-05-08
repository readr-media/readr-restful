package models

import (
	"bytes"
	"fmt"
	"log"
	"strings"
)

type Points struct {
	PointsID   int64    `json:"id" db:"id"`
	MemberID   int64    `json:"member_id" db:"member_id" binding:"required"`
	ObjectType int      `json:"object_type" db:"object_type" binding:"required"`
	ObjectID   int      `json:"object_id" db:"object_id" binding:"required"`
	Points     int      `json:"points" db:"points"`
	Balance    int      `json:"balance" db:"balance"`
	CreatedAt  NullTime `json:"created_at" db:"created_at"`
	UpdatedBy  NullInt  `json:"updated_by" db:"updated_by"`
	UpdatedAt  NullTime `json:"updated_at" db:"updated_at"`
}

type pointsAPI struct{}

var PointsAPI PointsInterface = new(pointsAPI)

type PointsInterface interface {
	Get(args *PointsArgs) (result []Points, err error)
	Insert(pts Points) (result int, err error)
}

type PointsArgs struct {
	ID         int64
	ObjectType *int64
	ObjectIDs  []int

	MaxResult uint8  `form:"max_result"`
	Page      uint16 `form:"page"`
	OrderBy   string `form:"sort"`

	query bytes.Buffer
	where []interface{}
}

func (a *PointsArgs) parseWhere() {
	restricts := make([]string, 0)
	if a.ID != 0 {
		restricts = append(restricts, "member_id = ?")
		a.where = append(a.where, a.ID)
	}
	if a.ObjectType != nil {
		restricts = append(restricts, "object_type = ?")
		a.where = append(a.where, int(*a.ObjectType))
	}
	if a.ObjectIDs != nil {
		ph := make([]string, len(a.ObjectIDs))
		for i := range a.ObjectIDs {
			ph[i] = "?"
			a.where = append(a.where, a.ObjectIDs[i])
		}
		restricts = append(restricts, fmt.Sprintf("object_id IN (%s)", strings.Join(ph, ",")))
	}
	if len(restricts) > 0 {
		a.query.WriteString(fmt.Sprintf(" WHERE %s", strings.Join(restricts, " AND ")))
	}
}

func (a *PointsArgs) parseLimit() {
	restricts := make([]string, 0)
	if a.OrderBy != "" {
		restricts = append(restricts, fmt.Sprintf("ORDER BY %s", orderByHelper(a.OrderBy)))
	}
	if a.MaxResult != 0 {
		restricts = append(restricts, "LIMIT ?")
		a.where = append(a.where, a.MaxResult)
	}
	if a.Page != 0 {
		restricts = append(restricts, "OFFSET ?")
		a.where = append(a.where, (a.Page-1)*uint16(a.MaxResult))
	}
	if len(restricts) > 0 {
		a.query.WriteString(fmt.Sprintf(" %s", strings.Join(restricts, " ")))
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
func (a *PointsArgs) selectQuery(initial string) {

	a.query.WriteString(initial)
	a.parseWhere()
	a.parseLimit()
	a.query.WriteString(";")
}

func (p *pointsAPI) Get(args *PointsArgs) (result []Points, err error) {

	// GET should return point history of certain member_id rather than points id
	args.selectQuery(`SELECT * FROM points`)
	if err = DB.Select(&result, args.query.String(), args.where...); err != nil {
		return []Points{}, err
	}
	return result, err
}

func (p *pointsAPI) Insert(pts Points) (result int, err error) {
	tags := getStructDBTags("full", pts)

	tx, err := DB.Beginx()
	if err != nil {
		log.Printf("Fail to get sql connection: %v", err)
		return 0, err
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
		return 0, err
	}
	// New Balance
	result = result - pts.Points
	pts.Balance = result
	pointsU := fmt.Sprintf(`INSERT INTO points (%s) VALUES (:%s)`,
		strings.Join(tags, ","), strings.Join(tags, ",:"))

	if _, err := tx.NamedExec(pointsU, pts); err != nil {
		return 0, err
	}
	if _, err = tx.Exec(`UPDATE members SET points = ?, updated_at = ? WHERE id = ?`, result, pts.CreatedAt, pts.MemberID); err != nil {
		return 0, err
	}
	return result, err
}
