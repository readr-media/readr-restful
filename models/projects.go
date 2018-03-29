package models

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"database/sql"
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
}

type projectAPI struct{}

type ProjectAPIInterface interface {
	DeleteProjects(p Project) error
	GetProject(p Project) (Project, error)
	GetProjects(args GetProjectArgs) ([]Project, error)
	InsertProject(p Project) error
	UpdateProjects(p Project) error
}

type GetProjectArgs struct {
	// Match List
	IDs   []int    `form:"ids" json:"ids"`
	Slugs []string `form:"slugs" json:"slugs"`
	// IN/NOT IN
	Active map[string][]int `form:"active" json:"active"`
	Status map[string][]int `form:"status" json:"status"`
	// Where
	Keyword string `form:"keyword" json:"keyword"`
	// Result Shaper
	MaxResult int    `form:"max_result" json:"max_result"`
	Page      int    `form:"page" json:"page"`
	Sorting   string `form:"sort" json:"sort"`
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
	if len(p.IDs) != 0 {
		where = append(where, fmt.Sprintf("%s %s (?)", "projects.project_id", operatorHelper("in")))
		values = append(values, p.IDs)
	}
	if len(p.Slugs) != 0 {
		where = append(where, fmt.Sprintf("%s %s (?)", "projects.slug", operatorHelper("in")))
		values = append(values, p.Slugs)
	}
	if p.Keyword != "" {
		where = append(where, "(projects.title LIKE ? OR projects.description LIKE ?)")
		values = append(values, p.Keyword, p.Keyword)
	}

	if len(where) > 1 {
		restricts = strings.Join(where, " AND ")
	} else if len(where) == 1 {
		restricts = where[0]
	}
	return restricts, values
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

func (a *projectAPI) GetProjects(args GetProjectArgs) ([]Project, error) {
	restricts, values := args.parse()
	query := fmt.Sprintf("SELECT * FROM projects WHERE %s ORDER BY %s LIMIT ? OFFSET ?;", restricts, orderByHelper(args.Sorting))
	values = append(values, args.MaxResult, (args.Page-1)*args.MaxResult)

	query, values, err := sqlx.In(query, values...)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	query = DB.Rebind(query)

	rows, err := DB.Queryx(query, values...)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	var result = []Project{}
	for rows.Next() {
		var project Project
		if err = rows.StructScan(&project); err != nil {
			result = []Project{}
			return result, err
		}
		result = append(result, project)
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

	// Only insert a post when it's published
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
			log.Println("Error When Getting Project to Insert to Algolia: %v", err.Error())
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
			log.Println("Error When Getting Project to Insert to Algolia: %v", err.Error())
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

var ProjectAPI ProjectAPIInterface = new(projectAPI)
var ProjectActive map[string]interface{}
var ProjectStatus map[string]interface{}
