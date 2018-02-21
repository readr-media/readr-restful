package routes

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/readr-media/readr-restful/models"
	"github.com/readr-media/readr-restful/utils"
)

type memberHandler struct{}

func (r *memberHandler) bindQuery(c *gin.Context, args *models.MemberArgs) (err error) {

	if err = c.ShouldBindQuery(args); err == nil {
		// No active pass in parameter. Set default
		args.DefaultActive()
		return nil
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
	return nil
}

func (r *memberHandler) GetAll(c *gin.Context) {

	var args = &models.MemberArgs{}
	args = args.Default()
	if err := r.bindQuery(c, args); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	result, err := models.MemberAPI.GetMembers(args)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"_items": result})
}

func (r *memberHandler) Get(c *gin.Context) {

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
	if !member.Active.Valid {
		member.Active = models.NullInt{1, true}
	}

	err := models.MemberAPI.InsertMember(member)
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

	id := c.Param("id")
	err := models.MemberAPI.DeleteMember(id)
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

func (r *memberHandler) Count(c *gin.Context) {

	var args = &models.MemberArgs{}
	args = args.Default()
	if err := r.bindQuery(c, args); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	count, err := models.MemberAPI.Count(args)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
	resp := map[string]int{"total": count}
	c.JSON(http.StatusOK, gin.H{"_meta": resp})
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
	}
}

var MemberHandler memberHandler
