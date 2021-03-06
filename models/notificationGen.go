package models

import (
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/readr-media/readr-restful/config"
	"github.com/readr-media/readr-restful/internal/rrsql"
)

type notificationGenerator struct{}

type NotificationGenInterface interface {
	GenerateCommentNotifications(comment InsertCommentArgs) (err error)
	GenerateProjectNotifications(resource interface{}, resourceTyep string) (err error)
	GeneratePostNotifications(p TaggedPostMember) (err error)
}

func (c *notificationGenerator) getFollowers(resourceID int, resourceType int, emotion []int) (followers []int, err error) {
	fInterface, err := FollowingAPI.Get(&GetFollowerMemberIDsArgs{int64(resourceID), resourceType, emotion})
	if err != nil {
		//log.Println("Error get followers type:", resourceType, " id:", resourceID, err.Error())
		return followers, err
	}
	followers, ok := fInterface.([]int)
	if !ok {
		//log.Println("Error assert fInterface type:", resourceType, err.Error())
		return followers, errors.New(fmt.Sprintf("Error assert Interface resource type:%d when get followers", resourceType))
	}
	return followers, err
}

func (c *notificationGenerator) mergeFollowerSlices(a []int, b []int) (r []int) {
	r = a
	for _, bf := range b {
		for _, af := range a {
			if af == bf {
				break
			}
			r = append(r, bf)
		}
	}
	return r
}

func (c notificationGenerator) GenerateCommentNotifications(comment InsertCommentArgs) (err error) {
	ns := Notifications{}
	resourceID := int(comment.ResourceID.Int)
	resourceName := comment.ResourceName.String

	commentDetail, err := CommentAPI.GetComment(int(comment.ID))
	if err != nil {
		log.Println("Error get comment", comment.ID, err.Error())
	}

	var parentCommentDetail CommentAuthor
	if commentDetail.ParentID.Valid {
		parentCommentDetail, err = CommentAPI.GetComment(int(commentDetail.ParentID.Int))
		if err != nil {
			log.Println("Error get parent comment", commentDetail.ParentID.Int, err.Error())
		}
	}

	switch resourceName {
	case "post":

		post, err := PostAPI.GetPost(uint32(resourceID), &PostArgs{
			ProjectID:  -1,
			ShowAuthor: true,
		})
		if err != nil {
			log.Println("Error get post", resourceID, err.Error())
			return err
		}
		if len(post.Authors) == 0 {
			log.Println("Error post no author", resourceID)
			return errors.New(fmt.Sprintf("Error post %d has no author", resourceID))
		}

		postFollowers, err := c.getFollowers(resourceID, config.Config.Models.FollowingType["post"],
			[]int{config.Config.Models.Emotions["like"], config.Config.Models.Emotions["dislike"]})
		if err != nil {
			log.Println("Error get post followers", resourceID, err.Error())
			return err
		}
		//log.Println(postFollowers)

		var commentors []int
		// rows, err := rrsql.DB.Queryx(fmt.Sprintf(`SELECT DISTINCT author FROM comments WHERE resource="%s" AND status = %d AND active = %d;`, commentDetail.Resource.String, int(CommentStatus["show"].(float64)), int(CommentActive["active"].(float64))))
		rows, err := rrsql.DB.Queryx(fmt.Sprintf(`SELECT DISTINCT author FROM comments WHERE resource="%s" AND status = %d AND active = %d;`, commentDetail.Resource.String, config.Config.Models.CommentStatus["show"], config.Config.Models.Comment["active"]))

		if err != nil {
			log.Println("Error get commentors", commentDetail.Resource.String, err.Error())
			return err
		}
		for rows.Next() {
			var i int
			err := rows.Scan(&i)
			if err != nil {
				log.Println("Error scan commentors", err)
				return err
			}
			commentors = append(commentors, i)
		}

		for _, v := range postFollowers {
			if v != int(commentDetail.Author.Int) {
				ns = append(ns, NewNotification("follow_post_reply", v))
			}
		}

		if c.checkForAuthor(int(commentDetail.Author.Int), post.Authors) {
			ns = append(ns, NewNotification("post_reply", int(post.Author.Int)))
		}

		if len(commentors) > 0 {
			for _, v := range commentors {
				if v != int(commentDetail.Author.Int) {
					ns = append(ns, NewNotification("comment_comment", v))
				}
			}
		}

		if parentCommentDetail.Author.Valid && parentCommentDetail.Author.Int != commentDetail.Author.Int {
			if c.checkForAuthor(int(commentDetail.Author.Int), post.Authors) {
				ns = append(ns, NewNotification("comment_reply_author", int(parentCommentDetail.Author.Int)))
			} else {
				ns = append(ns, NewNotification("comment_reply", int(parentCommentDetail.Author.Int)))
			}
		}

		for k, v := range ns {
			v.SubjectID = strconv.Itoa(int(commentDetail.Author.Int))
			v.Nickname = commentDetail.AuthorNickname.String
			v.ProfileImage = commentDetail.AuthorImage.String
			v.ObjectName = post.Authors[0].Nickname.String
			v.ObjectType = resourceName
			v.ObjectID = strconv.Itoa(resourceID)
			v.PostType = strconv.Itoa(int(post.Type.Int))
			ns[k] = v
		}

		break

	case "project":
		project, err := ProjectAPI.GetProject(Project{ID: resourceID})
		if err != nil {
			log.Println("Error get project", resourceID, err.Error())
		}

		projectFollowers, err := c.getFollowers(resourceID, config.Config.Models.FollowingType["project"], []int{0})
		if err != nil {
			log.Println("Error get project followers", resourceID, err.Error())
		}
		log.Println(projectFollowers)

		var commentors []int
		// rows, err := rrsql.DB.Queryx(fmt.Sprintf(`SELECT DISTINCT author FROM comments WHERE resource="%s" AND status = %d AND active = %d;`, commentDetail.Resource.String, int(CommentStatus["show"].(float64)), int(CommentActive["active"].(float64))))
		rows, err := rrsql.DB.Queryx(fmt.Sprintf(`SELECT DISTINCT author FROM comments WHERE resource="%s" AND status = %d AND active = %d;`, commentDetail.Resource.String, config.Config.Models.CommentStatus["show"], config.Config.Models.Comment["active"]))
		if err != nil {
			log.Println("Error get commentors", commentDetail.Resource.String, err.Error())
			return err
		}
		for rows.Next() {
			var i int
			err := rows.Scan(&i)
			if err != nil {
				log.Println("Error scan commentors", err)
				return err
			}
			commentors = append(commentors, i)
		}

		for _, v := range projectFollowers {
			if v != int(commentDetail.Author.Int) {
				ns = append(ns, NewNotification("follow_project_reply", v))
			}
		}

		if len(commentors) > 0 {
			for _, v := range commentors {
				if v != int(commentDetail.Author.Int) {
					ns = append(ns, NewNotification("comment_comment", v))
				}
			}
		}

		if parentCommentDetail.Author.Valid && parentCommentDetail.Author.Int != commentDetail.Author.Int {
			ns = append(ns, NewNotification("comment_reply", int(parentCommentDetail.Author.Int)))
		}

		for k, v := range ns {
			v.SubjectID = strconv.Itoa(int(commentDetail.Author.Int))
			v.Nickname = commentDetail.AuthorNickname.String
			v.ProfileImage = commentDetail.AuthorImage.String
			v.ObjectName = project.Title.String
			v.ObjectType = resourceName
			v.ObjectID = strconv.Itoa(resourceID)
			v.ObjectSlug = project.Slug.String
			ns[k] = v
		}

		break

	case "memo":
		memo, err := MemoAPI.GetMemo(resourceID)
		if err != nil {
			log.Println("Error get memo", resourceID, err.Error())
		}
		project, err := ProjectAPI.GetProject(Project{ID: int(memo.ProjectID.Int)})
		if err != nil {
			log.Println("Error get project", memo.ProjectID.Int, err.Error())
		}

		projectFollowers, err := c.getFollowers(int(memo.ProjectID.Int), config.Config.Models.FollowingType["project"], []int{0})
		if err != nil {
			log.Println("Error get project followers", memo.ProjectID.Int, err.Error())
		}
		memoFollowers, err := c.getFollowers(resourceID, config.Config.Models.FollowingType["memo"],
			[]int{config.Config.Models.Emotions["like"], config.Config.Models.Emotions["dislike"]})
		if err != nil {
			log.Println("Error get project followers", resourceID, err.Error())
		}

		followers := c.mergeFollowerSlices(memoFollowers, projectFollowers)

		var commentors []int
		// rows, err := rrsql.DB.Queryx(fmt.Sprintf(`SELECT DISTINCT author FROM comments WHERE resource="%s" AND status = %d AND active = %d;`, commentDetail.Resource.String, int(CommentStatus["show"].(float64)), int(CommentActive["active"].(float64))))
		rows, err := rrsql.DB.Queryx(fmt.Sprintf(`SELECT DISTINCT author FROM comments WHERE resource="%s" AND status = %d AND active = %d;`, commentDetail.Resource.String, config.Config.Models.CommentStatus["show"], config.Config.Models.Comment["active"]))
		if err != nil {
			log.Println("Error get commentors", commentDetail.Resource.String, err.Error())
			return err
		}
		for rows.Next() {
			var i int
			err := rows.Scan(&i)
			if err != nil {
				log.Println("Error scan commentors", err)
				return err
			}
			commentors = append(commentors, i)
		}

		for _, v := range followers {
			if v != int(commentDetail.Author.Int) {
				ns = append(ns, NewNotification("follow_memo_reply", v))
			}
		}

		if len(commentors) > 0 {
			for _, v := range commentors {
				if v != int(commentDetail.Author.Int) {
					ns = append(ns, NewNotification("comment_comment", v))
				}
			}
		}

		if parentCommentDetail.Author.Valid && parentCommentDetail.Author.Int != commentDetail.Author.Int {
			ns = append(ns, NewNotification("comment_reply", int(parentCommentDetail.Author.Int)))
		}

		for k, v := range ns {
			v.SubjectID = strconv.Itoa(int(commentDetail.Author.Int))
			v.Nickname = commentDetail.AuthorNickname.String
			v.ProfileImage = commentDetail.AuthorImage.String
			v.ObjectName = memo.Title.String
			v.ObjectType = resourceName
			v.ObjectID = strconv.Itoa(resourceID)
			v.ObjectSlug = project.Slug.String
			ns[k] = v
		}
		break

	case "report":
		report, err := ReportAPI.GetReport(Report{ID: uint32(resourceID)})
		if err != nil {
			log.Println("Error get memo", resourceID, err.Error())
		}
		projectFollowers, err := c.getFollowers(int(report.ProjectID.Int), config.Config.Models.FollowingType["project"], []int{0})
		if err != nil {
			log.Println("Error get project followers", report.ProjectID, err.Error())
		}
		reportFollowers, err := c.getFollowers(resourceID, config.Config.Models.FollowingType["report"],
			[]int{config.Config.Models.Emotions["like"], config.Config.Models.Emotions["dislike"]})
		if err != nil {
			log.Println("Error get report followers", resourceID, err.Error())
		}

		followers := c.mergeFollowerSlices(reportFollowers, projectFollowers)

		var commentors []int
		// rows, err := rrsql.DB.Queryx(fmt.Sprintf(`SELECT DISTINCT author FROM comments WHERE resource="%s" AND status = %d AND active = %d;`, commentDetail.Resource.String, int(CommentStatus["show"].(float64)), int(CommentActive["active"].(float64))))
		rows, err := rrsql.DB.Queryx(fmt.Sprintf(`SELECT DISTINCT author FROM comments WHERE resource="%s" AND status = %d AND active = %d;`, commentDetail.Resource.String, config.Config.Models.CommentStatus["show"], config.Config.Models.Comment["active"]))
		if err != nil {
			log.Println("Error get commentors", commentDetail.Resource.String, err.Error())
			return err
		}
		for rows.Next() {
			var i int
			err := rows.Scan(&i)
			if err != nil {
				log.Println("Error scan commentors", err)
				return err
			}
			commentors = append(commentors, i)
		}

		for _, v := range followers {
			if v != int(commentDetail.Author.Int) {
				ns = append(ns, NewNotification("follow_report_reply", v))
			}
		}

		if len(commentors) > 0 {
			for _, v := range commentors {
				if v != int(commentDetail.Author.Int) {
					ns = append(ns, NewNotification("comment_comment", v))
				}
			}
		}

		if parentCommentDetail.Author.Valid && parentCommentDetail.Author.Int != commentDetail.Author.Int {
			ns = append(ns, NewNotification("comment_reply", int(parentCommentDetail.Author.Int)))
		}

		for k, v := range ns {
			v.SubjectID = strconv.Itoa(int(commentDetail.Author.Int))
			v.Nickname = commentDetail.AuthorNickname.String
			v.ProfileImage = commentDetail.AuthorImage.String
			v.ObjectName = report.Title.String
			v.ObjectType = resourceName
			v.ObjectID = strconv.Itoa(resourceID)
			v.ObjectSlug = report.Slug.String
			ns[k] = v
		}
		break
	default:

		if parentCommentDetail.Author.Valid && parentCommentDetail.Author.Int != commentDetail.Author.Int {
			ns = append(ns, NewNotification("comment_reply", int(parentCommentDetail.Author.Int)))
		}

		for k, v := range ns {
			v.SubjectID = strconv.Itoa(int(commentDetail.Author.Int))
			v.Nickname = commentDetail.AuthorNickname.String
			v.ProfileImage = commentDetail.AuthorImage.String
			v.ObjectName = commentDetail.Resource.String
			ns[k] = v
		}

		break
	}
	ns.Send()
	return err
}

func (c notificationGenerator) GenerateProjectNotifications(resource interface{}, resourceTyep string) (err error) { //memo, report, project
	ns := Notifications{}

	switch resourceTyep {
	case "project":
		p := resource.(Project)
		var eventType, tagEventType string

		if p.Status.Valid && p.Status.Int == int64(config.Config.Models.ProjectsStatus["done"]) {
			eventType = "follow_project_status"
			tagEventType = "follow_tag_project_status"
		} else if p.Progress.Valid {
			eventType = "follow_project_progress"
			tagEventType = "follow_tag_project_progress"
		} else {
			return nil
		}

		_ = c.generateTagNotifications(p, tagEventType)

		projectFollowers, err := c.getFollowers(p.ID, config.Config.Models.FollowingType["project"], []int{0})
		if err != nil {
			log.Println("Error get project followers", p.ID, err.Error())
		}

		for _, v := range projectFollowers {
			n := NewNotification(eventType, v)
			n.SubjectID = strconv.Itoa(p.ID)
			n.Nickname = p.Title.String
			n.ProfileImage = p.HeroImage.String
			n.ObjectName = p.Title.String
			n.ObjectType = "project"
			n.ObjectID = strconv.Itoa(p.ID)
			n.ObjectSlug = p.Slug.String
			ns = append(ns, n)
		}

	case "report":
		r := resource.(ReportAuthors)
		eventType := "follow_project_report"
		tagEventType := "follow_tag_report"

		projectFollowers, err := c.getFollowers(int(r.ProjectID.Int), config.Config.Models.FollowingType["project"], []int{0})
		if err != nil {
			log.Println("Error get project followers", r.ProjectID, err.Error())
		}

		project, err := ProjectAPI.GetProject(Project{ID: int(r.ProjectID.Int)})
		if err != nil {
			log.Println("Error get project", r.ProjectID, err.Error())
		}

		_ = c.generateTagNotifications(project, tagEventType)

		for _, v := range projectFollowers {
			n := NewNotification(eventType, v)
			n.SubjectID = strconv.Itoa(int(r.ID))
			n.Nickname = r.Title.String
			n.ProfileImage = project.HeroImage.String
			n.ObjectName = project.Title.String
			n.ObjectType = "project"
			n.ObjectID = strconv.Itoa(project.ID)
			n.ObjectSlug = project.Slug.String
			ns = append(ns, n)
		}

	case "memo":
		m := resource.(MemoDetail)
		eventType := "follow_project_memo"
		tagEventType := "follow_tag_memo"

		projectFollowers, err := c.getFollowers(int(m.Project.ID), config.Config.Models.FollowingType["project"], []int{0})
		if err != nil {
			log.Println("Error get project followers", m.Project.ID, err.Error())
		}

		_ = c.generateTagNotifications(m.Project.Project, tagEventType)

		for _, v := range projectFollowers {
			n := NewNotification(eventType, v)
			n.SubjectID = strconv.Itoa(int(m.ID))
			n.Nickname = m.Title.String
			n.ProfileImage = m.Project.HeroImage.String
			n.ObjectName = m.Project.Title.String
			n.ObjectType = "project"
			n.ObjectID = strconv.Itoa(m.Project.ID)
			n.ObjectSlug = m.Project.Slug.String
			ns = append(ns, n)
		}

	default:
	}

	ns.Send()
	return err
}

func (c notificationGenerator) GeneratePostNotifications(p TaggedPostMember) (err error) {
	ns := Notifications{}

	for _, tb := range p.Tags {

		tagFollowers, err := c.getFollowers(tb.ID, config.Config.Models.FollowingType["tag"], []int{0})
		if err != nil {
			log.Println("Error get tag followers", tb.ID, err.Error())
		}
		for _, v := range tagFollowers {
			if !c.checkForAuthor(v, p.Authors) && len(p.Authors) > 0 {
				n := NewNotification("follow_tag_post", v)
				n.SubjectID = strconv.Itoa(int(p.ID))
				n.Nickname = p.Title.String
				n.ProfileImage = p.Authors[0].ProfileImage.String
				n.ObjectName = tb.Text
				n.ObjectType = "tag"
				n.ObjectID = strconv.Itoa(tb.ID)
				ns = append(ns, n)
			}
		}
	}

	ns.Send()
	return err
}

func (c *notificationGenerator) generateTagNotifications(p Project, eventType string) (err error) {
	ns := Notifications{}
	query := "SELECT tags.tag_id, tags.tag_content FROM tagging LEFT JOIN tags ON tags.tag_id = tagging.tag_id WHERE type = ? AND target_id = ? AND active = ?"
	rows, err := rrsql.DB.Queryx(query, config.Config.Models.TaggingType["project"], p.ID, config.Config.Models.Tags["active"])
	if err != nil {
		log.Println("Error get project tags", p.ID, err.Error())
		return err
	}

	for rows.Next() {
		var tag Tag
		err = rows.StructScan(&tag)
		if err != nil {
			log.Println("Error scan tag into Tag struct when generate tag notifications:\n", err.Error())
			return err
		}

		tagFollowers, err := c.getFollowers(tag.ID, config.Config.Models.FollowingType["tag"], []int{0})
		if err != nil {
			log.Println("Error get tag followers", tag.ID, err.Error())
			return err
		}

		for _, v := range tagFollowers {
			n := NewNotification(eventType, v)
			n.SubjectID = strconv.Itoa(int(p.ID))
			n.Nickname = p.Title.String
			n.ProfileImage = p.HeroImage.String
			n.ObjectName = tag.Text
			n.ObjectType = "tag"
			n.ObjectID = strconv.Itoa(tag.ID)
			ns = append(ns, n)
		}
	}

	ns.Send()
	return err
}

func (c notificationGenerator) checkForAuthor(id int, authors []AuthorBasic) (isAuthor bool) {
	for _, author := range authors {
		if id == int(author.ID) {
			return true
		}
	}
	return false
}

var NotificationGen NotificationGenInterface = notificationGenerator{}
