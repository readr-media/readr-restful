package routes

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/readr-media/readr-restful/models"
)

type memberHandler struct{}

func (r *memberHandler) MemberGetHandler(c *gin.Context) {

	// input := models.Member{ID: c.Param("id")}
	// member, err := models.DS.Get(input)
	id := c.Param("id")
	member, err := models.MemberAPI.GetMember(id)
	if err != nil {
		switch err.Error() {
		case "User Not Found":
			c.JSON(http.StatusNotFound, gin.H{"Error": "User Not Found"})
			return

		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Internal Server Error"})
			return
		}
	}
	c.JSON(http.StatusOK, member)
}

func (r *memberHandler) MemberPostHandler(c *gin.Context) {

	member := models.Member{}
	c.Bind(&member)

	// Pre-request test
	if member.ID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid User"})
		return
	}
	if !member.CreatedAt.Valid {
		member.CreatedAt.Time = time.Now()
		member.CreatedAt.Valid = true
	}
	if !member.UpdatedAt.Valid {
		member.UpdatedAt.Time = time.Now()
		member.UpdatedAt.Valid = true
	}

	err := models.MemberAPI.InsertMember(member)
	// result, err := models.DS.Create(member)
	// var req models.Databox = &member
	// result, err := req.Create()
	if err != nil {
		switch err.Error() {
		case "Duplicate entry":
			c.JSON(http.StatusBadRequest, gin.H{"Error": "User Already Existed"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Internal Server Error"})
			return
		}
	}
	c.JSON(http.StatusOK, models.Member{})
}

func (r *memberHandler) MemberPutHandler(c *gin.Context) {

	member := models.Member{}
	c.Bind(&member)
	// Use id field to check if Member Struct was binded successfully
	// If the binding failed, id would be emtpy string
	if member.ID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Member Data"})
		return
	}
	if member.CreatedAt.Valid {
		member.CreatedAt.Time = time.Time{}
		member.CreatedAt.Valid = false
	}
	if !member.UpdatedAt.Valid {
		member.UpdatedAt.Time = time.Now()
		member.UpdatedAt.Valid = true
	}
	// var req models.Databox = &member
	// result, err := req.Update()
	// result, err := models.DS.Update(member)
	err := models.MemberAPI.UpdateMember(member)
	if err != nil {
		switch err.Error() {
		case "User Not Found":
			c.JSON(http.StatusBadRequest, gin.H{"Error": "User Not Found"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Internal Server Error"})
			return
		}
	}
	c.JSON(http.StatusOK, models.Member{})
}

func (r *memberHandler) MemberDeleteHandler(c *gin.Context) {

	// input := models.Member{ID: c.Param("id")}
	// var req models.Databox = &models.Member{ID: userID}
	// member, err := models.DS.Delete(input)
	id := c.Param("id")
	member, err := models.MemberAPI.DeleteMember(id)
	// member, err := req.Delete()
	if err != nil {
		switch err.Error() {
		case "User Not Found":
			c.JSON(http.StatusNotFound, gin.H{"Error": "User Not Found"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Internal Server Error"})
			return
		}
	}
	c.JSON(http.StatusOK, member)
}

func (r *memberHandler) SetRoutes(router *gin.Engine) {
	memberRouter := router.Group("/member")
	{
		memberRouter.GET("/:id", r.MemberGetHandler)
		memberRouter.POST("", r.MemberPostHandler)
		memberRouter.PUT("", r.MemberPutHandler)
		memberRouter.DELETE("/:id", r.MemberDeleteHandler)
	}
}

var MemberHandler memberHandler
