package cards

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/readr-media/readr-restful/config"
	"github.com/readr-media/readr-restful/internal/rrsql"
	//"github.com/readr-media/readr-restful/utils"
)

type NewsCard struct {
	ID              uint32           `json:"id" db:"id" redis:"id"`
	PostID          uint32           `json:"post_id" db:"post_id" redis:"post_id"`
	Title           rrsql.NullString `json:"title" db:"title" redis:"title"`
	Description     rrsql.NullString `json:"description" db:"description" redis:"description"`
	BackgroundImage rrsql.NullString `json:"background_image" db:"background_image" redis:"background_image"`
	BackgroundColor rrsql.NullString `json:"background_color" db:"background_color" redis:"background_color"`
	Image           rrsql.NullString `json:"image" db:"image" redis:"image"`
	Video           rrsql.NullString `json:"video" db:"video" redis:"video"`
	CreatedAt       rrsql.NullTime   `json:"created_at" db:"created_at" redis:"created_at"`
	UpdatedAt       rrsql.NullTime   `json:"updated_at" db:"updated_at" redis:"updated_at"`
	Order           rrsql.NullInt    `json:"order" db:"order" redis:"order"`
	Active          rrsql.NullInt    `json:"active" db:"active" redis:"active"`
	Status          rrsql.NullInt    `json:"status" db:"status" redis:"status"`
}

type NewsCardArgs struct {
	MaxResult uint8            `form:"max_result"`
	Page      uint16           `form:"page"`
	Sorting   string           `form:"sorting"`
	PostID    uint32           `form:"post_id"`
	IDs       []uint32         `form:"ids"`
	Active    map[string][]int `form:"active"`
	Status    map[string][]int `form:"status"`
}

func DefaultNewsCardArgs() (result *NewsCardArgs) {
	return &NewsCardArgs{
		MaxResult: 15,
		Page:      1,
		Active:    map[string][]int{"$nin": []int{config.Config.Models.Cards["deactive"]}},
	}
}

func (p *NewsCardArgs) parse() (restricts string, values []interface{}) {
	where := make([]string, 0)

	if p.Active != nil {
		for k, v := range p.Active {
			where = append(where, fmt.Sprintf("%s %s (?)", "newscards.active", operatorHelper(k)))
			values = append(values, v)
		}
	}
	if p.Status != nil {
		for k, v := range p.Status {
			where = append(where, fmt.Sprintf("%s %s (?)", "newscards.status", operatorHelper(k)))
			values = append(values, v)
		}
	}
	if p.IDs != nil {
		where = append(where, fmt.Sprintf("%s %s (?)", "newscards.id", "IN"))
		values = append(values, p.IDs)
	}
	if p.PostID != 0 {
		where = append(where, "newscards.post_id = ?")
		values = append(values, p.PostID)
	}
	if len(where) > 1 {
		restricts = fmt.Sprintf("WHERE %s", strings.Join(where, " AND "))
	} else if len(where) == 1 {
		restricts = fmt.Sprintf("WHERE %s", where[0])
	}
	return restricts, values
}

func (p *NewsCardArgs) parseResultLimit() (restricts string, values []interface{}) {
	sortingString := "created_at DESC"
	if p.Sorting != "" {
		sortingString = fmt.Sprintf("%s, %s", orderByHelper(p.Sorting), sortingString)
	}
	restricts = fmt.Sprintf("%s ORDER BY %s", restricts, sortingString)

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

func (p *NewsCardArgs) validateSorting(sort string) bool {
	for _, v := range strings.Split(sort, ",") {
		if matched, err := regexp.MatchString("-?(updated_at|created_at|id|order)", v); err != nil || !matched {
			return false
		}
	}
	return true
}

type newscardAPI struct{}

var NewsCardAPI NewsCardInterface = new(newscardAPI)

type NewsCardInterface interface {
	DeleteCard(id uint32) error
	GetCards(args *NewsCardArgs) (result []NewsCard, err error)
	InsertCard(c NewsCard) (int, error)
	UpdateCard(c NewsCard) error
}

func (a *newscardAPI) DeleteCard(id uint32) error {

	result, err := rrsql.DB.Exec(fmt.Sprintf("UPDATE newscards SET active = %d WHERE id = ?", config.Config.Models.Cards["deactive"]), id)
	if err != nil {
		return err
	}
	rowCnt, err := result.RowsAffected()
	if rowCnt > 1 {
		return errors.New("More Than One Rows Affected")
	} else if rowCnt == 0 {
		return errors.New("Card Not Found")
	}
	return err
}

func (a *newscardAPI) GetCards(rowargs *NewsCardArgs) (result []NewsCard, err error) {

	query, args := a.buildGetQuery(rowargs)
	// To give adaptability to where clauses, have to use ... operator here
	// Therefore split query into two parts, assembling them after sqlx.Rebind
	query, args, err = sqlx.In(query, args...)
	if err != nil {
		return nil, err
	}
	query = rrsql.DB.Rebind(query)

	rows, err := rrsql.DB.Queryx(query, args...)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var singleCard NewsCard
		if err = rows.StructScan(&singleCard); err != nil {
			result = []NewsCard{}
			return result, err
		}
		result = append(result, singleCard)
	}

	return result, err
}

func (a *newscardAPI) buildGetQuery(args *NewsCardArgs) (query string, values []interface{}) {
	selectedFields := []string{"newscards.*"}
	var restricts string

	restricts, restrictVals := args.parse()
	resultLimit, resultLimitVals := args.parseResultLimit()
	values = append(values, restrictVals...)
	values = append(values, resultLimitVals...)

	query = fmt.Sprintf(`
		SELECT %s FROM newscards %s `,
		strings.Join(selectedFields, ","),
		restricts+resultLimit,
	)

	return query, values
}

func (a *newscardAPI) InsertCard(n NewsCard) (int, error) {

	tags := getStructDBTags(n)
	fmt.Println(tags)
	query := fmt.Sprintf(`INSERT INTO newscards (%s) VALUES (:%s)`,
		strings.Join(tags, ","), strings.Join(tags, ",:"))

	result, err := rrsql.DB.NamedExec(query, n)
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
		return 0, errors.New("Card Not Found")
	}
	lastID, err := result.LastInsertId()
	if err != nil {
		log.Printf("Fail to get last insert ID when insert a card: %v", err)
		return 0, err
	}

	return int(lastID), err
}

func (a *newscardAPI) UpdateCard(n NewsCard) error {

	tags := getStructDBTags(n)
	fields := makeFieldString(
		`%s = :%s`, tags)
	query := fmt.Sprintf(`UPDATE newscards SET %s WHERE id = :id`,
		strings.Join(fields, ", "))

	result, err := rrsql.DB.NamedExec(query, n)

	if err != nil {
		return err
	}
	rowCnt, err := result.RowsAffected()
	if rowCnt > 1 {
		return errors.New("More Than One Rows Affected")
	} else if rowCnt == 0 {
		return errors.New("Card Not Found")
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
