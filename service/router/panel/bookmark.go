package panel

import (
	"sun-panel/api/api_v1"
	"sun-panel/api/api_v1/middleware"

	"github.com/gin-gonic/gin"
)

// InitBookmark 初始化书签相关路由
func InitBookmark(router *gin.RouterGroup) {
	bookmarkApi := api_v1.ApiGroupApp.ApiPanel.Bookmark

	// 需要登录的路由
	private := router.Group("", middleware.LoginInterceptor)
	{
		private.POST("/panel/bookmark/addMultiple", bookmarkApi.AddMultiple)
		private.POST("/panel/bookmark/add", bookmarkApi.Add)
		private.POST("/panel/bookmark/update", bookmarkApi.Update)
		private.POST("/panel/bookmark/deletes", bookmarkApi.Deletes)
	}

	// 公开模式下可以获取书签列表
	public := router.Group("", middleware.PublicModeInterceptor)
	{
		public.POST("/panel/bookmark/getList", bookmarkApi.GetList)
	}
}
