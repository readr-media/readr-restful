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

type Project struct {
	ID            int        `json:"id" db:"project_id"`
	CreatedAt     NullTime   `json:"created_at" db:"created_at"`
	UpdatedAt     NullTime   `json:"updated_at" db:"updated_at"`
	UpdatedBy     NullString `json:"updated_by" db:"updated_by"`
	PublishedAt   NullString `json:"published_at" db:"published_at"`
	PostID        int        `json:"post_id" db:"post_id"`
	LikeAmount    NullInt    `json:"like_amount" db:"like_amount"`
	CommentAmount NullInt    `json:"comment_amount" db:"comment_amount"`
	Active        NullInt    `json:"active" db:"active"`
	HeroImage     NullString `json:"hero_image" db:"hero_image"`
	Title         NullString `json:"title" db:"title"`
	Description   NullString `json:"description" db:"description"`
	Author        NullString `json:"author" db:"author"`
	OgTitle       NullString `json:"og_title" db:"og_title"`
	OgDescription NullString `json:"og_description" db:"og_description"`
	OgImage       NullString `json:"og_image" db:"og_image"`
	Order         NullInt    `json:"project_order" db:"project_order" redis:"project_order"`
	Status        NullInt    `json:"status" db:"status" redis:"status"`
	Slug          NullString `json:"slug" db:"slug" redis:"slug"`
	Views         NullInt    `json:"views" db:"views" redis:"views"`
	PublishStatus NullInt    `json:"publish_status" db:"publish_status" redis:"publish_status"`
	Progress      NullFloat  `json:"progress" db:"progress" redis:"progress"`
	MemoPoints    NullInt    `json:"memo_points" db:"memo_points" redis:"memo_points"`
}

type projectAPI struct{}

type ProjectAPIInterface interface {
	CountProjects(args GetProjectArgs) (int, error)
	DeleteProjects(p Project) error
	GetProject(p Project) (Project, error)
	GetProjects(args GetProjectArgs) ([]ProjectAuthors, error)
	InsertProject(p Project) error
	UpdateProjects(p Project) error
	SchedulePublish() error
	InsertAuthors(id int, authors []int) (err error)
	UpdateAuthors(id int, authors []int) (err error)
}

type GetProjectArgs struct {
	// Match List
	IDs   []int    `form:"ids" json:"ids"`
	Slugs []string `form:"slugs" json:"slugs"`
	// IN/NOT IN
	Active        map[string][]int `form:"active" json:"active"`
	Status        map[string][]int `form:"status" json:"status"`
	PublishStatus map[string][]int `form:"publish_status" json:"publish_status"`
	// Where
	Keyword string `form:"keyword" json:"keyword"`
	// Result Shaper
	MaxResult int    `form:"max_result" json:"max_result"`
	Page      int    `form:"page" json:"page"`
	Sorting   string `form:"sort" json:"sort"`

	Fields sqlfields `form:"fields"`
}

func (g *GetProjectArgs) Default() {
	g.MaxResult = 20
	g.Page = 1
	g.Sorting = "-project_order,-updated_at"
}

func (g *GetProjectArgs) DefaultActive() {
	g.Active = map[string][]int{"$nin": []int{int(ProjectActive["deactive"].(float64))}}
}

func (p *GetProjectArgs) parse() (restricts string, values []interface{}) {
	where := make([]string, 0)

	if p.Active != nil {
		for k, v := range p.Active {
			where = append(where, fmt.Sprintf("%s %s (?)", "projects.active", operatorHelper(k)))
			values = append(values, v)
		}
	}
	if p.Status != nil {
		for k, v := range p.Status {
			where = append(where, fmt.Sprintf("%s %s (?)", "projects.status", operatorHelper(k)))
			values = append(values, v)
		}
	}
	if p.PublishStatus != nil {
		for k, v := range p.PublishStatus {
			where = append(where, fmt.Sprintf("%s %s (?)", "projects.publish_status", operatorHelper(k)))
			values = append(values, v)
		}
	}
	if len(p.IDs) != 0 {
		where = append(where, fmt.Sprintf("%s %s (?)", "projects.project_id", operatorHelper("in")))
		values = append(values, p.IDs)
	}
	if len(p.Slugs) != 0 {
		where = append(where, fmt.Sprintf("%s %s (?)", "projects.slug", operatorHelper("in")))
		values = append(values, p.Slugs)
	}
	if p.Keyword != "" {
		p.Keyword = fmt.Sprintf("%s%s%s", "%", p.Keyword, "%")
		where = append(where, "(projects.title LIKE ? OR projects.project_id LIKE ?)")
		values = append(values, p.Keyword, p.Keyword)
	}

	if len(where) > 1 {
		restricts = strings.Join(where, " AND ")
	} else if len(where) == 1 {
		restricts = where[0]
	}
	return restricts, values
}

func (p *GetProjectArgs) parseLimit() (limit map[string]string, values []interface{}) {
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

func (g *GetProjectArgs) FullAuthorTags() (result []string) {
	return getStructDBTags("full", Member{})
}

type ProjectAuthors struct {
	Project
	Authors []Stunt `json:"authors"`
}

type ProjectAuthor struct {
	Project
	Author Stunt `json:"author" db:"author"`
}

func (a *projectAPI) CountProjects(arg GetProjectArgs) (result int, err error) {
	restricts, values := arg.parse()
	query := fmt.Sprintf(`SELECT COUNT(project_id) FROM projects WHERE %s`, restricts)

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

func (a *projectAPI) GetProject(p Project) (Project, error) {
	project := Project{}
	err := DB.QueryRowx("SELECT * FROM projects WHERE project_id = ?", p.ID).StructScan(&project)
	switch {
	case err == sql.ErrNoRows:
		err = errors.New("Project Not Found")
		project = Project{}
	case err != nil:
		log.Println(err.Error())
		project = Project{}
	default:
		err = nil
	}
	return project, err
}

func (a *projectAPI) GetProjects(args GetProjectArgs) (result []ProjectAuthors, err error) {
	// Init appendable result slice
	result = make([]ProjectAuthors, 0)

	restricts, values := args.parse()
	if len(restricts) > 0 {
		restricts = fmt.Sprintf("WHERE %s", restricts)
	}
	limit, largs := args.parseLimit()
	// select *, a.nickname "a.nickname", a.member_id "a.member_id", a.points "a.points" from projects left join project_authors pa on projects.project_id = pa.project_id left join members a on pa.author_id = a.id where projects.project_id in (1000010, 1000013);
	values = append(values, largs...)

	query := fmt.Sprintf("SELECT projects.*, %s FROM (SELECT * FROM projects %s %s) AS projects LEFT JOIN project_authors pa ON projects.project_id = pa.project_id LEFT JOIN members author ON pa.author_id = author.id %s;",
		args.Fields.GetFields(`author.%s "author.%s"`), restricts, limit["full"], limit["order"])

	query, values, err = sqlx.In(query, values...)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	query = DB.Rebind(query)
	var pa []ProjectAuthor
	if err = DB.Select(&pa, query, values...); err != nil {
		log.Println(err.Error())
		return []ProjectAuthors{}, err
	}
	// For returning {"_items":null}
	if len(pa) == 0 {
		return nil, nil
	}
	for _, project := range pa {

		var notNullAuthor = func(in ProjectAuthor) ProjectAuthors {
			pas := ProjectAuthors{Project: in.Project}
			if project.Author != (Stunt{}) {
				pas.Authors = append(pas.Authors, in.Author)
			}
			return pas
		}
		// First Project
		if len(result) == 0 {
			result = append(result, notNullAuthor(project))
		} else {
			for i, v := range result {
				if v.ID == project.ID {
					result[i].Authors = append(result[i].Authors, project.Author)
					break
				} else {
					if i != (len(result) - 1) {
						continue
					} else {
						result = append(result, notNullAuthor(project))
					}
				}
			}
		}
	}
	return result, nil
}

func (a *projectAPI) InsertProject(p Project) error {

	query, _ := generateSQLStmt("insert", "projects", p)
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
		log.Printf("Fail to get last insert ID when insert a project: %v", err)
		return err
	}

	// Only insert a project when it's active
	if p.Active.Valid == true && p.Active.Int == 1 {
		if p.ID == 0 {
			p.ID = int(lastID)
		}
		arg := GetProjectArgs{}
		arg.Default()
		arg.IDs = []int{p.ID}
		arg.MaxResult = 1
		arg.Page = 1
		projects, err := ProjectAPI.GetProjects(arg)
		if err != nil {
			log.Printf("Error When Getting Project to Insert to Algolia: %v", err.Error())
			return nil
		}
		go Algolia.InsertProject(projects)
	}

	return nil
}

func (a *projectAPI) UpdateProjects(p Project) error {

	query, _ := generateSQLStmt("partial_update", "projects", p)
	result, err := DB.NamedExec(query, p)

	if err != nil {
		return err
	}
	rowCnt, err := result.RowsAffected()
	if rowCnt > 1 {
		return errors.New("More Than One Rows Affected") //Transaction rollback?
	} else if rowCnt == 0 {
		return errors.New("Project Not Found")
	}

	if p.Active.Valid == true && p.Active.Int != 1 {
		// Case: Set a project to unpublished state, Delete the project from cache/searcher
		go Algolia.DeleteProject([]int{p.ID})
	} else {
		// Case: Publish a project or update a project.
		// Read whole project from database, then store to cache/searcher.
		arg := GetProjectArgs{}
		arg.Default()
		arg.IDs = []int{p.ID}
		arg.MaxResult = 1
		arg.Page = 1
		projects, err := ProjectAPI.GetProjects(arg)
		if err != nil {
			log.Printf("Error When Getting Project to Insert to Algolia: %v", err.Error())
			return nil
		}
		go Algolia.InsertProject(projects)
	}
	return nil
}

func (a *projectAPI) DeleteProjects(p Project) error {

	result, err := DB.NamedExec("UPDATE projects SET active = 0 WHERE project_id = :project_id", p)
	if err != nil {
		log.Fatal(err)
	}
	afrows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if afrows == 0 {
		return errors.New("Project Not Found")
	}

	go Algolia.DeleteProject([]int{p.ID})

	return err
}

func (a *projectAPI) SchedulePublish() error {
	_, err := DB.Exec("UPDATE projects SET publish_status=2 WHERE publish_status=3 AND published_at <= cast(now() as datetime);")
	if err != nil {
		return err
	}
	return nil
}

// func (a *projectAPI) GetAuthors(args GetProjectArgs) (result []Stunt, err error) {
// 	//select a.nickname, a.member_id, a.active from project_authors pa left join members a on pa.author_id = a.id where pa.project_id in (1000010, 1000013);
// 	restricts, values := args.parse()
// 	fmt.Printf("restricts: %v\n,values:%v\n", restricts, values)
// 	fmt.Printf("args: %v\n", args)

// 	// projects.project_id IN (?), [1, 2]
// 	var where string
// 	if len(restricts) > 0 {
// 		where = fmt.Sprintf(" WHERE %s", restricts)
// 	}
// 	query := fmt.Sprintf(`SELECT %s FROM project_authors projects LEFT JOIN members author ON projects.author_id = author.id %s;`,
// 		args.Fields.GetFields(`author.%s "%s"`), where)
// 	fmt.Printf("query is :%s\n", query)
// 	fmt.Printf("values is %v\n", values)
// 	query, params, err := sqlx.In(query, values...)
// 	if err != nil {
// 		return []Stunt{}, err
// 	}

// 	query = DB.Rebind(query)
// 	if err := DB.Select(&result, query, params...); err != nil {
// 		return []Stunt{}, err
// 	}
// 	return result, nil
// }

func (a *projectAPI) InsertAuthors(projectID int, authorIDs []int) (err error) {

	var (
		valueStr     []string
		insertValues []interface{}
	)
	for _, author := range authorIDs {
		valueStr = append(valueStr, `(?, ?)`)
		insertValues = append(insertValues, projectID, author)
	}
	//INSERT IGNORE INTO project_authorIDs (project_id, author_id) VALUES ( ?, ? ), ( ?, ? );
	query := fmt.Sprintf(`INSERT IGNORE INTO project_authors (project_id, author_id) VALUES %s;`, strings.Join(valueStr, ", "))
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

func (a *projectAPI) UpdateAuthors(projectID int, authorIDs []int) (err error) {

	// Delete all author record if authorIDs is null
	if authorIDs == nil || len(authorIDs) == 0 {
		_, err = DB.Exec(`DELETE FROM project_authors WHERE project_id = ?`, projectID)
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

	del, args, err := sqlx.In(`DELETE FROM project_authors WHERE project_id = ? AND author_id NOT IN (?)`, projectID, authorIDs)
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
		insertValues = append(insertValues, projectID, author)
	}
	//INSERT IGNORE INTO project_authorIDs (project_id, author_id) VALUES ( ?, ? ), ( ?, ? );
	ins := fmt.Sprintf(`INSERT IGNORE INTO project_authors (project_id, author_id) VALUES %s;`, strings.Join(valueStr, ", "))
	_, err = tx.Exec(ins, insertValues...)
	if err != nil {
		return err
	}
	return err
}

var ProjectAPI ProjectAPIInterface = new(projectAPI)
var ProjectActive map[string]interface{}
var ProjectStatus map[string]interface{}
var ProjectPublishStatus map[string]interface{}
