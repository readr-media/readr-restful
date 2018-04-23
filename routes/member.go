package routes

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/readr-media/readr-restful/models"
	"github.com/readr-media/readr-restful/utils"
)

type memberHandler struct{}

func (r *memberHandler) bindQuery(c *gin.Context, args *models.MemberArgs) (err error) {

	args.SetDefault()
	if err = c.ShouldBindQuery(args); err != nil {
		// Return if error is other than Unknown type
		if err.Error() != "Unknown type" {
			return err
		}
	}
	if c.Query("active") != "" && args.Active == nil {
		if err := json.Unmarshal([]byte(c.Query("active")), &args.Active); err != nil {
			return err
		} else if err == nil {
			if err = models.ValidateActive(args.Active, models.MemberStatus); err != nil {
				return err
			}
		}
	}
	if c.Query("role") != "" && args.Role == nil {
		var role int64
		if role, err = strconv.ParseInt(c.Query("role"), 10, 64); err != nil {
			return err
		}
		args.Role = &role
	}
	if c.Query("uuids") != "" {
		if err = json.Unmarshal([]byte(c.Query("uuids")), &args.UUIDs); err != nil {
			return err
		}
	}
	return nil
}

func (r *memberHandler) GetAll(c *gin.Context) {

	var args = &models.MemberArgs{}
	if err := r.bindQuery(c, args); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	if args.Active == nil {
		args.DefaultActive()
	}
	result, err := models.MemberAPI.GetMembers(args)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"_items": result})
}

func (r *memberHandler) Get(c *gin.Context) {

	var idType string
	id := c.Param("id")
	if err := utils.ValidateUUID(id); err != nil {
		idType = "member_id"
	} else {
		idType = "uuid"
	}
	member, err := models.MemberAPI.GetMember(idType, id)
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
	if member.MemberID == "" {
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
	if !member.Active.Valid {
		member.Active = models.NullInt{1, true}
	}
	uuid, err := utils.NewUUIDv4()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "Unable to generate uuid for user"})
		return
	}
	member.UUID = uuid.String()
	err = models.MemberAPI.InsertMember(member)
	if err != nil {
		switch err.Error() {
		case "Duplicate entry":
			c.JSON(http.StatusBadRequest, gin.H{"Error": "User Already Existed"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
	}
	c.Status(http.StatusOK)
}

func (r *memberHandler) Put(c *gin.Context) {

	member := models.Member{}
	c.Bind(&member)
	// Use id field to check if Member Struct was binded successfully
	// If the binding failed, id would be emtpy string
	if member.MemberID == "" {
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

	err = models.MemberAPI.UpdateAll(ids, int(models.MemberStatus["delete"].(float64)))
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

	var idType string
	id := c.Param("id")
	if err := utils.ValidateUUID(id); err != nil {
		idType = "member_id"
	} else {
		idType = "uuid"
	}
	err := models.MemberAPI.DeleteMember(idType, id)
	if err != nil {
		switch err.Error() {
		case "User Not Found":
			c.JSON(http.StatusNotFound, gin.H{"Error": "User Not Found"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
	}
	c.Status(http.StatusOK)
}

func (r *memberHandler) ActivateAll(c *gin.Context) {
	payload := struct {
		IDs []string `json:"ids"`
	}{}
	err := c.Bind(&payload)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	if payload.IDs == nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Request Body"})
		return
	}
	err = models.MemberAPI.UpdateAll(payload.IDs, int(models.MemberStatus["active"].(float64)))
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
func (r *memberHandler) PutPassword(c *gin.Context) {

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
	member, err := models.MemberAPI.GetMember("member_id", input.ID)
	if err != nil {
		switch err.Error() {
		case "User Not Found":
			c.JSON(http.StatusNotFound, gin.H{"Error": "User Not Found"})
			return
		default:
			log.Println(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"Error": fmt.Sprintf("Internal Server Error. %s", err.Error())})
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
		c.JSON(http.StatusInternalServerError, gin.H{"Error": fmt.Sprintf("Internal Server Error. %s", err.Error())})
		return
	}

	hpw, err := utils.CryptGenHash(input.NewPassword, string(salt))
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"Error": fmt.Sprintf("Internal Server Error. %s", err.Error())})
		return
	}

	err = models.MemberAPI.UpdateMember(models.Member{
		MemberID: member.MemberID,
		Password: models.NullString{hpw, true},
		Salt:     models.NullString{salt, true},
	})
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"Error": fmt.Sprintf("Internal Server Error. %s", err.Error())})
		return
	}

	c.Status(http.StatusOK)
}

func (r *memberHandler) Count(c *gin.Context) {

	var args = &models.MemberArgs{}
	if err := r.bindQuery(c, args); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	if args.Active == nil {
		args.DefaultActive()
	}
	count, err := models.MemberAPI.Count(args)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
	resp := map[string]int{"total": count}
	c.JSON(http.StatusOK, gin.H{"_meta": resp})
}

func (r *memberHandler) SearchKeyNickname(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid keyword"})
		return
	}
	var roles map[string][]int

	if c.Query("roles") != "" {
		err := json.Unmarshal([]byte(c.Query("roles")), &roles)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid roles"})
			return
		}
	}

	members, err := models.MemberAPI.GetUUIDsByNickname(keyword, roles)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"_items": members})
}

func (r *memberHandler) SetRoutes(router *gin.Engine) {

	memberRouter := router.Group("/member")
	{
		memberRouter.GET("/:id", r.Get)
		memberRouter.POST("", r.Post)
		memberRouter.PUT("", r.Put)
		memberRouter.DELETE("/:id", r.Delete)

		memberRouter.PUT("/password", r.PutPassword)
	}
	membersRouter := router.Group("/members")
	{
		membersRouter.GET("", r.GetAll)
		membersRouter.PUT("", r.ActivateAll)
		membersRouter.DELETE("", r.DeleteAll)

		membersRouter.GET("/count", r.Count)
		membersRouter.GET("/nickname", r.SearchKeyNickname)
	}
}

var MemberHandler memberHandler
