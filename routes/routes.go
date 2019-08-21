package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/readr-media/readr-restful/pkg/asset"
	"github.com/readr-media/readr-restful/pkg/cards"
	"github.com/readr-media/readr-restful/pkg/mail"
	"github.com/readr-media/readr-restful/pkg/poll"
	promotion "github.com/readr-media/readr-restful/pkg/promotion/http"
)

type RouterHandler interface {
	SetRoutes(router *gin.Engine)
}

func SetRoutes(router *gin.Engine) {
	for _, h := range []RouterHandler{
		&asset.Router,
		&AuthHandler,
		&CommentsHandler,
		&cards.Router,
		&FilterHandler,
		&FollowingHandler,
		&mail.Router,
		&MemberHandler,
		//&MemoHandler,
		&MiscHandler,
		&NotificationHandler,
		&PermissionHandler,
		&PointsHandler,
		&PostHandler,
		&ProjectHandler,
		&PubsubHandler,
		//&ReportHandler,
		&TagHandler,
		&poll.Router,
		&promotion.Router,
	} {
		h.SetRoutes(router)
	}
}
