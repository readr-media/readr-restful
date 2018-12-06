package poll

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type router struct{}

// GetPolls returns list of polls. If filter provided,
// it could return polls with corresponding choices
func (r *router) GetPolls(c *gin.Context) {

	filter, err := SetListPollsFilter(BindListPollsFilter(c))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	results, err := PollData.Get(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"_items": results})
}

// PostPolls accepts a single poll w/o choices
// and create new poll using input data
func (r *router) PostPolls(c *gin.Context) {

	poll := ChoicesEmbeddedPoll{}
	if err := c.Bind(&poll); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	// ----------- Start of input validation
	poll.Poll.Validate(
		ValidatePollInsertID,
		ValidatePollCreatedAt,
		ValidatePollUpdatedAt,
	)
	// Validate attached choices
	for i := range poll.Choices {
		poll.Choices[i].Validate(
			ValidateChoiceInsertID,
			ValidateChoiceCreatedAt,
			ValidateChoiceUpdatedAt,
		)
	} // --------- End of input validation
	if err := PollData.Insert(poll); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
	c.Status(http.StatusCreated)
}

// PutPolls updates contents of a single poll
func (r *router) PutPolls(c *gin.Context) {

	poll := Poll{}
	if err := c.Bind(&poll); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	// validation sections
	poll.Validate(ValidatePollUpdatedAt)
	err := PollData.Update(poll)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (r *router) GetChoices(c *gin.Context) {

	pollID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	results, err := ChoiceData.Get(int(pollID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"_items": results})
}

func (r *router) InsertChoices(c *gin.Context) {

	pollID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var input struct {
		Choices []Choice `json:"choices"`
	}
	if err := c.Bind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	// ------------ Start of input validation
	for i := range input.Choices {
		input.Choices[i].Validate(
			ValidateChoiceInsertID,
			ValidateChoiceCreatedAt,
			ValidateChoiceUpdatedAt,
			ValidateChoicePollID(pollID))
	} // ---------- End of input validation
	if err := ChoiceData.Insert(input.Choices); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
	c.Status(http.StatusCreated)
}

func (r *router) PutChoices(c *gin.Context) {

	pollID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var input struct {
		Choices []Choice `json:"choices"`
	}
	if err := c.Bind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	for i := range input.Choices {
		input.Choices[i].Validate(
			ValidateChoicePollID(pollID),
			ValidateChoiceUpdatedAt,
		)
	}
	if err := ChoiceData.Update(input.Choices); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (r *router) GetPicks(c *gin.Context) {
	filter, err := SetListPicksFilter(BindListPicksFilter(c))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	picks, err := PickData.Get(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"_items": picks})
}

func (r *router) InsertPicks(c *gin.Context) {
	c.Status(http.StatusCreated)
}

func (r *router) UpdatePicks(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// PubsubMessage hosts the message body of Google Pub/Sub
type PubsubMessage struct {
	ID   string            `json:"messageId"`
	Body []byte            `json:"data"`
	Attr map[string]string `json:"attributes"`
}

// PubsubData is total message format of Google Pub/Sub
type PubsubData struct {
	Subscription string `json:"subscription"`
	Message      PubsubMessage
}

func (r *router) Push(c *gin.Context) {

	defer c.Status(http.StatusOK)

	var (
		input PubsubData
		pick  ChosenChoice
		err   error
	)
	c.ShouldBindJSON(&input)
	action := input.Message.Attr["action"]
	if err = json.Unmarshal(input.Message.Body, &pick); err != nil {
		log.Printf("Fail to parse %s message: %s\n", action, err.Error())
		return
	}
	if !pick.CreatedAt.Valid {
		pick.CreatedAt.Time = time.Now()
		pick.CreatedAt.Valid = true
	}
	switch action {
	case "insert":
		if err := PickData.Insert(pick); err != nil {
			log.Printf("Insert error:%s\n", err.Error())
			return
		}
	case "update":
		if err := PickData.Update(pick); err != nil {
			log.Printf("Update error:%s\n", err.Error())
			return
		}
	default:
		log.Println("Pubsub Message Action Not Support", action)
		return
	}
}

func (r *router) SetRoutes(router *gin.Engine) {

	v2 := router.Group("/v2/polls")
	{
		v2.GET("", r.GetPolls)
		v2.POST("", r.PostPolls)
		v2.PUT("", r.PutPolls)

		v2.GET("/:id/choices", r.GetChoices)
		v2.POST("/:id/choices", r.InsertChoices)
		v2.PUT("/:id/choices", r.PutChoices)

		v2.GET("/:id/picks", r.GetPicks)
		v2.POST("/:id/picks", r.InsertPicks)
		v2.PUT("/:id/picks", r.UpdatePicks)

	}
	// "/v2/polls/pubsub" collides with "/v2/polls/:id"
	// It's natural restriction with httprouter. It might be solved in v2
	// Now use "/v2/pubsub/polls/" to avoid collision
	router.POST("/v2/pubsub/polls", r.Push)
}

// Router is the single routing instance used in registration in routes/routes.go
var Router router
