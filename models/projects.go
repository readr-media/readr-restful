package models

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"database/sql"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/readr-media/readr-restful/config"
)

type Project struct {
	ID            int        `json:"id" db:"project_id"`
	CreatedAt     NullTime   `json:"created_at" db:"created_at"`
	UpdatedAt     NullTime   `json:"updated_at" db:"updated_at"`
	UpdatedBy     NullInt    `json:"updated_by" db:"updated_by"`
	PublishedAt   NullTime   `json:"published_at" db:"published_at"`
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

type FilteredProject struct {
	ID            int        `json:"id" db:"project_id"`
	Title         NullString `json:"title" db:"title"`
	Slug          NullString `json:"slug" db:"slug"`
	Progress      NullFloat  `json:"progress" db:"progress"`
	Status        NullInt    `json:"status" db:"status"`
	PublishStatus NullInt    `json:"publish_status" db:"publish_status"`
	PublishedAt   NullTime   `json:"published_at" db:"published_at"`
}

type projectAPI struct{}

type ProjectAPIInterface interface {
	CountProjects(args GetProjectArgs) (int, error)
	DeleteProjects(p Project) error
	GetProject(p Project) (Project, error)
	GetProjects(args GetProjectArgs) ([]ProjectAuthors, error)
	GetContents(id int, args GetProjectArgs) ([]interface{}, error)
	FilterProjects(args GetProjectArgs) ([]interface{}, error)
	InsertProject(p Project) error
	UpdateProjects(p Project) error
	SchedulePublish() error
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
	// For determining to show memo abstract or not
	MemberID int64 `form:"member_id"`

	//Generate select fields
	Fields sqlfields `form:"fields"`

	// Filter operation
	FilterID          int64
	FilterSlug        string
	FilterTitle       []string
	FilterDescription []string
	FilterTagName     []string
	FilterPublishedAt map[string]time.Time
	FilterUpdatedAt   map[string]time.Time
}

func (g *GetProjectArgs) Default() {
	g.MaxResult = 20
	g.Page = 1
	g.Sorting = "-project_order,-updated_at"
}

func (g *GetProjectArgs) DefaultActive() {
	// g.Active = map[string][]int{"$nin": []int{int(ProjectActive["deactive"].(float64))}}
	g.Active = map[string][]int{"$nin": []int{config.Config.Models.ProjectsActive["deactive"]}}
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
		tmp := strings.Split(p.Sorting, ",")
		for i, v := range tmp {
			if v := strings.TrimSpace(v); strings.HasPrefix(v, "-") {
				tmp[i] = "-projects." + v[1:]
			} else {
				tmp[i] = "projects." + v
			}
		}

		p.Sorting = strings.Join(tmp, ",")

		restricts = append(restricts, fmt.Sprintf("ORDER BY %s", orderByHelper(p.Sorting)))
		limit["order"] = fmt.Sprintf("ORDER BY %s", orderByHelper(p.Sorting))
	}
	if p.MaxResult != 0 {
		restricts = append(restricts, "LIMIT ?")
		values = append(values, p.MaxResult)
		if p.Page != 0 {
			restricts = append(restricts, "OFFSET ?")
			values = append(values, (p.Page-1)*(p.MaxResult))
		}
	}
	if len(restricts) > 0 {
		limit["full"] = fmt.Sprintf(" %s", strings.Join(restricts, " "))
	}
	return limit, values
}

func (p *GetProjectArgs) parseFilterRestricts() (restrictString string, values []interface{}) {
	restricts := make([]string, 0)

	if p.FilterID != 0 {
		restricts = append(restricts, `CAST(projects.project_id as CHAR) LIKE ?`)
		values = append(values, fmt.Sprintf("%s%d%s", "%", p.FilterID, "%"))
	}
	if p.FilterSlug != "" {
		restricts = append(restricts, `projects.slug LIKE ?`)
		values = append(values, fmt.Sprintf("%s%s%s", "%", p.FilterSlug, "%"))
	}
	if len(p.FilterTitle) != 0 {
		subRestricts := make([]string, 0)
		for _, v := range p.FilterTitle {
			subRestricts = append(subRestricts, `projects.title LIKE ?`)
			values = append(values, fmt.Sprintf("%s%s%s", "%", v, "%"))
		}
		restricts = append(restricts, fmt.Sprintf("%s%s%s", "(", strings.Join(subRestricts, " OR "), ")"))
	}
	if len(p.FilterDescription) != 0 {
		subRestricts := make([]string, 0)
		for _, v := range p.FilterDescription {
			subRestricts = append(subRestricts, `projects.description LIKE ?`)
			values = append(values, fmt.Sprintf("%s%s%s", "%", v, "%"))
		}
		restricts = append(restricts, fmt.Sprintf("%s%s%s", "(", strings.Join(subRestricts, " OR "), ")"))
	}
	if len(p.FilterTagName) != 0 {
		subRestricts := make([]string, 0)
		for _, v := range p.FilterTagName {
			subRestricts = append(subRestricts, `tags.tag_content LIKE ?`)
			values = append(values, fmt.Sprintf("%s%s%s", "%", v, "%"))
		}
		restricts = append(restricts, fmt.Sprintf("(%s)", strings.Join(subRestricts, " OR ")))
	}

	if len(p.FilterPublishedAt) != 0 {
		if v, ok := p.FilterPublishedAt["$gt"]; ok {
			restricts = append(restricts, "projects.published_at >= ?")
			values = append(values, v)
		}
		if v, ok := p.FilterPublishedAt["$lt"]; ok {
			restricts = append(restricts, "projects.published_at <= ?")
			values = append(values, v)
		}
	}
	if len(p.FilterUpdatedAt) != 0 {
		if v, ok := p.FilterUpdatedAt["$gt"]; ok {
			restricts = append(restricts, "projects.updated_at >= ?")
			values = append(values, v)
		}
		if v, ok := p.FilterUpdatedAt["$lt"]; ok {
			restricts = append(restricts, "projects.updated_at <= ?")
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

func (p *GetProjectArgs) parseFilterQuery() (query string, values []interface{}) {

	var selectedFields string
	if len(p.Fields) == 0 {
		selectedFields = "*"
	} else {
		selectedFields = p.Fields.GetFields(`projects.%s "%s"`)
	}

	restricts, restrictVals := p.parseFilterRestricts()
	limit, limitVals := p.parseLimit()
	values = append(values, restrictVals...)
	values = append(values, limitVals...)

	var joinedTables []string
	if len(p.FilterTagName) > 0 {
		joinedTables = append(joinedTables, fmt.Sprintf(`
		LEFT JOIN tagging AS tagging ON tagging.target_id = projects.project_id AND tagging.type = %d LEFT JOIN tags AS tags ON tags.tag_id = tagging.tag_id 
		`, config.Config.Models.TaggingType["project"]))
	}

	query = fmt.Sprintf(`
		SELECT %s FROM projects AS projects %s %s `,
		selectedFields,
		strings.Join(joinedTables, " "),
		restricts+limit["full"],
	)

	return query, values
}

func (g *GetProjectArgs) FullAuthorTags() (result []string) {
	return getStructDBTags("full", Member{})
}

type ProjectAuthor struct {
	Project
	Tags   NullString `json:"-" db:"tags"`
	Author Stunt      `json:"author" db:"author"`
}

type SimpleTag struct {
	ID      int    `json:"id"`
	Content string `json:"text"`
}

type ProjectAuthors struct {
	Project
	Tags    NullString  `json:"-" db:"tags"`
	Authors []Stunt     `json:"authors"`
	TagList []SimpleTag `json:"tags"`
}

func (p *ProjectAuthors) formatTags() {
	if p.Tags.Valid != false {
		tas := strings.Split(p.Tags.String, ",")
	OuterLoop:
		for _, value := range tas {
			t := strings.Split(value, ":")
			id, _ := strconv.Atoi(t[0])
			for _, tag := range p.TagList {
				if id == tag.ID {
					continue OuterLoop
				}
			}
			p.TagList = append(p.TagList, SimpleTag{ID: id, Content: t[1]})
		}
	}
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

	query := fmt.Sprintf(`
		SELECT projects.*, t.tags, %s FROM (SELECT * FROM projects %s %s) AS projects 
		LEFT JOIN (
			SELECT pt.target_id as project_id, GROUP_CONCAT(CONCAT(t.tag_id, ":", t.tag_content) SEPARATOR ',') as tags
			FROM tagging as pt LEFT JOIN tags as t ON t.tag_id = pt.tag_id WHERE pt.type=%d 
			GROUP BY pt.target_id
			) AS t ON t.project_id = projects.project_id
		LEFT JOIN (
			SELECT DISTINCT authors.author_id, posts.project_id FROM authors 
			LEFT JOIN posts ON authors.resource_id = posts.post_id AND authors.resource_type = posts.type 
			WHERE posts.type = %d OR posts.type = %d 
			) AS pa ON projects.project_id = pa.project_id 
		LEFT JOIN members author ON pa.author_id = author.id %s;`,
		args.Fields.GetFields(`author.%s "author.%s"`),
		restricts,
		limit["full"],
		config.Config.Models.TaggingType["project"],
		config.Config.Models.PostType["memo"],
		config.Config.Models.PostType["report"],
		limit["order"])

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
	// For returning {"_items":[]}
	if len(pa) == 0 {
		return result, nil
	}
	for _, project := range pa {
		var notNullAuthor = func(in ProjectAuthor) ProjectAuthors {
			pas := ProjectAuthors{Project: in.Project, Tags: in.Tags}
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
		for k, pas := range result {
			pas.formatTags()
			result[k] = pas
		}
	}
	return result, nil
}

func (a *projectAPI) GetContents(id int, args GetProjectArgs) (result []interface{}, err error) {

	if args.MaxResult == 0 {
		args.MaxResult = 10
	}
	if args.Page == 0 {
		args.Page = 1
	}
	result = make([]interface{}, 0)

	query := fmt.Sprintf(`
		SELECT r.id, r.type FROM (
			SELECT post_id AS id, updated_at, CASE type WHEN 4 THEN 'report' WHEN 5 THEN 'memo' ELSE 'post' END AS type 
			FROM posts 
			WHERE active = %d AND publish_status = %d AND project_id = %d AND type IN (%d, %d, %d, %d) 
		) as r ORDER BY r.updated_at DESC LIMIT %d OFFSET %d;`,
		config.Config.Models.Posts["active"],
		config.Config.Models.PostPublishStatus["publish"],
		id,
		config.Config.Models.PostType["review"],
		config.Config.Models.PostType["news"],
		config.Config.Models.PostType["memo"],
		config.Config.Models.PostType["report"],
		args.MaxResult, (args.Page-1)*(args.MaxResult))

	var orderedList []struct {
		ID   int    `db:"id"`
		Type string `db:"type"`
	}

	if err = DB.Select(&orderedList, query); err != nil {
		log.Println(err.Error())
		return result, err
	}

	var postIDs []uint32
	var memoIDs []int64
	var reportIDs []int
	var posts []TaggedPostMember
	var memos []MemoDetail
	var reports []ReportAuthors

	for _, v := range orderedList {
		switch v.Type {
		case "post":
			postIDs = append(postIDs, uint32(v.ID))
		case "memo":
			memoIDs = append(memoIDs, int64(v.ID))
		case "report":
			reportIDs = append(reportIDs, v.ID)
		}
	}

	if len(postIDs) > 0 {
		posts, err = PostAPI.GetPosts(&PostArgs{
			ProjectID:    -1,
			MaxResult:    uint8(args.MaxResult),
			IDs:          postIDs,
			ShowAuthor:   true,
			ShowCommment: true,
			ShowTag:      true,
		})
		if err != nil {
			return result, err
		}
	}
	if len(memoIDs) > 0 {
		memos, err = MemoAPI.GetMemos(&MemoGetArgs{
			MaxResult:      args.MaxResult,
			IDs:            memoIDs,
			MemberID:       args.MemberID,
			AbstractLength: 20,
		})
		if err != nil {
			return result, err
		}
	}
	if len(reportIDs) > 0 {
		reports, err = ReportAPI.GetReports(GetReportArgs{
			MaxResult: args.MaxResult,
			Fields:    []string{"nickname"},
			IDs:       reportIDs,
		})
		if err != nil {
			return result, err
		}
	}

	for _, v := range orderedList {
		switch v.Type {
		case "post":
			for _, item := range posts {
				if v.ID == int(item.Post.ID) {
					result = append(result, item)
				}
			}
		case "memo":
			for _, item := range memos {
				if v.ID == int(item.ID) {
					result = append(result, item)
				}
			}
		case "report":
			for _, item := range reports {
				if v.ID == int(item.ID) {
					result = append(result, item)
				}
			}
		}
	}

	return result, nil
}

func (a *projectAPI) FilterProjects(args GetProjectArgs) (result []interface{}, err error) {
	query, values := args.parseFilterQuery()

	rows, err := DB.Queryx(query, values...)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var project FilteredProject
		if err = rows.StructScan(&project); err != nil {
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

	// Only insert a project when it's active
	if p.Active.Valid == true && p.Active.Int == 1 {
		if p.ID == 0 {
			p.ID = int(lastID)
		}
		arg := GetProjectArgs{}
		arg.Default()
		arg.IDs = []int{p.ID}
		arg.Fields = arg.FullAuthorTags()
		arg.MaxResult = 1
		arg.Page = 1
		projects, err := ProjectAPI.GetProjects(arg)
		if err != nil {
			log.Printf("Error When Getting Project to Insert to SearchFeed: %v", err.Error())
			return nil
		}
		go SearchFeed.InsertProject(projects)
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

	if (p.PublishStatus.Valid && p.PublishStatus.Int != int64(config.Config.Models.ProjectsPublishStatus["publish"])) ||
		(p.Active.Valid == true && p.Active.Int != int64(config.Config.Models.ProjectsActive["active"])) {
		// Case: Set a project to unpublished state, Delete the project from cache/searcher
		go SearchFeed.DeleteProject([]int{p.ID})
	} else if p.PublishStatus.Valid || p.Active.Valid {
		// Case: Publish a project or update a project.
		// Read whole project from database, then store to cache/searcher.
		arg := GetProjectArgs{}
		arg.Default()
		arg.IDs = []int{p.ID}
		arg.Fields = arg.FullAuthorTags()
		arg.MaxResult = 1
		arg.Page = 1
		projects, err := ProjectAPI.GetProjects(arg)
		if err != nil {
			log.Printf("Error When Getting Project to Insert to SearchFeed: %v", err.Error())
			return nil
		}

		if projects[0].PublishStatus.Int == int64(config.Config.Models.ProjectsPublishStatus["publish"]) &&
			projects[0].Active.Int == int64(config.Config.Models.ProjectsActive["active"]) {
			go SearchFeed.InsertProject(projects)
		}
	}

	if (p.Status.Valid && p.Status.Int == int64(config.Config.Models.ProjectsStatus["done"])) ||
		p.Progress.Valid {
		project, err := a.GetProject(p)
		if err != nil {
			return errors.New(fmt.Sprintf("Fail get project: %d", p.ID))
		}
		if project.PublishStatus.Int == int64(config.Config.Models.ProjectsPublishStatus["publish"]) &&
			project.Active.Int == int64(config.Config.Models.ProjectsActive["active"]) {
			go NotificationGen.GenerateProjectNotifications(p, "project")
		}
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

	go SearchFeed.DeleteProject([]int{p.ID})

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

// var ProjectActive map[string]interface{}
// var ProjectStatus map[string]interface{}
// var ProjectPublishStatus map[string]interface{}
