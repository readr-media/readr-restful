package routes

import "github.com/gin-gonic/gin"

type RouterHandler interface {
	SetRoutes(router *gin.Engine)
}

func SetRoutes(router *gin.Engine) {
	for _, h := range []RouterHandler{
		&AuthHandler,
		&CommentsHandler,
		&FollowingHandler,
		&MailHandler,
		&MemberHandler,
		&MemoHandler,
		&MiscHandler,
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
