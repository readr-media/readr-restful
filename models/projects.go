package models

import (
)

type Project struct {
	ID string `json:"id" db:"project_id"`

	CreateTime NullTime   `json:"created_at" db:"create_time"` //NEED TO CHECK NAMING
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

type projectAPI struct {}

func (a *projectAPI) GetProjects(p Project) (Project, error) {
	//to be implemented...
	return p, nil
}

var ProjectAPI projectAPI

