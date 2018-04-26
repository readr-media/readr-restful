package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/readr-media/readr-restful/models"
)

type PermissionQueryItem struct {
	Role   int    `json:"role" binding:"required"`
	Object string `json:"object"  binding:"required"`
}

type PermissionQuery struct {
	Query []models.Permission `json:"query"`
}

type permissionHandler struct{}

func (r *permissionHandler) Get(c *gin.Context) {

	var input PermissionQuery
	err := c.ShouldBindJSON(&input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Bad Request. Empty Payload"})
		return
	}
	if !validatePermission(input) {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Bad Request"})
		return
	}

	permissions, err := models.PermissionAPI.GetPermissions(input.Query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "Internal Server Error"})
		return
	}

	var result []models.Permission

OuterLoop:
	for _, p := range input.Query {
		for _, permission := range permissions {
			if p.Role == permission.Role && p.Object == permission.Object {
				result = append(result, permission)
				continue OuterLoop
			}
		}
		p.Permission = models.NullInt{0, true}
		result = append(result, p)
	}

	c.JSON(http.StatusOK, result)
}

func (r *permissionHandler) GetAll(c *gin.Context) {

	permissions, err := models.PermissionAPI.GetPermissionsAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "Internal Server Error"})
		return
	}

	c.JSON(http.StatusOK, permissions)

}

func (r *permissionHandler) Post(c *gin.Context) {

	var input PermissionQuery
	c.ShouldBindJSON(&input)

	if !validatePermission(input) {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Bad Request"})
		return
	}

	err := models.PermissionAPI.InsertPermissions(input.Query)
	switch {
	case err != nil && err.Error() == "Duplicate Entry":
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Duplicate Entry"})
		return
	case err != nil:
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "Internal Server Error"})
		return
	}

	c.Status(http.StatusOK)
}

func (r *permissionHandler) Delete(c *gin.Context) {

	var input PermissionQuery
	c.ShouldBindJSON(&input)

	if !validatePermission(input) {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Bad Request"})
		return
	}

	err := models.PermissionAPI.DeletePermissions(input.Query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "Internal Server Error"})
		return
	}

	c.Status(http.StatusOK)
}

func (r *permissionHandler) SetRoutes(router *gin.Engine) {

	permissionRouter := router.Group("/permission")
	{
		permissionRouter.GET("", r.Get)
		permissionRouter.GET("/all", r.GetAll)
		permissionRouter.POST("", r.Post)
		permissionRouter.DELETE("", r.Delete)
	}
}

var PermissionHandler permissionHandler

func validatePermission(pq PermissionQuery) bool {
	for _, p := range pq.Query {
		if object, _ := p.Object.Value(); p.Role == 0 || object == nil {
			return false
		}
	}
	return true
}
