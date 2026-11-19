package middleware

import (
	"sun-panel/api/api_v1/common/apiReturn"
	"sun-panel/api/api_v1/common/base"
	"sun-panel/global"
	"sun-panel/lib/cmn/systemSetting"
	"sun-panel/models"

	"github.com/gin-gonic/gin"
)

// 公开访问模式（访客模式）
// [有token将自动登录，无token/过期将使用公开账号，不可以与LoginInterceptor一起使用]
func PublicModeInterceptor(c *gin.Context) {
	// 获得token
	cToken := c.GetHeader("token")

	// 有 token 时尝试验证当前登录用户
	if cToken != "" {
		token := cToken
		if mappedToken, ok := global.CUserToken.Get(cToken); ok && mappedToken != "" {
			token = mappedToken
		}

		// 1. 第一级：内存缓存查找
		if userInfo, success := global.UserToken.Get(token); success && userInfo.ID != 0 {
			c.Set("userInfo", userInfo)
			return
		}

		// 2. 第二级：查 user_token 表
		mUserToken := models.UserToken{}
		if user, _, err := mUserToken.GetUserByToken(cToken); err == nil && user.ID != 0 && user.Status == 1 {
			user.Password = ""
			user.ReferralCode = ""
			user.Token = cToken
			global.CUserToken.SetDefault(cToken, cToken)
			global.UserToken.SetDefault(cToken, user)
			c.Set("userInfo", user)
			return
		}

		// 3. 兼容旧版 users 表
		mUser := models.User{}
		if info, err := mUser.GetUserInfoByToken(cToken); err == nil && info.ID != 0 && info.Status == 1 {
			info.Password = ""
			info.ReferralCode = ""
			_ = mUserToken.AddToken(info.ID, cToken)
			global.CUserToken.SetDefault(cToken, cToken)
			global.UserToken.SetDefault(cToken, info)
			c.Set("userInfo", info)
			return
		}
	}

	// 获取公开账号的信息
	var userId *uint
	if err := global.SystemSetting.GetValueByInterface(systemSetting.PANEL_PUBLIC_USER_ID, &userId); err == nil && userId != nil {
		userInfo := models.User{}
		if err := global.Db.First(&userInfo, "id=?", userId).Error; err != nil {
			apiReturn.ErrorCode(c, 1001, global.Lang.Get("login.err_token_expire"), nil)
			c.Abort()
			return
		}
		global.Logger.Debug("访客模式，客户端TOKEN:", cToken)
		global.Logger.Debug("访客用户ID:", userInfo.ID)
		c.Set("userInfo", userInfo)
		c.Set(base.GIN_GET_VISIT_MODE, base.VISIT_MODE_PUBLIC)
		return
	} else {
		global.Logger.Debug("访客用户不存在:", userId, " ", cToken)
		apiReturn.ErrorCode(c, 1001, global.Lang.Get("login.err_token_expire"), nil)
		c.Abort()
		return
	}

}
