package models

import (
	"database/sql"
	"errors"
	//"fmt"
	"log"
	"strings"
)

type Project struct {
	ID string `json:"id" db:"project_id"`

	CreateTime NullTime   `json:"created_at" db:"created_at"` //NEED TO CHECK NAMING
	UpdatedAt  NullTime   `json:"updated_at" db:"updated_at"`
	UpdatedBy  NullString `json:"updated_by" db:"updated_by"`

	PostID        int `json:"post_id" db:"post_id"`
	LikeAmount    int `json:"like_amount" db:"like_amount"`
	CommentAmount int `json:"comment_amount" db:"comment_amount"`
	Active        int `json:"active" db:"active"`

	HeroImage     NullString `json:"hero_image" db:"hero_image"`
	Title         NullString `json:"title" db:"title"`
	Description   NullString `json:"description" db:"description"`
	Author        NullString `json:"author" db:"author"`
	OgTitle       NullString `json:"og_title" db:"og_title"`
	OgDescription NullString `json:"og_description" db:"og_description"`
	OgImage       NullString `json:"og_image" db:"og_image"`
}

type projectAPI struct{}

var ProjectAPI ProjectAPIInterface = new(projectAPI)

type ProjectAPIInterface interface {
	GetProject(p Project) (Project, error)
	GetProjects(ps ...Project) ([]Project, error)
	PostProject(p Project) error
	UpdateProjects(p Project) error
	DeleteProjects(p Project) error
}

func (a *projectAPI) GetProject(p Project) (Project, error) {

	project := Project{}
	err := DB.QueryRowx("SELECT * FROM project_infos WHERE project_id = ?", p.ID).StructScan(&project)
	switch {
	case err == sql.ErrNoRows:
		err = errors.New("Project Not Found")
		project = Project{}
	case err != nil:
		log.Fatal(err)
		project = Project{}
	default:
		//fmt.Printf("Successfully get project: %v\n", p.ID)
		err = nil
	}
	return project, err
}

func (a *projectAPI) GetProjects(ps ...Project) ([]Project, error) {
	return nil, nil
}

func (a *projectAPI) PostProject(p Project) error {

	query, _ := generateSQLStmt(p, "insert", "project_infos")
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
	return nil
}

func (a *projectAPI) UpdateProjects(p Project) error {

	query, _ := generateSQLStmt(p, "partial_update", "project_infos")
	result, err := DB.NamedExec(query, p)

	if err != nil {
		log.Fatal(err)
		return err
	}
	rowCnt, err := result.RowsAffected()
	if rowCnt > 1 {
		return errors.New("More Than One Rows Affected") //Transaction rollback?
	} else if rowCnt == 0 {
		return errors.New("Project Not Found")
	}
	return nil
}

func (a *projectAPI) DeleteProjects(p Project) error {

	_, err := DB.Exec("UPDATE members SET active = 0 WHERE user_id = ?", p.ID)
	if err != nil {
		log.Fatal(err)
	}
	return err
}
