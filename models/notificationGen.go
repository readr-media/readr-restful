package models

import (
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/readr-media/readr-restful/config"
)

type notificationGenerator struct{}

func (c *notificationGenerator) getFollowers(resourceID int, resourceType int) (followers []int, err error) {
	fInterface, err := FollowingAPI.Get(&GetFollowerMemberIDsArgs{int64(resourceID), resourceType})
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

func (c *notificationGenerator) GenerateCommentNotifications(comment InsertCommentArgs) (err error) {
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
		post, err := PostAPI.GetPost(uint32(resourceID))
		if err != nil {
			log.Println("Error get post", resourceID, err.Error())
		}

		postFollowers, err := c.getFollowers(resourceID, 2)
		if err != nil {
			log.Println("Error get post followers", resourceID, err.Error())
		}
		//log.Println(postFollowers)

		authorFollowers, err := c.getFollowers(int(post.Author.Int), 1)
		if err != nil {
			log.Println("Error get author followers", post.Author.Int, err.Error())
		}
		//log.Println(authorFollowers)

		var commentors []int
		// rows, err := DB.Queryx(fmt.Sprintf(`SELECT DISTINCT author FROM comments WHERE resource="%s" AND status = %d AND active = %d;`, commentDetail.Resource.String, int(CommentStatus["show"].(float64)), int(CommentActive["active"].(float64))))
		rows, err := DB.Queryx(fmt.Sprintf(`SELECT DISTINCT author FROM comments WHERE resource="%s" AND status = %d AND active = %d;`, commentDetail.Resource.String, config.Config.Models.CommentStatus["show"], config.Config.Models.Comment["active"]))

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

		for _, v := range authorFollowers {
			if v != int(commentDetail.Author.Int) {
				ns = append(ns, NewNotification("follow_member_reply", v))
			}
		}

		if commentDetail.Author.Int != post.Author.Int {
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
			if commentDetail.Author.Int == post.Author.Int {
				ns = append(ns, NewNotification("comment_reply_author", int(parentCommentDetail.Author.Int)))
			} else {
				ns = append(ns, NewNotification("comment_reply", int(parentCommentDetail.Author.Int)))
			}
		}

		for k, v := range ns {
			v.SubjectID = int(commentDetail.Author.Int)
			v.Nickname = commentDetail.AuthorNickname.String
			v.ProfileImage = commentDetail.AuthorImage.String
			v.ObjectName = post.Member.Nickname.String
			v.ObjectType = resourceName
			v.ObjectID = resourceID
			v.PostType = int(post.Type.Int)
			ns[k] = v
		}

		break

	case "project":
		project, err := ProjectAPI.GetProject(Project{ID: resourceID})
		if err != nil {
			log.Println("Error get project", resourceID, err.Error())
		}

		projectFollowers, err := c.getFollowers(resourceID, 3)
		if err != nil {
			log.Println("Error get project followers", resourceID, err.Error())
		}
		log.Println(projectFollowers)

		var commentors []int
		// rows, err := DB.Queryx(fmt.Sprintf(`SELECT DISTINCT author FROM comments WHERE resource="%s" AND status = %d AND active = %d;`, commentDetail.Resource.String, int(CommentStatus["show"].(float64)), int(CommentActive["active"].(float64))))
		rows, err := DB.Queryx(fmt.Sprintf(`SELECT DISTINCT author FROM comments WHERE resource="%s" AND status = %d AND active = %d;`, commentDetail.Resource.String, config.Config.Models.CommentStatus["show"], config.Config.Models.Comment["active"]))
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
			v.SubjectID = int(commentDetail.Author.Int)
			v.Nickname = commentDetail.AuthorNickname.String
			v.ProfileImage = commentDetail.AuthorImage.String
			v.ObjectName = project.Title.String
			v.ObjectType = resourceName
			v.ObjectID = resourceID
			v.ObjectSlug = project.Slug.String
			v.PostType = 0
			ns[k] = v
		}

		break

	case "memo":
		memo, err := MemoAPI.GetMemo(resourceID)
		if err != nil {
			log.Println("Error get memo", resourceID, err.Error())
		}
		project, err := ProjectAPI.GetProject(Project{ID: int(memo.Project.Int)})
		if err != nil {
			log.Println("Error get project", memo.Project.Int, err.Error())
		}

		projectFollowers, err := c.getFollowers(int(memo.Project.Int), 3)
		if err != nil {
			log.Println("Error get project followers", memo.Project.Int, err.Error())
		}
		memoFollowers, err := c.getFollowers(resourceID, 4)
		if err != nil {
			log.Println("Error get project followers", resourceID, err.Error())
		}

		followers := c.mergeFollowerSlices(memoFollowers, projectFollowers)

		var commentors []int
		// rows, err := DB.Queryx(fmt.Sprintf(`SELECT DISTINCT author FROM comments WHERE resource="%s" AND status = %d AND active = %d;`, commentDetail.Resource.String, int(CommentStatus["show"].(float64)), int(CommentActive["active"].(float64))))
		rows, err := DB.Queryx(fmt.Sprintf(`SELECT DISTINCT author FROM comments WHERE resource="%s" AND status = %d AND active = %d;`, commentDetail.Resource.String, config.Config.Models.CommentStatus["show"], config.Config.Models.Comment["active"]))
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
			v.SubjectID = int(commentDetail.Author.Int)
			v.Nickname = commentDetail.AuthorNickname.String
			v.ProfileImage = commentDetail.AuthorImage.String
			v.ObjectName = memo.Title.String
			v.ObjectType = resourceName
			v.ObjectID = resourceID
			v.ObjectSlug = project.Slug.String
			ns[k] = v
		}
		break

	case "report":
		report, err := ReportAPI.GetReport(Report{ID: resourceID})
		if err != nil {
			log.Println("Error get memo", resourceID, err.Error())
		}
		projectFollowers, err := c.getFollowers(report.ProjectID, 3)
		if err != nil {
			log.Println("Error get project followers", report.ProjectID, err.Error())
		}
		reportFollowers, err := c.getFollowers(resourceID, 5)
		if err != nil {
			log.Println("Error get report followers", resourceID, err.Error())
		}

		followers := c.mergeFollowerSlices(reportFollowers, projectFollowers)

		var commentors []int
		// rows, err := DB.Queryx(fmt.Sprintf(`SELECT DISTINCT author FROM comments WHERE resource="%s" AND status = %d AND active = %d;`, commentDetail.Resource.String, int(CommentStatus["show"].(float64)), int(CommentActive["active"].(float64))))
		rows, err := DB.Queryx(fmt.Sprintf(`SELECT DISTINCT author FROM comments WHERE resource="%s" AND status = %d AND active = %d;`, commentDetail.Resource.String, config.Config.Models.CommentStatus["show"], config.Config.Models.Comment["active"]))
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
			v.SubjectID = int(commentDetail.Author.Int)
			v.Nickname = commentDetail.AuthorNickname.String
			v.ProfileImage = commentDetail.AuthorImage.String
			v.ObjectName = report.Title.String
			v.ObjectType = resourceName
			v.ObjectID = resourceID
			v.ObjectSlug = report.Slug.String
			ns[k] = v
		}
		break
	default:

		if parentCommentDetail.Author.Valid && parentCommentDetail.Author.Int != commentDetail.Author.Int {
			ns = append(ns, NewNotification("comment_reply", int(parentCommentDetail.Author.Int)))
		}

		for k, v := range ns {
			v.SubjectID = int(commentDetail.Author.Int)
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

func (c *notificationGenerator) GenerateProjectNotifications(resource interface{}, resourceTyep string) (err error) { //memo, report, project
	ns := Notifications{}

	switch resourceTyep {
	case "project":
		p := resource.(Project)
		eventType := ""

		if p.PublishStatus.Valid && p.PublishStatus.Int == int64(config.Config.Models.ProjectsPublishStatus["publish"]) {
			eventType = "follow_project_status"
		} else if p.Progress.Valid {
			eventType = "follow_project_progress"
		}

		projectFollowers, err := c.getFollowers(p.ID, 3)
		if err != nil {
			log.Println("Error get project followers", p.ID, err.Error())
		}

		for _, v := range projectFollowers {
			ns = append(ns, NewNotification(eventType, v))
		}

		for k, v := range ns {
			v.SubjectID = p.ID
			v.Nickname = p.Title.String
			v.ProfileImage = p.HeroImage.String
			v.ObjectName = p.Title.String
			v.ObjectType = "project"
			v.ObjectID = p.ID
			v.ObjectSlug = p.Slug.String
			ns[k] = v
		}
	case "report":
		log.Println("ok")
		r := resource.(ReportAuthors)
		eventType := "follow_project_report"

		projectFollowers, err := c.getFollowers(r.ProjectID, 3)
		if err != nil {
			log.Println("Error get project followers", r.ProjectID, err.Error())
		}

		log.Println(projectFollowers)

		project, err := ProjectAPI.GetProject(Project{ID: r.ProjectID})
		if err != nil {
			log.Println("Error get project", r.ProjectID, err.Error())
		}

		for _, v := range projectFollowers {
			ns = append(ns, NewNotification(eventType, v))
		}

		for k, v := range ns {
			v.SubjectID = r.ID
			v.Nickname = r.Title.String
			v.ProfileImage = project.HeroImage.String
			v.ObjectName = project.Title.String
			v.ObjectType = "project"
			v.ObjectID = project.ID
			v.ObjectSlug = project.Slug.String
			ns[k] = v
		}
	case "memo":
		m := resource.(Memo)
		eventType := "follow_project_memo"

		projectFollowers, err := c.getFollowers(int(m.Project.Int), 3)
		if err != nil {
			log.Println("Error get project followers", m.Project.Int, err.Error())
		}

		project, err := ProjectAPI.GetProject(Project{ID: int(m.Project.Int)})
		if err != nil {
			log.Println("Error get project", m.Project.Int, err.Error())
		}

		for _, v := range projectFollowers {
			ns = append(ns, NewNotification(eventType, v))
		}

		for k, v := range ns {
			v.SubjectID = m.ID
			v.Nickname = m.Title.String
			v.ProfileImage = project.HeroImage.String
			v.ObjectName = project.Title.String
			v.ObjectType = "project"
			v.ObjectID = project.ID
			v.ObjectSlug = project.Slug.String
			ns[k] = v
		}
	default:
	}

	ns.Send()
	return err
}

func (c *notificationGenerator) GeneratePostNotifications(p TaggedPostMember) (err error) {
	ns := Notifications{}

	authorFollowers, err := c.getFollowers(int(p.Author.Int), 1)
	if err != nil {
		log.Println("Error get post followers", p.Author.Int, err.Error())
	}
	for _, v := range authorFollowers {
		if v != int(p.Author.Int) {
			ns = append(ns, NewNotification("follow_member_post", v))
		}
	}

	authorInfo, err := MemberAPI.GetMember("ID", strconv.Itoa(int(p.Author.Int)))
	if err != nil {
		log.Println("Error get post author", p.Author.Int, err.Error())
	}

	for k, v := range ns {
		v.SubjectID = int(p.ID)
		v.Nickname = p.Title.String
		v.ProfileImage = authorInfo.ProfileImage.String
		v.ObjectName = authorInfo.Nickname.String
		v.ObjectType = "member"
		v.ObjectID = int(p.Author.Int)
		v.PostType = int(p.Type.Int)
		ns[k] = v
	}
	ns.Send()
	return err
}

var NotificationGen notificationGenerator = notificationGenerator{}
