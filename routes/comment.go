package routes

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/readr-media/readr-restful/config"
	"github.com/readr-media/readr-restful/internal/rrsql"
	"github.com/readr-media/readr-restful/models"
)

type commentsHandler struct{}

func (r *commentsHandler) bindCommentQuery(c *gin.Context, args *models.GetCommentArgs) (err error) {
	if err = c.ShouldBindQuery(args); err != nil {
		// Return if error is other than Unknown type
		// Because binding status (map[string][]int) would give "Unknown Type" error.
		// We want to process the rest binding even if an error of Unknown Type is given
		if !strings.Contains(err.Error(), "Unknown type") && !strings.Contains(err.Error(), "strconv.ParseInt") {
			return err
		}
	}
	// Start parsing rest of request arguments
	if c.Query("author") != "" && args.Author == nil {
		if err = json.Unmarshal([]byte(c.Query("author")), &args.Author); err != nil {
			return err
		}
	}
	if c.Query("resource") != "" {
		if err = json.Unmarshal([]byte(c.Query("resource")), &args.Resource); err != nil {
			return err
		}
	}
	if c.Query("parent") != "" && args.Parent == nil {
		if err = json.Unmarshal([]byte(c.Query("parent")), &args.Parent); err != nil {
			return err
		}
	}
	if c.Query("status") != "" && args.Status == nil {
		if err = json.Unmarshal([]byte(c.Query("status")), &args.Status); err != nil {
			return err
		} else if err == nil {
			if err = rrsql.ValidateActive(args.Status, config.Config.Models.CommentStatus); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *commentsHandler) bindReportQuery(c *gin.Context, args *models.GetReportedCommentArgs) (err error) {
	if err = c.ShouldBindQuery(args); err == nil {
		return nil
	}
	// Start parsing rest of request arguments
	if c.Query("reporter") != "" && args.Reporter == nil {
		if err = json.Unmarshal([]byte(c.Query("reporter")), &args.Reporter); err != nil {
			return err
		}
	}
	if c.Query("solved") != "" && args.Solved == nil {
		if err = json.Unmarshal([]byte(c.Query("solved")), &args.Solved); err != nil {
			return err
		} else if err == nil {
			if err = rrsql.ValidateActive(args.Solved, config.Config.Models.ReportedCommentStatus); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *commentsHandler) GetComment(c *gin.Context) {

	id, _ := strconv.Atoi(c.Param("id"))
	result, err := models.CommentAPI.GetComment(id)

	if err != nil {
		switch err.Error() {
		case "Comment Not Found":
			c.JSON(http.StatusNotFound, gin.H{"Error": "Comment Not Found"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Internal Server Error"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"_items": result})
}

func (r *commentsHandler) GetComments(c *gin.Context) {

	args, _ := models.NewGetCommentArgs()
	if err := r.bindCommentQuery(c, args); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	result, err := models.CommentAPI.GetComments(args)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"_items": result})
}

func (r *commentsHandler) GetRC(c *gin.Context) {
	var args = &models.GetReportedCommentArgs{}
	args = args.Default()
	if err := r.bindReportQuery(c, args); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	result, err := models.CommentAPI.GetReportedComments(args)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"_items": result})
}

func (r *commentsHandler) PostRC(c *gin.Context) {
	report := models.ReportedComment{}
	err := c.Bind(&report)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	if report.CommentID.Int == 0 || !report.Reporter.Valid {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Missing Required Parameters."})
	}

	report.CreatedAt = rrsql.NullTime{Time: time.Now(), Valid: true}
	report.Solved = rrsql.NullInt{Int: int64(config.Config.Models.ReportedCommentStatus["pending"]), Valid: true}

	_, err = models.CommentAPI.InsertReportedComments(report)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

func (r *commentsHandler) PutRC(c *gin.Context) {
	report := models.ReportedComment{}
	err := c.Bind(&report)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	if report.ID == 0 || report.Reporter.Valid || report.CreatedAt.Valid {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Parameters"})
		return
	}

	report.UpdatedAt = rrsql.NullTime{Time: time.Now(), Valid: true}

	err = models.CommentAPI.UpdateReportedComments(report)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func (r *commentsHandler) UpdateCommentCounts(c *gin.Context) {
	err := models.CommentAPI.UpdateAllCommentAmount()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

func (r *commentsHandler) GetLatestComments(c *gin.Context) {
	comments, err := models.CommentCache.Obtain()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"_items": comments})
}

func (r *commentsHandler) UpdateLatestComments(c *gin.Context) {
	err := models.CommentCache.Generate()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func (r *commentsHandler) SetRoutes(router *gin.Engine) {
	commentRouter := router.Group("/comment")
	{
		commentRouter.GET("/:id", r.GetComment)
		commentRouter.GET("", r.GetComments)
		commentRouter.PUT("/counts", r.UpdateCommentCounts)
	}
	commentsRouter := router.Group("/comments")
	{
		commentsRouter.GET("/latest", r.GetLatestComments)
		commentsRouter.PUT("/latest", r.UpdateLatestComments)
	}
	reportcommentsRouter := router.Group("/reported_comment")
	{
		reportcommentsRouter.GET("", r.GetRC)
		reportcommentsRouter.POST("", r.PostRC)
		reportcommentsRouter.PUT("", r.PutRC)
	}
}

var CommentsHandler commentsHandler
