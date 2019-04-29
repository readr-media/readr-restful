package cards

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/readr-media/readr-restful/config"
	"github.com/readr-media/readr-restful/models"
)

type newscardHandler struct{}

func (r *newscardHandler) bindQuery(c *gin.Context, args *NewsCardArgs) (err error) {
	_ = c.ShouldBindQuery(args)

	// Start parsing rest of request arguments
	if c.Query("sorting") != "" {
		sorting := c.Query("sorting")
		if !args.validateSorting(sorting) {
			return errors.New("Invalid Sorting Value")
		}
		args.Sorting = sorting
	}
	if c.Query("ids") != "" {
		if err = json.Unmarshal([]byte(c.Query("ids")), &args.IDs); err != nil {
			return err
		}
	}
	if c.Query("active") != "" {
		args.Active = nil
		if err = json.Unmarshal([]byte(c.Query("active")), &args.Active); err != nil {
			return err
		} else if err == nil {
			if err = models.ValidateActive(args.Active, config.Config.Models.Cards); err != nil {
				return err
			}
		}
	}
	if c.Query("status") != "" {
		args.Status = nil
		if err = json.Unmarshal([]byte(c.Query("status")), &args.Status); err != nil {
			return err
		} else if err == nil {
			if err = models.ValidateActive(args.Status, config.Config.Models.CardStatus); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *newscardHandler) Get(c *gin.Context) {

	iduint64, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	id := uint32(iduint64)

	args := DefaultNewsCardArgs()
	args.IDs = []uint32{id}

	cards, err := NewsCardAPI.GetCards(args)

	if err != nil {
		switch err.Error() {
		case "Card Not Found":
			c.JSON(http.StatusNotFound, gin.H{"Error": "Card Not Found"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Internal Server Error"})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"_items": cards})
}

func (r *newscardHandler) GetAll(c *gin.Context) {

	args := DefaultNewsCardArgs()
	if err := r.bindQuery(c, args); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	cards, err := NewsCardAPI.GetCards(args)

	if err != nil {
		switch err.Error() {
		case "Card Not Found":
			c.JSON(http.StatusNotFound, gin.H{"Error": "Card Not Found"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Internal Server Error"})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"_items": cards})
}

func (r *newscardHandler) Post(c *gin.Context) {

	card := NewsCard{}

	err := c.Bind(&card)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	if card == (NewsCard{}) {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Card"})
		return
	}
	if card.PostID == 0 || !card.Title.Valid {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Title or CardID"})
		return
	}

	// CreatedAt and UpdatedAt set default to now
	card.CreatedAt = models.NullTime{Time: time.Now(), Valid: true}
	card.UpdatedAt = models.NullTime{Time: time.Now(), Valid: true}

	if !card.Active.Valid {
		card.Active = models.NullInt{int64(config.Config.Models.Cards["inctive"]), true}
	}
	if !card.Status.Valid {
		card.Status = models.NullInt{int64(config.Config.Models.CardStatus["draft"]), true}
	}

	_, err = NewsCardAPI.InsertCard(card)
	if err != nil {
		switch err.Error() {
		case "Duplicate entry":
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Card ID Already Taken"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
	}
	c.Status(http.StatusOK)
}

func (r *newscardHandler) Put(c *gin.Context) {

	card := NewsCard{}

	err := c.ShouldBindJSON(&card)
	// Check if post struct was binded successfully
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	if card == (NewsCard{}) {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Card"})
		return
	}

	// Discard CreatedAt even if there is data
	if card.CreatedAt.Valid {
		card.CreatedAt.Time = time.Time{}
		card.CreatedAt.Valid = false
	}

	card.UpdatedAt = models.NullTime{Time: time.Now(), Valid: true}

	err = NewsCardAPI.UpdateCard(card)
	if err != nil {
		switch err.Error() {
		case "Card Not Found":
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Card Not Found"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
	}

	c.Status(http.StatusOK)
}

func (r *newscardHandler) Delete(c *gin.Context) {

	iduint64, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	id := uint32(iduint64)

	err := NewsCardAPI.DeleteCard(id)
	if err != nil {

		switch err.Error() {
		case "Card Not Found":
			c.JSON(http.StatusNotFound, gin.H{"Error": "Card Not Found"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Internal Server Error"})
			return
		}
	}
	c.Status(http.StatusOK)
}

func (r *newscardHandler) SetRoutes(router *gin.Engine) {

	cardRouter := router.Group("/cards")
	{
		cardRouter.GET("/:id", r.Get)
		cardRouter.GET("", r.GetAll)
		cardRouter.POST("", r.Post)
		cardRouter.PUT("", r.Put)
		cardRouter.DELETE("/:id", r.Delete)
	}
}

var Router newscardHandler
