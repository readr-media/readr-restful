package asset

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/readr-media/readr-restful/config"
	"github.com/readr-media/readr-restful/models"
)

type Asset struct {
	ID            int64             `json:"id" db:"id"`
	Active        models.NullInt    `json:"active" db:"active"`
	CreatedAt     models.NullTime   `json:"created_at" db:"created_at"`
	CreatedBy     models.NullInt    `json:"created_by" db:"created_by"`
	UpdatedAt     models.NullTime   `json:"updated_at" db:"updated_at"`
	UpdatedBy     models.NullInt    `json:"updated_by" db:"updated_by"`
	Destination   models.NullString `json:"destination" db:"destination"`
	FileType      models.NullString `json:"file_type" db:"file_type"`
	FileName      models.NullString `json:"file_name" db:"file_name"`
	FileExtension models.NullString `json:"file_extension" db:"file_extension"`
	Title         models.NullString `json:"title" db:"title"`
	AssetType     models.NullInt    `json:"asset_type" db:"asset_type"`
	Copyright     models.NullInt    `json:"copyright" db:"copyright"`
}

type GetAssetArgs struct {
	MaxResult uint8            `form:"max_result"`
	Page      uint16           `form:"page"`
	Sorting   string           `form:"sort"`
	IDs       []uint32         `form:"ids"`
	AssetType []int            `form:"asset_type"`
	FileType  []string         `form:"file_type"`
	Active    map[string][]int `form:"active"`
}

func NewAssetArgs() *GetAssetArgs {
	return &GetAssetArgs{MaxResult: 20, Page: 1, Sorting: "-updated_at"}
}

func (a *GetAssetArgs) DefaultActive() {
	a.Active = map[string][]int{"$nin": []int{config.Config.Models.Assets["deactive"]}}
}

func (a *GetAssetArgs) parseCondition() (restricts string, values []interface{}) {
	where := make([]string, 0)

	if a.Active != nil {
		for k, v := range a.Active {
			where = append(where, fmt.Sprintf("%s %s (?)", "assets.active", operatorHelper(k)))
			values = append(values, v)
		}
	}
	if a.AssetType != nil {
		where = append(where, fmt.Sprintf("%s %s (?)", "assets.asset_type", "IN"))
		values = append(values, a.AssetType)
	}
	if a.FileType != nil {
		where = append(where, fmt.Sprintf("%s %s (?)", "assets.file_type", "IN"))
		values = append(values, a.FileType)
	}
	if a.IDs != nil {
		where = append(where, fmt.Sprintf("%s %s (?)", "assets.id", "IN"))
		values = append(values, a.IDs)
	}

	if len(where) > 1 {
		restricts = strings.Join(where, " AND ")
	} else if len(where) == 1 {
		restricts = where[0]
	}
	return restricts, values
}

func (p *GetAssetArgs) parseLimit() (restricts string, values []interface{}) {

	if p.Sorting != "" {
		restricts = fmt.Sprintf("%s ORDER BY %s", restricts, orderByHelper(p.Sorting))
	}

	if p.MaxResult > 0 {
		restricts = fmt.Sprintf("%s LIMIT ?", restricts)
		values = append(values, p.MaxResult)
		if p.Page > 0 {
			restricts = fmt.Sprintf("%s OFFSET ?", restricts)
			values = append(values, (p.Page-1)*uint16(p.MaxResult))
		}
	}
	return restricts, values
}

type assetAPI struct{}

type AssetInterface interface {
	Count(args *GetAssetArgs) (count int, err error)
	Delete(ids []int) (err error)
	GetAssets(args *GetAssetArgs) (result []Asset, err error)
	Insert(asset Asset) (lastID int64, err error)
	Update(asset Asset) (err error)
}

func (a *assetAPI) Count(args *GetAssetArgs) (count int, err error) {

	conditionSQL, conditionArgs := args.parseCondition()
	limitingSQL, limitingArgs := args.parseLimit()

	values := append(conditionArgs, limitingArgs...)
	query := fmt.Sprintf(`
		SELECT COUNT(*) FROM assets WHERE %s %s`,
		conditionSQL, limitingSQL,
	)

	query, values, err = sqlx.In(query, values...)
	if err != nil {
		return 0, err
	}
	query = models.DB.Rebind(query)

	rows, err := models.DB.Queryx(query, values...)
	if err != nil {
		return 0, err
	}
	for rows.Next() {
		if err = rows.Scan(&count); err != nil {
			return 0, err
		}
	}

	return count, err
}

func (m *assetAPI) Delete(ids []int) (err error) {

	query, args, err := sqlx.In(`UPDATE assets SET active = ? AND updated_at = ? WHERE id IN (?);`,
		models.NullInt{Int: int64(config.Config.Models.Posts["deactive"]), Valid: true},
		models.NullTime{Time: time.Now(), Valid: true},
		ids,
	)
	if err != nil {
		return err
	}

	query = models.DB.Rebind(query)

	result, err := models.DB.Exec(query, args...)
	if err != nil {
		return err
	}
	rowCnt, err := result.RowsAffected()
	if rowCnt > int64(len(ids)) {
		return errors.New("More Rows Affected")
	} else if rowCnt == 0 {
		return errors.New("Assets Not Found")
	}

	return nil
}

func (a *assetAPI) GetAssets(args *GetAssetArgs) (result []Asset, err error) {

	conditionSQL, conditionArgs := args.parseCondition()
	limitingSQL, limitingArgs := args.parseLimit()

	values := append(conditionArgs, limitingArgs...)
	query := fmt.Sprintf(`
		SELECT * FROM assets WHERE %s %s`,
		conditionSQL, limitingSQL,
	)
	query, values, err = sqlx.In(query, values...)
	if err != nil {
		return nil, err
	}
	query = models.DB.Rebind(query)
	rows, err := models.DB.Queryx(query, values...)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var asset Asset
		if err = rows.StructScan(&asset); err != nil {
			result = []Asset{}
			return result, err
		}
		result = append(result, asset)
	}

	return result, err
}

func (m *assetAPI) Insert(asset Asset) (lastID int64, err error) {

	tags := getStructDBTags(asset)
	query := fmt.Sprintf(`INSERT INTO assets (%s) VALUES (:%s)`,
		strings.Join(tags, ","), strings.Join(tags, ",:"))

	result, err := models.DB.NamedExec(query, asset)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return 0, errors.New("Duplicate entry")
		}
		return 0, err
	}
	rowCnt, err := result.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}
	if rowCnt > 1 {
		return 0, errors.New("More Than One Rows Affected")
	} else if rowCnt == 0 {
		return 0, errors.New("Insert Asset Fail")
	}
	lastID, err = result.LastInsertId()
	if err != nil {
		log.Printf("Fail to get last insert ID when insert an asset: %v", err)
		return 0, err
	}

	return lastID, err
}

func (m *assetAPI) Update(asset Asset) (err error) {

	tags := getStructDBTags(asset)
	fields := makeFieldString(`%s = :%s`, tags)
	query := fmt.Sprintf(`UPDATE assets SET %s WHERE id = :id`,
		strings.Join(fields, ", "))

	result, err := models.DB.NamedExec(query, asset)

	if err != nil {
		return err
	}
	rowCnt, err := result.RowsAffected()
	if rowCnt > 1 {
		return errors.New("More Than One Rows Affected")
	} else if rowCnt == 0 {
		return errors.New("Assets Not Found")
	}

	return err
}

func getStructDBTags(input interface{}) []string {
	columns := make([]string, 0)
	u := reflect.ValueOf(input)
	for i := 0; i < u.NumField(); i++ {
		tag := u.Type().Field(i).Tag
		field := u.Field(i).Interface()

		switch field := field.(type) {
		case string:
			if field != "" {
				columns = append(columns, tag.Get("db"))
			}
		case models.NullString:
			if field.Valid {
				columns = append(columns, tag.Get("db"))
			}
		case models.NullTime:
			if field.Valid {
				columns = append(columns, tag.Get("db"))
			}
		case models.NullInt:
			if field.Valid {
				columns = append(columns, tag.Get("db"))
			}
		case models.NullBool:
			if field.Valid {
				columns = append(columns, tag.Get("db"))
			}
		case models.NullIntSlice:
			if field.Valid {
				columns = append(columns, tag.Get("db"))
			}
		case bool, int, uint32, int64:
			columns = append(columns, tag.Get("db"))
		default:
			fmt.Println("unrecognised format: ", u.Field(i).Type())
		}
	}
	return columns
}

func makeFieldString(pattern string, tags []string) (result []string) {
	for _, value := range tags {
		result = append(result, fmt.Sprintf(pattern, value, value))
	}
	return result
}

func operatorHelper(ops string) (result string) {
	switch ops {
	case "$in":
		result = `IN`
	case "$nin":
		result = `NOT IN`
	default:
		result = `IN`
	}
	return result
}

func orderByHelper(sortMethod string) (result string) {
	// if strings.Contains(sortMethod, )
	tmp := strings.Split(sortMethod, ",")
	for i, v := range tmp {
		if v := strings.TrimSpace(v); strings.HasPrefix(v, "-") {
			tmp[i] = v[1:] + " DESC"
		} else {
			tmp[i] = v
		}
	}
	result = strings.Join(tmp, ",")
	return result
}

var AssetAPI AssetInterface = new(assetAPI)
