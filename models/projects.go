package models

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"strings"

	"database/sql"

	"github.com/jmoiron/sqlx"
)

type Project struct {
	ID int `json:"id" db:"project_id"`

	CreatedAt   NullTime   `json:"created_at" db:"created_at"`
	UpdatedAt   NullTime   `json:"updated_at" db:"updated_at"`
	UpdatedBy   NullString `json:"updated_by" db:"updated_by"`
	PublishedAt NullString `json:"published_at" db:"published_at"`

	PostID        int     `json:"post_id" db:"post_id"`
	LikeAmount    NullInt `json:"like_amount" db:"like_amount"`
	CommentAmount NullInt `json:"comment_amount" db:"comment_amount"`
	Active        NullInt `json:"active" db:"active"`

	HeroImage     NullString `json:"hero_image" db:"hero_image"`
	Title         NullString `json:"title" db:"title"`
	Description   NullString `json:"description" db:"description"`
	Author        NullString `json:"author" db:"author"`
	OgTitle       NullString `json:"og_title" db:"og_title"`
	OgDescription NullString `json:"og_description" db:"og_description"`
	OgImage       NullString `json:"og_image" db:"og_image"`
	Order         NullInt    `json:"order" db:"project_order" redis:"order"`
}

type projectAPI struct{}

type ProjectAPIInterface interface {
	DeleteProjects(p Project) error
	//GetProject(p Project) (Project, error)
	GetProjects(args GetProjectArgs) ([]Project, error)
	InsertProject(p Project) error
	UpdateProjects(p Project) error
}

type GetProjectArgs struct {
	IDs       []int `form:"ids" json:"ids"`
	MaxResult int   `form:"max_result" json:"max_result"`
	Page      int   `form:"page" json:"page"`
}

func (g *GetProjectArgs) Init() {
	g.MaxResult = 50
	g.Page = 1
}

func (a *projectAPI) GetProject(p Project) (Project, error) {

	project := Project{}
	err := DB.QueryRowx("SELECT * FROM projects WHERE project_id = ?", p.ID).StructScan(&project)
	switch {
	case err == sql.ErrNoRows:
		err = errors.New("Project Not Found")
		project = Project{}
	case err != nil:
		log.Fatal(err)
		project = Project{}
	default:
		err = nil
	}
	return project, err
}

func (a *projectAPI) GetProjects(args GetProjectArgs) ([]Project, error) {
	var query bytes.Buffer
	var bindvars []interface{}
	query.WriteString(fmt.Sprintf(`SELECT p.* FROM projects as p WHERE active = %d`, int(ProjectStatus["active"].(float64))))
	if len(args.IDs) > 0 {
		inquery, inargs, err := sqlx.In(` AND project_id IN (?)`, args.IDs)
		if err != nil {
			log.Println(err.Error())
			return nil, err
		}

		bindvars = inargs
		inquery = DB.Rebind(inquery)
		query.WriteString(inquery)

	}
	query.WriteString(` ORDER BY project_order DESC, updated_at DESC LIMIT ? OFFSET ?;`)
	bindvars = append(bindvars, args.MaxResult, (args.Page-1)*args.MaxResult)

	rows, err := DB.Queryx(query.String(), bindvars...)
	if err != nil {
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
		arg := GetProjectArgs{IDs: []int{p.ID}, MaxResult: 1, Page: 1}
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
		arg := GetProjectArgs{IDs: []int{p.ID}, MaxResult: 1, Page: 1}
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
var ProjectStatus map[string]interface{}
