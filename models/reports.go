package models

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"database/sql"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type Report struct {
	ID            int        `json:"id" db:"id"`
	CreatedAt     NullTime   `json:"created_at" db:"created_at"`
	LikeAmount    NullInt    `json:"like_amount" db:"like_amount"`
	CommentAmount NullInt    `json:"comment_amount" db:"comment_amount"`
	Title         NullString `json:"title" db:"title"`
	Description   NullString `json:"description" db:"description"`
	HeroImage     NullString `json:"hero_image" db:"hero_image"`
	OgTitle       NullString `json:"og_title" db:"og_title"`
	OgDescription NullString `json:"og_description" db:"og_description"`
	OgImage       NullString `json:"og_image" db:"og_image"`
	Active        NullInt    `json:"active" db:"active"`
	ProjectID     int        `json:"project_id" db:"project_id"`
	UpdatedAt     NullTime   `json:"updated_at" db:"updated_at"`
	UpdatedBy     NullString `json:"updated_by" db:"updated_by"`
	PublishedAt   NullTime   `json:"published_at" db:"published_at"`
	Slug          NullString `json:"slug" db:"slug"`
	Views         NullInt    `json:"views" db:"views"`
	PublishStatus NullInt    `json:"publish_status" db:"publish_status"`
}

type reportAPI struct{}

type ReportAPIInterface interface {
	CountReports(args GetReportArgs) (int, error)
	DeleteReport(p Report) error
	GetReport(p Report) (Report, error)
	GetReports(args GetReportArgs) ([]ReportAuthors, error)
	InsertReport(p Report) error
	UpdateReport(p Report) error
	SchedulePublish() error
	InsertAuthors(id int, authors []int) (err error)
	UpdateAuthors(id int, authors []int) (err error)
}

type GetReportArgs struct {
	// Match List
	IDs   []int    `form:"ids" json:"ids"`
	Slugs []string `form:"slugs" json:"slugs"`
	// IN/NOT IN
	Active        map[string][]int `form:"active" json:"active"`
	PublishStatus map[string][]int `form:"publish_status" json:"publish_status"`
	// Where
	Keyword string `form:"keyword" json:"keyword"`
	// Result Shaper
	MaxResult int    `form:"max_result" json:"max_result"`
	Page      int    `form:"page" json:"page"`
	Sorting   string `form:"sort" json:"sort"`

	Fields sqlfields `form:"fields"`
}

func (g *GetReportArgs) Default() {
	g.MaxResult = 20
	g.Page = 1
	g.Sorting = "-updated_at"
}

func (g *GetReportArgs) DefaultActive() {
	g.Active = map[string][]int{"$nin": []int{int(ReportActive["deactive"].(float64))}}
}

func (p *GetReportArgs) parse() (restricts string, values []interface{}) {
	where := make([]string, 0)

	if p.Active != nil {
		for k, v := range p.Active {
			where = append(where, fmt.Sprintf("%s %s (?)", "reports.active", operatorHelper(k)))
			values = append(values, v)
		}
	}
	if p.PublishStatus != nil {
		for k, v := range p.PublishStatus {
			where = append(where, fmt.Sprintf("%s %s (?)", "reports.publish_status", operatorHelper(k)))
			values = append(values, v)
		}
	}
	if len(p.IDs) != 0 {
		where = append(where, fmt.Sprintf("%s %s (?)", "reports.id", operatorHelper("in")))
		values = append(values, p.IDs)
	}
	if len(p.Slugs) != 0 {
		where = append(where, fmt.Sprintf("%s %s (?)", "reports.slug", operatorHelper("in")))
		values = append(values, p.Slugs)
	}
	if p.Keyword != "" {
		p.Keyword = fmt.Sprintf("%s%s%s", "%", p.Keyword, "%")
		where = append(where, "(reports.title LIKE ? OR reports.id LIKE ?)")
		values = append(values, p.Keyword, p.Keyword)
	}

	if len(where) > 1 {
		restricts = strings.Join(where, " AND ")
	} else if len(where) == 1 {
		restricts = where[0]
	}
	return restricts, values
}

func (p *GetReportArgs) parseLimit() (limit map[string]string, values []interface{}) {
	restricts := make([]string, 0)
	limit = make(map[string]string, 2)
	if p.Sorting != "" {
		restricts = append(restricts, fmt.Sprintf("ORDER BY %s", orderByHelper(p.Sorting)))
		limit["order"] = fmt.Sprintf("ORDER BY %s", orderByHelper(p.Sorting))
	}
	if p.MaxResult != 0 {
		restricts = append(restricts, "LIMIT ?")
		values = append(values, p.MaxResult)
	}
	if p.Page != 0 {
		restricts = append(restricts, "OFFSET ?")
		values = append(values, (p.Page-1)*(p.MaxResult))
	}
	if len(restricts) > 0 {
		limit["full"] = fmt.Sprintf(" %s", strings.Join(restricts, " "))
	}
	return limit, values
}

func (g *GetReportArgs) FullAuthorTags() (result []string) {
	return getStructDBTags("full", Member{})
}

type ReportAuthors struct {
	Report
	Authors []Stunt `json:"authors"`
	Project Project `json:"project"`
}

type ReportAuthor struct {
	Report
	Author  Stunt   `json:"author" db:"author"`
	Project Project `json:"projects" db:"projects"`
}

func (a *reportAPI) CountReports(arg GetReportArgs) (result int, err error) {
	restricts, values := arg.parse()
	query := fmt.Sprintf(`SELECT COUNT(id) FROM reports WHERE %s`, restricts)

	query, args, err := sqlx.In(query, values...)
	if err != nil {
		return 0, err
	}
	query = DB.Rebind(query)
	count, err := DB.Queryx(query, args...)
	if err != nil {
		return 0, err
	}
	for count.Next() {
		if err = count.Scan(&result); err != nil {
			return 0, err
		}
	}
	return result, err
}

func (a *reportAPI) GetReport(p Report) (Report, error) {
	report := Report{}
	err := DB.QueryRowx("SELECT * FROM reports WHERE id = ?", p.ID).StructScan(&report)
	switch {
	case err == sql.ErrNoRows:
		err = errors.New("Report Not Found")
		report = Report{}
	case err != nil:
		log.Println(err.Error())
		report = Report{}
	default:
		err = nil
	}
	return report, err
}

func (a *reportAPI) GetReports(args GetReportArgs) (result []ReportAuthors, err error) {
	// Init appendable result slice
	result = make([]ReportAuthors, 0)

	restricts, values := args.parse()
	if len(restricts) > 0 {
		restricts = fmt.Sprintf("WHERE %s", restricts)
	}
	limit, largs := args.parseLimit()
	values = append(values, largs...)

	projectTags := getStructDBTags("full", Project{})
	projectField := makeFieldString("get", `projects.%s "projects.%s"`, projectTags)
	projectIDQuery := strings.Split(projectField[0], " ")
	projectPostQuery := strings.Split(projectField[5], " ")
	projectField[0] = fmt.Sprintf(`IFNULL(%s, 0) %s`, projectIDQuery[0], projectIDQuery[1])
	projectField[5] = fmt.Sprintf(`IFNULL(%s, 0) %s`, projectPostQuery[0], projectPostQuery[1])

	query := fmt.Sprintf("SELECT reports.*, %s, %s FROM (SELECT * FROM reports %s %s) AS reports LEFT JOIN report_authors ra ON reports.id = ra.report_id LEFT JOIN members author ON ra.author_id = author.id LEFT JOIN projects ON reports.project_id = projects.project_id %s;",
		args.Fields.GetFields(`author.%s "author.%s"`), strings.Join(projectField, ","), restricts, limit["full"], limit["order"])

	query, values, err = sqlx.In(query, values...)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	query = DB.Rebind(query)
	var ra []ReportAuthor

	if err = DB.Select(&ra, query, values...); err != nil {
		log.Println(err.Error())
		return []ReportAuthors{}, err
	}
	// For returning {"_items":null}
	if len(ra) == 0 {
		return result, nil
	}
	for _, report := range ra {

		var notNullAuthor = func(in ReportAuthor) ReportAuthors {
			ras := ReportAuthors{Report: in.Report}
			if report.Author != (Stunt{}) {
				ras.Authors = append(ras.Authors, in.Author)
			}
			return ras
		}
		// First Report
		if len(result) == 0 {
			result = append(result, notNullAuthor(report))
		} else {
			for i, v := range result {
				if v.ID == report.ID {
					result[i].Authors = append(result[i].Authors, report.Author)
					break
				} else {
					if i != (len(result) - 1) {
						continue
					} else {
						result = append(result, notNullAuthor(report))
					}
				}
			}
		}
	}
	return result, nil
}

func (a *reportAPI) InsertReport(p Report) error {

	query, _ := generateSQLStmt("insert", "reports", p)
	result, err := DB.NamedExec(query, p)

	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return errors.New("Duplicate entry")
		}
		return err
	}
	rowCnt, err := result.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}
	if rowCnt > 1 {
		return errors.New("More Than One Rows Affected") //Transaction rollback?
	} else if rowCnt == 0 {
		return errors.New("No Row Inserted")
	}
	lastID, err := result.LastInsertId()
	if err != nil {
		log.Printf("Fail to get last insert ID when insert a report: %v", err)
		return err
	}

	// Only insert a report when it's active
	if p.Active.Valid == true && p.Active.Int == int64(ReportActive["active"].(float64)) {
		if p.ID == 0 {
			p.ID = int(lastID)
		}
		arg := GetReportArgs{}
		arg.Default()
		arg.IDs = []int{p.ID}
		arg.Fields = arg.FullAuthorTags()
		arg.MaxResult = 1
		arg.Page = 1
		reports, err := ReportAPI.GetReports(arg)
		if err != nil {
			log.Printf("Error When Getting Report to Insert to Algolia: %v", err.Error())
			return nil
		}
		go Algolia.InsertReport(reports)
	}

	return nil
}

func (a *reportAPI) UpdateReport(p Report) error {

	query, _ := generateSQLStmt("partial_update", "reports", p)
	result, err := DB.NamedExec(query, p)

	if err != nil {
		return err
	}
	rowCnt, err := result.RowsAffected()
	if rowCnt > 1 {
		return errors.New("More Than One Rows Affected") //Transaction rollback?
	} else if rowCnt == 0 {
		return errors.New("Report Not Found")
	}

	if p.Active.Valid == true && p.Active.Int != int64(ReportActive["active"].(float64)) {
		// Case: Set a report to unpublished state, Delete the report from cache/searcher
		go Algolia.DeleteReport([]int{p.ID})
	} else {
		// Case: Publish a report or update a report.
		// Read whole report from database, then store to cache/searcher.
		arg := GetReportArgs{}
		arg.Default()
		arg.IDs = []int{p.ID}
		arg.Fields = arg.FullAuthorTags()
		arg.MaxResult = 1
		arg.Page = 1
		reports, err := ReportAPI.GetReports(arg)
		if err != nil {
			log.Printf("Error When Getting Report to Insert to Algolia: %v", err.Error())
			return nil
		}
		go Algolia.InsertReport(reports)
	}
	return nil
}

func (a *reportAPI) DeleteReport(p Report) error {

	result, err := DB.NamedExec("UPDATE reports SET active = 0 WHERE id = :id", p)
	if err != nil {
		log.Fatal(err)
	}
	afrows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if afrows == 0 {
		return errors.New("Report Not Found")
	}

	go Algolia.DeleteReport([]int{p.ID})

	return err
}

func (a *reportAPI) SchedulePublish() error {
	_, err := DB.Exec("UPDATE reports SET publish_status=2 WHERE publish_status=3 AND published_at <= cast(now() as datetime);")
	if err != nil {
		return err
	}
	return nil
}

func (a *reportAPI) InsertAuthors(reportID int, authorIDs []int) (err error) {

	var (
		valueStr     []string
		insertValues []interface{}
	)
	for _, author := range authorIDs {
		valueStr = append(valueStr, `(?, ?)`)
		insertValues = append(insertValues, reportID, author)
	}
	query := fmt.Sprintf(`INSERT IGNORE INTO report_authors (report_id, author_id) VALUES %s;`, strings.Join(valueStr, ", "))
	_, err = DB.Exec(query, insertValues...)
	if err != nil {
		sqlerr, ok := err.(*mysql.MySQLError)
		if ok && sqlerr.Number == 1062 {
			return DuplicateError
		}
		return err
	}
	return err
}

func (a *reportAPI) UpdateAuthors(reportID int, authorIDs []int) (err error) {

	// Delete all author record if authorIDs is null
	if authorIDs == nil || len(authorIDs) == 0 {
		_, err = DB.Exec(`DELETE FROM report_authors WHERE report_id = ?`, reportID)
		if err != nil {
			return err
		}
		return nil
	}
	tx, err := DB.Beginx()
	if err != nil {
		log.Printf("Fail to get sql connection: %v", err)
		return err
	}
	// Either rollback or commit transaction
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	del, args, err := sqlx.In(`DELETE FROM report_authors WHERE report_id = ? AND author_id NOT IN (?)`, reportID, authorIDs)
	if err != nil {
		log.Printf("Fail to generate query: %v", err)
		return err
	}
	del = DB.Rebind(del)
	_, err = tx.Exec(del, args...)
	if err != nil {

	}
	var (
		valueStr     []string
		insertValues []interface{}
	)
	for _, author := range authorIDs {
		valueStr = append(valueStr, `(?, ?)`)
		insertValues = append(insertValues, reportID, author)
	}
	ins := fmt.Sprintf(`INSERT IGNORE INTO report_authors (report_id, author_id) VALUES %s;`, strings.Join(valueStr, ", "))
	_, err = tx.Exec(ins, insertValues...)
	if err != nil {
		return err
	}
	return err
}

var ReportAPI ReportAPIInterface = new(reportAPI)
var ReportActive map[string]interface{}
var ReportPublishStatus map[string]interface{}
