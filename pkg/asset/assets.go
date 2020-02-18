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
	"github.com/readr-media/readr-restful/internal/args"
	"github.com/readr-media/readr-restful/internal/rrsql"
	"github.com/readr-media/readr-restful/models"
)

type Asset struct {
	ID            int64            `json:"id" db:"id"`
	Active        rrsql.NullInt    `json:"active" db:"active"`
	CreatedAt     rrsql.NullTime   `json:"created_at" db:"created_at"`
	CreatedBy     rrsql.NullInt    `json:"created_by" db:"created_by"`
	UpdatedAt     rrsql.NullTime   `json:"updated_at" db:"updated_at"`
	UpdatedBy     rrsql.NullInt    `json:"updated_by" db:"updated_by"`
	Destination   rrsql.NullString `json:"destination" db:"destination"`
	FileType      rrsql.NullString `json:"file_type" db:"file_type"`
	FileName      rrsql.NullString `json:"file_name" db:"file_name"`
	FileExtension rrsql.NullString `json:"file_extension" db:"file_extension"`
	Title         rrsql.NullString `json:"title" db:"title"`
	AssetType     rrsql.NullInt    `json:"asset_type" db:"asset_type"`
	Copyright     rrsql.NullInt    `json:"copyright" db:"copyright"`
}

type FilteredAsset struct {
	ID        int64            `json:"id" db:"id"`
	FileName  rrsql.NullString `json:"file_name" db:"file_name"`
	AssetType rrsql.NullInt    `json:"asset_type" db:"asset_type"`
	UpdatedAt rrsql.NullTime   `json:"updated_at" db:"updated_at"`
}

type GetAssetArgs struct {
	MaxResult uint8            `form:"max_result"`
	Page      uint16           `form:"page"`
	Sorting   string           `form:"sort"`
	IDs       []uint32         `form:"ids"`
	AssetType []int            `form:"asset_type"`
	FileType  []string         `form:"file_type"`
	Active    map[string][]int `form:"active"`
	Total     bool             `form:"total"`
}

func NewAssetArgs() *GetAssetArgs {
	return &GetAssetArgs{MaxResult: 20, Page: 1, Sorting: "-updated_at"}
}

func (a *GetAssetArgs) DefaultActive() {
	a.Active = map[string][]int{"$nin": []int{config.Config.Models.Assets["deactive"]}}
}
func (a *GetAssetArgs) ParseCountQuery() (query string, values []interface{}) {
	conditionSQL, conditionArgs := a.parseRestricts()

	query = fmt.Sprintf(`
		SELECT COUNT(*) FROM assets WHERE %s`,
		conditionSQL,
	)

	return query, conditionArgs
}
func (a *GetAssetArgs) parseRestricts() (restricts string, values []interface{}) {
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

func (a *GetAssetArgs) parseLimit() (restricts string, values []interface{}) {

	if a.Sorting != "" {
		tmp := strings.Split(a.Sorting, ",")
		for i, v := range tmp {
			if v := strings.TrimSpace(v); strings.HasPrefix(v, "-") {
				tmp[i] = "-assets." + v[1:]
			} else {
				tmp[i] = "assets." + v
			}
		}
		sortFields := strings.Join(tmp, ",")
		restricts = fmt.Sprintf("%s ORDER BY %s", restricts, orderByHelper(sortFields))
	}

	if a.MaxResult > 0 {
		restricts = fmt.Sprintf("%s LIMIT ?", restricts)
		values = append(values, a.MaxResult)
		if a.Page > 0 {
			restricts = fmt.Sprintf("%s OFFSET ?", restricts)
			values = append(values, (a.Page-1)*uint16(a.MaxResult))
		}
	}
	return restricts, values
}

type FilterAssetArgs struct {
	models.FilterArgs
}

func (a *FilterAssetArgs) ParseQuery() (query string, values []interface{}) {
	return a.parse(false)
}
func (a *FilterAssetArgs) ParseCountQuery() (query string, values []interface{}) {
	return a.parse(true)
}

func (a *FilterAssetArgs) parse(doCount bool) (query string, values []interface{}) {
	fields := rrsql.GetStructDBTags("full", FilteredAsset{})
	for k, v := range fields {
		fields[k] = fmt.Sprintf("assets.%s", v)
	}
	selectedFields := strings.Join(fields, ",")

	restricts, restrictVals := a.parseFilterRestricts()
	limit, limitVals := a.parseLimit()
	values = append(values, restrictVals...)
	values = append(values, limitVals...)

	var joinedTables []string
	if len(a.Tag) > 0 {
		joinedTables = append(joinedTables, fmt.Sprintf(`
		LEFT JOIN tagging AS tagging ON tagging.target_id = assets.id AND tagging.type = %d LEFT JOIN tags AS tags ON tags.tag_id = tagging.tag_id
		`, config.Config.Models.TaggingType["asset"]))
	}

	if doCount {
		query = fmt.Sprintf(`
		SELECT %s FROM assets AS assets %s %s `,
			"COUNT(*)",
			strings.Join(joinedTables, " "),
			restricts,
		)
		values = restrictVals
	} else {
		query = fmt.Sprintf(`
		SELECT %s FROM assets AS assets %s %s `,
			selectedFields,
			strings.Join(joinedTables, " "),
			restricts+limit,
		)
	}

	return query, values
}
func (a *FilterAssetArgs) parseFilterRestricts() (restrictString string, values []interface{}) {
	restricts := make([]string, 0)

	if a.ID != 0 {
		restricts = append(restricts, `CAST(assets.id as CHAR) LIKE ?`)
		values = append(values, fmt.Sprintf("%s%d%s", "%", a.ID, "%"))
	}
	if len(a.Title) != 0 {
		subRestricts := make([]string, 0)
		for _, v := range a.Title {
			subRestricts = append(subRestricts, `assets.title LIKE ?`)
			values = append(values, fmt.Sprintf("%s%s%s", "%", v, "%"))
		}
		restricts = append(restricts, fmt.Sprintf("%s%s%s", "(", strings.Join(subRestricts, " OR "), ")"))
	}
	if len(a.Tag) != 0 {
		subRestricts := make([]string, 0)
		for _, v := range a.Tag {
			subRestricts = append(subRestricts, `tags.tag_content LIKE ?`)
			values = append(values, fmt.Sprintf("%s%s%s", "%", v, "%"))
		}
		restricts = append(restricts, fmt.Sprintf("(%s)", strings.Join(subRestricts, " OR ")))
	}
	if len(a.CreatedAt) != 0 {
		if v, ok := a.CreatedAt["$gt"]; ok {
			restricts = append(restricts, "assets.created_at >= ?")
			values = append(values, v)
		}
		if v, ok := a.CreatedAt["$lt"]; ok {
			restricts = append(restricts, "assets.created_at <= ?")
			values = append(values, v)
		}
	}
	if len(a.UpdatedAt) != 0 {
		if v, ok := a.UpdatedAt["$gt"]; ok {
			restricts = append(restricts, "assets.updated_at >= ?")
			values = append(values, v)
		}
		if v, ok := a.UpdatedAt["$lt"]; ok {
			restricts = append(restricts, "assets.updated_at <= ?")
			values = append(values, v)
		}
	}
	if len(restricts) > 1 {
		restrictString = fmt.Sprintf("WHERE %s", strings.Join(restricts, " AND "))
	} else if len(restricts) == 1 {
		restrictString = fmt.Sprintf("WHERE %s", restricts[0])
	}
	return restrictString, values
}

func (a *FilterAssetArgs) parseLimit() (restricts string, values []interface{}) {

	if a.Sorting != "" {
		tmp := strings.Split(a.Sorting, ",")
		for i, v := range tmp {
			if v := strings.TrimSpace(v); strings.HasPrefix(v, "-") {
				tmp[i] = "-assets." + v[1:]
			} else {
				tmp[i] = "assets." + v
			}
		}
		sortFields := strings.Join(tmp, ",")
		restricts = fmt.Sprintf("%s ORDER BY %s", restricts, orderByHelper(sortFields))
	}

	if a.MaxResult > 0 {
		restricts = fmt.Sprintf("%s LIMIT ?", restricts)
		values = append(values, a.MaxResult)
		if a.Page > 0 {
			restricts = fmt.Sprintf("%s OFFSET ?", restricts)
			values = append(values, (a.Page-1)*a.MaxResult)
		}
	}
	return restricts, values
}

type assetAPI struct{}

type AssetInterface interface {
	Count(args args.ArgsParser) (count int, err error)
	Delete(ids []int) (err error)
	FilterAssets(args *FilterAssetArgs) ([]FilteredAsset, error)
	GetAssets(args *GetAssetArgs) (result []Asset, err error)
	Insert(asset Asset) (lastID int64, err error)
	Update(asset Asset) (err error)
}

func (a *assetAPI) Count(args args.ArgsParser) (count int, err error) {

	query, values := args.ParseCountQuery()
	fmt.Println(query, values)
	query, values, err = sqlx.In(query, values...)
	if err != nil {
		return 0, err
	}
	query = rrsql.DB.Rebind(query)

	rows, err := rrsql.DB.Queryx(query, values...)
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
		rrsql.NullInt{Int: int64(config.Config.Models.Posts["deactive"]), Valid: true},
		rrsql.NullTime{Time: time.Now(), Valid: true},
		ids,
	)
	if err != nil {
		return err
	}

	query = rrsql.DB.Rebind(query)

	result, err := rrsql.DB.Exec(query, args...)
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
func (a *assetAPI) FilterAssets(args *FilterAssetArgs) (result []FilteredAsset, err error) {
	query, values := args.ParseQuery()

	rows, err := rrsql.DB.Queryx(query, values...)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var asset FilteredAsset
		if err = rows.StructScan(&asset); err != nil {
			return result, err
		}
		result = append(result, asset)
	}
	return result, nil
}

func (a *assetAPI) GetAssets(args *GetAssetArgs) (result []Asset, err error) {

	conditionSQL, conditionArgs := args.parseRestricts()
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
	query = rrsql.DB.Rebind(query)
	rows, err := rrsql.DB.Queryx(query, values...)
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

	result, err := rrsql.DB.NamedExec(query, asset)
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

	result, err := rrsql.DB.NamedExec(query, asset)

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
		case rrsql.NullString:
			if field.Valid {
				columns = append(columns, tag.Get("db"))
			}
		case rrsql.NullTime:
			if field.Valid {
				columns = append(columns, tag.Get("db"))
			}
		case rrsql.NullInt:
			if field.Valid {
				columns = append(columns, tag.Get("db"))
			}
		case rrsql.NullBool:
			if field.Valid {
				columns = append(columns, tag.Get("db"))
			}
		case rrsql.NullIntSlice:
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
