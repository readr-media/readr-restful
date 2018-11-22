package routes

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/readr-media/readr-restful/config"
	"github.com/readr-media/readr-restful/models"
	"github.com/readr-media/readr-restful/utils"
)

type authHandler struct {
}

type userLoginParams struct {
	ID       string `json:"id"`
	Password string `json:"password"`
	Mode     string `json:"register_mode"`
}

func (r *authHandler) userLogin(c *gin.Context) {

	// 1. check input entry: id, password if uesr is not logged-in from OAuth
	// 2. get user by id, check if user exsists, check if user is active

	p := userLoginParams{}
	err := c.Bind(&p)
	if err != nil {
		fmt.Printf("%v", err)
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Bad Request"})
		return
	}

	id, password, mode := p.ID, p.Password, p.Mode
	// fmt.Printf("id: %v, pwd: %v, mode: %v, p:%v\n", id, password, mode, p)

	switch {
	case !utils.ValidateUserID(id), !validateMode(mode):
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Bad Request"})
		return
	case mode == "ordinary" && password == "":
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Bad Request"})
		return
	}

	member, err := models.MemberAPI.GetMember("member_id", id)
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

	if member.Active.Int <= 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"Error": "User Not Activated"})
		return
	}

	if mode != member.RegisterMode.String {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Bad Request"})
		return
	}

	// 3. Password mode: use salt to hash user's password, compare to password from db
	if mode == "ordinary" {
		if member.Salt.Valid == false {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "User Data Misconfigured"})
			return
		}

		hpassword, err := utils.CryptGenHash(password, member.Salt.String)
		if err != nil {
			log.Printf("error when hashing password: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Internal Server Error"})
			return
		}
		if member.Password.String != hpassword {
			c.JSON(http.StatusUnauthorized, gin.H{"Error": "Login Fail"})
			return
		}
	}
	// 4. get user permission by id
	// 5. return user's profile and permission info

	userPermissions, err := models.PermissionAPI.GetPermissionsByRole(int(member.Role.Int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "Internal Server Error"})
		return
	}

	var permissions []string
	for _, userPermission := range userPermissions[:] {
		permissions = append(permissions, userPermission.Object.String)
	}
	c.JSON(http.StatusOK, gin.H{
		"member":      member,
		"permissions": permissions})
	return
}

type userRegisterParams struct {
	Name         string `json:"name" db:"name"`
	Nickname     string `json:"nickname" db:"nick"`
	Gender       string `json:"gender" db:"gender"`
	Mail         string `json:"mail" db:"mail"`
	RegisterMode string `json:"register_mode" db:"register_mode"`
	SocialID     string `json:"social_id,omitempty" db:"social_id"`
	Password     string `json:"password" db:"password"`
}

func (r *authHandler) userRegister(c *gin.Context) {
	// 1. check input: account, mode, password, role ...

	params := userRegisterParams{}
	err := c.Bind(&params)
	if err != nil {
		log.Println("Bind parameter error: ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Bad Request"})
		return
	}

	switch {
	case !validateMode(params.RegisterMode), !validateMail(params.Mail):
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Bad Request"})
		return
	}

	// 2. check if user exists
	m, err := models.MemberAPI.GetMember("mail", params.Mail)
	if err == nil || err.Error() != "User Not Found" {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "User Duplicated", "Mode": m.RegisterMode.String})
		return
	}

	//Try to solve the problem that can't marshal password into models.Member struct
	jsonStr, err := json.Marshal(&params)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Bad Request"})
		return
	}

	var member models.Member
	err = json.Unmarshal(jsonStr, &member)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Bad Request"})
		return
	}
	member.Password = models.NullString{params.Password, true}

	if member.RegisterMode.String == "ordinary" {
		if member.Password.String == "" || member.Mail.String == "" {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Bad Request"})
			return
		}

		// 3. generate salt and hash password
		salt, err := utils.CryptGenSalt()
		if err != nil {
			log.Println(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Internal Server Error"})
			return
		}

		hpw, err := utils.CryptGenHash(member.Password.String, string(salt))
		if err != nil {
			log.Println(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Internal Server Error"})
			return
		}

		member.MemberID = member.Mail.String
		member.Salt = models.NullString{string(salt), true}
		member.Password = models.NullString{string(hpw), true}
		member.Active = models.NullInt{int64(config.Config.Models.Members["deactive"]), true}

	} else {

		if member.SocialID.String == "" {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Bad Request"})
			return
		}

		member.MemberID = member.SocialID.String
		member.Active = models.NullInt{int64(config.Config.Models.Members["active"]), true}
	}

	// 4. fill in data and defaults
	uuid, err := utils.NewUUIDv4()
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "Unable to generate uuid for user"})
		return
	}
	member.UUID = uuid.String()
	member.CreatedAt = models.NullTime{time.Now(), true}
	member.UpdatedAt = models.NullTime{time.Now(), true}
	member.Points = models.NullInt{0, true}
	member.Role = models.NullInt{1, true}

	lastID, err := models.MemberAPI.InsertMember(member)

	if err != nil {
		switch err.Error() {
		case "More Than One Rows Affected", "No Row Inserted":
			log.Println(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Internal Server Error"})
			return
		default:
			log.Println(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Internal Server Error"})
			return
		}
	}
	resp := map[string]int{"last_id": lastID}
	c.JSON(http.StatusOK, gin.H{"_items": resp})
	return
}

func validateMode(mode string) bool {
	result := true
	if mode != "ordinary" && mode != "oauth-fb" && mode != "oauth-goo" {
		result = false
	}
	return result
}

func validateMail(mail string) bool {
	result := true
	if mail == "" {
		result = false
	}
	return result
}

//func convertMemberStruct()

func (r *authHandler) SetRoutes(router *gin.Engine) {
	router.POST("/login", r.userLogin)
	router.POST("/register", r.userRegister)
}

var AuthHandler authHandler
