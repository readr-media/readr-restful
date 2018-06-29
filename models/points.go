package models

import (
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
	Get(args *PointsArgs) (result []PointsProject, err error)
	Insert(pts Points) (result int, err error)
}

type PointsArgs struct {
	ID         int64
	ObjectType *int64
	ObjectIDs  []int

	MaxResult uint8  `form:"max_result"`
	Page      uint16 `form:"page"`
	OrderBy   string `form:"sort"`

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
