package models

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"database/sql"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/readr-media/readr-restful/config"
)

type Report struct {
	ID              uint32     `json:"id" db:"post_id" redis:"post_id"`
	Author          NullInt    `json:"author" db:"author" redis:"author"`
	CreatedAt       NullTime   `json:"created_at" db:"created_at" redis:"created_at"`
	LikeAmount      NullInt    `json:"like_amount" db:"like_amount" redis:"like_amount"`
	CommentAmount   NullInt    `json:"comment_amount" db:"comment_amount" redis:"comment_amount"`
	Title           NullString `json:"title" db:"title" redis:"title"`
	Subtitle        NullString `json:"subtitle" db:"subtitle" redis:"subtitle"`
	Content         NullString `json:"content" db:"content" redis:"content"` //change from "description"
	Type            NullInt    `json:"type" db:"type" redis:"type"`
	Link            NullString `json:"link" db:"link" redis:"link"`
	OgTitle         NullString `json:"og_title" db:"og_title" redis:"og_title"`
	OgDescription   NullString `json:"og_description" db:"og_description" redis:"og_description"`
	OgImage         NullString `json:"og_image" db:"og_image" redis:"og_image"`
	Active          NullInt    `json:"active" db:"active" redis:"active"`
	UpdatedAt       NullTime   `json:"updated_at" db:"updated_at" redis:"updated_at"`
	UpdatedBy       NullInt    `json:"updated_by" db:"updated_by" redis:"updated_by"`
	PublishedAt     NullTime   `json:"published_at" db:"published_at" redis:"published_at"`
	LinkTitle       NullString `json:"link_title" db:"link_title" redis:"link_title"`
	LinkDescription NullString `json:"link_description" db:"link_description" redis:"link_description"`
	LinkImage       NullString `json:"link_image" db:"link_image" redis:"link_image"`
	LinkName        NullString `json:"link_name" db:"link_name" redis:"link_name"`
	VideoID         NullString `json:"video_id" db:"video_id" redis:"video_id"`
	VideoViews      NullInt    `json:"video_views" db:"video_views" redis:"video_views"`
	PublishStatus   NullInt    `json:"publish_status" db:"publish_status" redis:"publish_status"`
	ProjectID       NullInt    `json:"project_id" db:"project_id" redis:"project_id"`
	Order           NullInt    `json:"post_order" db:"post_order" redis:"post_order"`
	HeroImage       NullString `json:"hero_image" db:"hero_image" redis:"hero_image"`
	Slug            NullString `json:"slug" db:"slug" redis:"slug"`
	CSS             NullString `json:"css" db:"css" redis:"css"`
	JS              NullString `json:"javascript" db:"javascript" redis:"javascript"`
}

type reportAPI struct{}

type ReportAPIInterface interface {
	CountReports(args GetReportArgs) (int, error)
	DeleteReport(p Report) error
	GetReport(p Report) (Report, error)
	GetReports(args GetReportArgs) ([]ReportAuthors, error)
	InsertReport(p Report) (int, error)
	UpdateReport(p Report) error
	InsertAuthors(id int, authors []int) (err error)
	UpdateAuthors(id int, authors []int) (err error)
	SchedulePublish() (ids []int, err error)
}

type GetReportArgs struct {
	// Match List
	IDs          []int    `form:"ids" json:"ids"`
	Slugs        []string `form:"report_slugs" json:"report_slugs"`
	ProjectSlugs []string `form:"project_slugs" json:"project_slugs"`
	Project      []int64  `form:"project_id" json:"project_id"`
	// IN/NOT IN
	Active               map[string][]int `form:"active" json:"active"`
	ReportPublishStatus  map[string][]int `form:"report_publish_status"`
	ProjectPublishStatus map[string][]int `form:"project_publish_status"`
	// Where
	Keyword string `form:"keyword" json:"keyword"`
	// Result Shaper
	MaxResult int    `form:"max_result" json:"max_result"`
	Page      int    `form:"page" json:"page"`
	Sorting   string `form:"sort" json:"sort"`

	Fields sqlfields `form:"fields"`

	Filter Filter
}

func NewGetReportArgs(options ...func(*GetReportArgs)) *GetReportArgs {
	args := GetReportArgs{MaxResult: 20, Page: 1, Sorting: "-updated_at"}
	for _, option := range options {
		option(&args)
	}
	return &args
}

func (g *GetReportArgs) Default() {
	g.MaxResult = 20
	g.Page = 1
	g.Sorting = "-updated_at"
}

func (g *GetReportArgs) DefaultActive() {
	g.Active = map[string][]int{"$nin": []int{config.Config.Models.Reports["deactive"]}}
}

func (p *GetReportArgs) parse() (restricts string, values []interface{}) {
	where := make([]string, 0)
	where = append(where, fmt.Sprintf("%s %s %d", "posts.type", "=", config.Config.Models.PostType["report"]))

	if p.Active != nil {
		for k, v := range p.Active {
			where = append(where, fmt.Sprintf("%s %s (?)", "posts.active", operatorHelper(k)))
			values = append(values, v)
		}
	}
	if p.ReportPublishStatus != nil {
		for k, v := range p.ReportPublishStatus {
			where = append(where, fmt.Sprintf("%s %s (?)", "posts.publish_status", operatorHelper(k)))
			values = append(values, v)
		}
	}
	if p.ProjectPublishStatus != nil {
		for k, v := range p.ProjectPublishStatus {
			where = append(where, fmt.Sprintf("%s %s (?)", "projects.publish_status", operatorHelper(k)))
			values = append(values, v)
		}
	}
	if len(p.IDs) != 0 {
		where = append(where, fmt.Sprintf("%s %s (?)", "posts.post_id", operatorHelper("in")))
		values = append(values, p.IDs)
	}
	if len(p.Slugs) != 0 {
		where = append(where, fmt.Sprintf("%s %s (?)", "posts.slug", operatorHelper("in")))
		values = append(values, p.Slugs)
	}
	if len(p.ProjectSlugs) != 0 {
		where = append(where, fmt.Sprintf("%s %s (?)", "projects.slug", operatorHelper("in")))
		values = append(values, p.ProjectSlugs)
	}
	if len(p.Project) > 0 {
		where = append(where, fmt.Sprintf("%s IN (?)", "posts.project_id"))
		values = append(values, p.Project)
	}
	if p.Keyword != "" {
		p.Keyword = fmt.Sprintf("%s%s%s", "%", p.Keyword, "%")
		where = append(where, "(posts.title LIKE ? OR posts.post_id LIKE ?)")
		values = append(values, p.Keyword, p.Keyword)
	}
	if p.Filter != (Filter{}) {
		where = append(where, fmt.Sprintf("posts.%s %s ?", p.Filter.Field, p.Filter.Operator))
		values = append(values, p.Filter.Condition)
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
		restricts = append(restricts, fmt.Sprintf("ORDER BY %s%s", "posts.", orderByHelper(p.Sorting)))
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

// ------------ ↓↓↓ Requirement to satisfy LastPNRInterface  ↓↓↓ ------------

// ReturnPublishedAt is created to return published_at and used in pnr API
func (ra ReportAuthors) ReturnPublishedAt() time.Time {
	if ra.PublishedAt.Valid {
		return ra.PublishedAt.Time
	}
	return time.Time{}
}

// ReturnCreatedAt is created to return created_at and used in pnr API
func (ra ReportAuthors) ReturnCreatedAt() time.Time {
	if ra.CreatedAt.Valid {
		return ra.CreatedAt.Time
	}
	return time.Time{}
}

// ReturnUpdatedAt is created to return updated_at and used in pnr API
func (ra ReportAuthors) ReturnUpdatedAt() time.Time {
	if ra.UpdatedAt.Valid {
		return ra.UpdatedAt.Time
	}
	return time.Time{}
}

// ------------ ↑↑↑ End of requirement to satisfy LastPNRInterface  ↑↑↑ ------------

type ReportAuthor struct {
	Report
	Author  Stunt   `json:"author" db:"author"`
	Project Project `json:"projects" db:"projects"`
}

func (a *reportAPI) CountReports(arg GetReportArgs) (result int, err error) {
	restricts, values := arg.parse()
	query := fmt.Sprintf(`SELECT COUNT(posts.post_id) FROM posts LEFT JOIN projects AS projects ON projects.project_id = posts.project_id WHERE %s`, restricts)

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
	err := DB.QueryRowx("SELECT * FROM posts WHERE post_id = ?", p.ID).StructScan(&report)
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

	query := fmt.Sprintf("SELECT posts.*, %s, %s FROM (SELECT posts.* FROM posts LEFT JOIN projects AS projects ON projects.project_id = posts.project_id %s %s) AS posts LEFT JOIN report_authors ra ON posts.post_id = ra.report_id LEFT JOIN members author ON ra.author_id = author.id LEFT JOIN projects ON posts.project_id = projects.project_id %s;",
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
			ras := ReportAuthors{Report: in.Report, Project: in.Project}
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

func (a *reportAPI) InsertReport(p Report) (lastID int, err error) {

	p.Type = NullInt{int64(config.Config.Models.PostType["report"]), true}

	query, _ := generateSQLStmt("insert", "posts", p)
	result, err := DB.NamedExec(query, p)

	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return lastID, errors.New("Duplicate entry")
		}
		return lastID, err
	}
	rowCnt, err := result.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}
	if rowCnt > 1 {
		return lastID, errors.New("More Than One Rows Affected") //Transaction rollback?
	} else if rowCnt == 0 {
		return lastID, errors.New("No Row Inserted")
	}
	lastid, err := result.LastInsertId()
	if err != nil {
		log.Printf("Fail to get last insert ID when insert a report: %v", err)
		return lastID, err
	}
	lastID = int(lastid)
	if p.ID == 0 {
		p.ID = uint32(lastID)
	}

	return int(lastID), nil
}

func (a *reportAPI) UpdateReport(p Report) error {
	tags := getStructDBTags("partial", p)
	fields := makeFieldString("update", `%s = :%s`, tags)
	query := fmt.Sprintf(`UPDATE posts SET %s WHERE post_id = :post_id`, strings.Join(fields, ", "))
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

	return nil
}

func (a *reportAPI) DeleteReport(p Report) error {

	result, err := DB.NamedExec("UPDATE posts SET active = 0 WHERE post_id = :post_id", p)
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

	return err
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

func (a *reportAPI) SchedulePublish() (ids []int, err error) {

	rows, err := DB.Queryx(fmt.Sprintf("SELECT id FROM posts WHERE publish_status=3 AND published_at <= cast(now() as datetime) AND type = %d;", config.Config.Models.PostType["report"]))
	if err != nil {
		log.Println("Getting report error when schedule publishing reports", err)
		return ids, err
	}

	for rows.Next() {
		var i int
		if err = rows.Scan(&i); err != nil {
			continue
		}
		ids = append(ids, i)
	}

	_, err = DB.Exec(fmt.Sprintf("UPDATE posts SET publish_status=2 WHERE publish_status=3 AND published_at <= cast(now() as datetime) AND type = %d;", config.Config.Models.PostType["report"]))
	if err != nil {
		return ids, err
	}
	return ids, nil
}

var ReportAPI ReportAPIInterface = new(reportAPI)
