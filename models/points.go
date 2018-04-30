package models

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
)

type Points struct {
	PointsID   uint32   `json:"id" db:"id"`
	MemberID   int64    `json:"member_id" db:"member_id" binding:"required"`
	ObjectType int      `json:"object_type" db:"object_type" binding:"required"`
	ObjectID   int      `json:"object_id" db:"object_id" binding:"required"`
	Points     int      `json:"points" db:"points"`
	CreatedAt  NullTime `json:"created_at" db:"created_at"`
	UpdatedBy  NullInt  `json:"updated_by" db:"updated_by"`
	UpdatedAt  NullTime `json:"updated_at" db:"updated_at"`
}

type pointsAPI struct{}

var PointsAPI PointsInterface = new(pointsAPI)

type PointsInterface interface {
	Get(id uint32, objType *int64) (result []Points, err error)
	Insert(pts Points) (result int, err error)
}

// Need to be change for the probability to accommodate id or id, objType type
func (p *pointsAPI) Get(id uint32, objType *int64) (result []Points, err error) {

	// GET should return point history of certain member_id rather than points id
	baseQ := `SELECT * FROM points WHERE member_id = ?`
	// Specifying object type case
	if objType != nil {
		var pts Points
		err = DB.QueryRowx(baseQ+" AND object_type = ?", id, int(*objType)).StructScan(&pts)
		switch {
		case err == sql.ErrNoRows:
			err = errors.New("Points Not Found")
			return []Points{}, err
		case err != nil:
			log.Fatal(err)
			return []Points{}, err
		default:
			err = nil
		}
		result = []Points{pts}
	} else {
		err = DB.Select(&result, baseQ, id)
		if err != nil {
			return []Points{}, err
		}
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
	pointsU := fmt.Sprintf(`INSERT INTO points (%s) VALUES (:%s)`,
		strings.Join(tags, ","), strings.Join(tags, ",:"))

	if _, err := tx.NamedExec(pointsU, pts); err != nil {
		return 0, err
	}
	memberU := fmt.Sprintf(`UPDATE members SET points = @updated_points := points - %d WHERE id = ?`, pts.Points)
	if _, err = tx.Exec(memberU, pts.MemberID); err != nil {
		return 0, err
	}
	row, err := tx.Queryx(`SELECT @updated_points`)
	if err != nil {
		return 0, err
	}
	for row.Next() {
		err = row.Scan(&result)
		if err != nil {
			return 0, err
		}
	}
	return result, err
}
