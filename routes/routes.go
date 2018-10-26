package routes

import "github.com/gin-gonic/gin"

import "github.com/readr-media/readr-restful/pkg/mail"

type RouterHandler interface {
	SetRoutes(router *gin.Engine)
}

func SetRoutes(router *gin.Engine) {
	for _, h := range []RouterHandler{
		&AuthHandler,
		&CommentsHandler,
		&FollowingHandler,
		//&MailHandler,
		&mail.Router,
		&MemberHandler,
		&MemoHandler,
		&MiscHandler,
		&NotificationHandler,
		&PermissionHandler,
		&PointsHandler,
		&PostHandler,
		&ProjectHandler,
		&PubsubHandler,
		&ReportHandler,
		&TagHandler,
	} {
		h.SetRoutes(router)
	}
}
