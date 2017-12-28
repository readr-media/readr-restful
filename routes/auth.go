package routes

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/readr-media/readr-restful/models"
	"golang.org/x/crypto/scrypt"
)

const (
	pw_salt_bytes = 32
	pw_hash_bytes = 64
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
	//fmt.Printf("id: %v, pwd: %v, mode: %v, p:%v", id, password, mode, p)

	switch {
	case !validateID(id), !validateMode(mode):
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Bad Request"})
		return
	case mode == "ordinary" && password == "":
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Bad Request"})
		return
	}

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

	if member.Active == 0 {
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

		hpassword, err := pwHash(password, member.Salt.String)
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

	userPermissions, err := models.PermissionAPI.GetPermissionsByRole(member.Role)
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
	ID           string `json:"id" db:"user_id"`
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
	// 2. check if user exists

	params := userRegisterParams{}
	err := c.Bind(&params)

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

	switch {
	case !validateID(member.ID), !validateMode(member.RegisterMode.String), !validateMail(member.Mail.String):
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Bad Request"})
		return
	}

	if member.RegisterMode.String == "ordinary" {

		if member.Password.String == "" {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Bad Request"})
			return
		}

		// 3. generate salt and hash password
		salt := make([]byte, pw_salt_bytes)
		_, err = io.ReadFull(rand.Reader, salt)
		if err != nil {
			log.Println(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Internal Server Error"})
			return
		}

		hpw, err := pwHash(member.Password.String, string(salt))
		if err != nil {
			log.Println(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Internal Server Error"})
			return
		}

		member.Salt = models.NullString{string(salt), true}
		member.Password = models.NullString{string(hpw), true}
		member.Active = 0

	} else {

		if member.SocialID.String != member.ID {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Bad Request"})
			return
		}

		member.Active = 1
	}

	// 4. create Member object, fill in data and defaults

	err = models.MemberAPI.InsertMember(member)

	if err != nil {
		switch err.Error() {
		case "Duplicate entry":
			c.JSON(http.StatusBadRequest, gin.H{"Error": "User Duplicated"})
			return
		case "More Than One Rows Affected", "No Row Inserted":
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Internal Server Error"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Internal Server Error"})
			return
		}
	}

	c.Status(http.StatusOK)
	return

}

func pwHash(pw, salt string) (string, error) {
	hpw, err := scrypt.Key([]byte(pw), []byte(salt), 32768, 8, 1, pw_hash_bytes)
	if err != nil {
		return "", err
	}
	return string(hpw), nil
}

func validateID(id string) bool {
	result := true
	if id == "" {
		result = false
	}
	return result
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
