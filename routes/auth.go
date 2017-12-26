package routes

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/readr-media/readr-restful/models"
	"golang.org/x/crypto/scrypt"
)

type authHandler struct {
}

type userLoginParams struct {
	id       string
	password string
}

type userInfoResponse struct {
	member      models.Member
	permissions []string
}

func (r *authHandler) userLogin(c *gin.Context) {

	// 1. check input entry: id, password if uesr is not logged-in from OAuth
	// 2. get user by id, check if user exsists, check if user is active

	id, hasid := c.GetPostForm("id")
	password, haspw := c.GetPostForm("password")
	mode, hasmode := c.GetPostForm("mode")

	fmt.Printf("id: %v, pwd: %v, mode: %v", id, password, mode)

	if !hasid || !hasmode || (mode == "password" && !haspw) {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Bad Request"})
		return
	}

	if idValid := validateID(id); !idValid {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Bad Request"})
		return
	}

	if modeValid := validateMode(mode); !modeValid {
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
	if mode == "password" {
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
	var permissions []string
	for _, userPermission := range userPermissions[:] {
		permissions = append(permissions, userPermission.Object.String)
	}
	c.JSON(http.StatusOK, userInfoResponse{member, permissions})
	return
}

func pwHash(pw, salt string) (string, error) {
	fmt.Printf("%s, %s", pw, salt)
	hpw, err := scrypt.Key([]byte(pw), []byte(salt), 32768, 8, 1, 64)
	if err != nil {
		return "", err
	}
	return string(hpw), nil
}

func validateID(id string) bool {
	//email, googleid, fbid
	result := true
	if id == "" {
		result = false
	}
	return result
}

func validateMode(mode string) bool {
	//email, googleid, fbid
	result := true
	if mode != "password" && mode != "facebook" && mode != "google" {
		result = false
	}
	return result
}

func (r *authHandler) SetRoutes(router *gin.Engine) {
	router.POST("/login", r.userLogin)
}

var AuthHandler authHandler
