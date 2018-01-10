package routes

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/readr-media/readr-restful/models"
	"github.com/readr-media/readr-restful/utils"
)

type memberHandler struct{}

func (r *memberHandler) GetAll(c *gin.Context) {

	mr := c.DefaultQuery("max_result", "20")
	u64MaxResult, _ := strconv.ParseUint(mr, 10, 8)
	maxResult := uint8(u64MaxResult)

	pg := c.DefaultQuery("page", "1")
	u64Page, _ := strconv.ParseUint(pg, 10, 16)
	page := uint16(u64Page)

	sorting := c.DefaultQuery("sort", "-updated_at")

	result, err := models.MemberAPI.GetMembers(maxResult, page, sorting)
	// fmt.Println(result)
	if err != nil {
		switch err.Error() {
		case "Members Not Found":
			c.JSON(http.StatusNotFound, gin.H{"Error": "Members Not Found"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Internal Server Error"})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"_items": result})
}
func (r *memberHandler) Get(c *gin.Context) {

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
	c.JSON(http.StatusOK, gin.H{"_items": []models.Member{member}})
}

func (r *memberHandler) Post(c *gin.Context) {

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
	// c.JSON(http.StatusOK, models.Member{})
	c.Status(http.StatusOK)
}

func (r *memberHandler) Put(c *gin.Context) {

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
	// c.JSON(http.StatusOK, models.Member{})
	c.Status(http.StatusOK)
}

func (r *memberHandler) DeleteAll(c *gin.Context) {
	ids := []string{}
	err := json.Unmarshal([]byte(c.Query("ids")), &ids)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	if len(ids) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "ID List Empty"})
		return
	}
	err = models.MemberAPI.SetMultipleActive(ids, 0)
	if err != nil {
		switch err.Error() {
		case "Members Not Found":
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Members Not Found"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
	}
	c.Status(http.StatusOK)
}

func (r *memberHandler) Delete(c *gin.Context) {

	// input := models.Member{ID: c.Param("id")}
	// var req models.Databox = &models.Member{ID: userID}
	// member, err := models.DS.Delete(input)
	id := c.Param("id")
	err := models.MemberAPI.DeleteMember(id)
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
	// c.JSON(http.StatusOK, member)
	c.Status(http.StatusOK)
}

func (r *memberHandler) ActivateAll(c *gin.Context) {
	payload := struct {
		MemberIDs []string `json:"member_ids"`
		Active    int      `json:"active"`
	}{}
	err := c.Bind(&payload)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	if payload.MemberIDs == nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Request Body"})
		return
	}
	err = models.MemberAPI.SetMultipleActive(payload.MemberIDs, payload.Active)
	if err != nil {
		switch err.Error() {
		case "Members Not Found":
			c.JSON(http.StatusNotFound, gin.H{"Error": "Members Not Found"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
	}
	c.Status(http.StatusOK)
}

// MemberPutPasswordHandler let caller to update a member's password.
//
func (r *memberHandler) MemberPutPasswordHandler(c *gin.Context) {

	input := struct {
		ID          string `json:"id"`
		NewPassword string `json:"password"`
		//OldPassword string `json="o"`
	}{}
	c.Bind(&input)

	if !utils.ValidateUserID(input.ID) || !utils.ValidatePassword(input.NewPassword) {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Input"})
		return
	}

	member, err := models.MemberAPI.GetMember(input.ID)
	if err != nil {
		switch err.Error() {
		case "User Not Found":
			c.JSON(http.StatusNotFound, gin.H{"Error": "User Not Found"})
			return
		default:
			log.Println(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Internal Server Error"})
			return
		}
	}

	/*
		oldHPW, err := utils.CryptGenHash(input.OldPassword, string(salt))
		if err != nil {
			log.Println(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Internal Server Error"})
			return
		}

		if oldHPW != member.Password.String {
			c.JSON(http.StatusUnauthorized, gin.H{"Error": "Wrong Password"})
		}
	*/

	// Gen salt and password
	salt, err := utils.CryptGenSalt()
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "Internal Server Error"})
		return
	}

	hpw, err := utils.CryptGenHash(input.NewPassword, string(salt))
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "Internal Server Error"})
		return
	}

	err = models.MemberAPI.UpdateMember(models.Member{
		ID:       member.ID,
		Password: models.NullString{hpw, true},
		Salt:     models.NullString{salt, true},
	})
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "Internal Server Error"})
		return
	}

	c.Status(http.StatusOK)
}

func (r *memberHandler) SetRoutes(router *gin.Engine) {

	// router.GET("/members", r.MembersGetHandler)

	memberRouter := router.Group("/member")
	{
		memberRouter.GET("/:id", r.Get)
		memberRouter.POST("", r.Post)
		memberRouter.PUT("", r.Put)
		memberRouter.DELETE("/:id", r.Delete)

		memberRouter.PUT("/password", r.MemberPutPasswordHandler)
	}
	membersRouter := router.Group("/members")
	{
		membersRouter.GET("", r.GetAll)
		membersRouter.PUT("", r.ActivateAll)
		membersRouter.DELETE("", r.DeleteAll)
	}
}

var MemberHandler memberHandler
