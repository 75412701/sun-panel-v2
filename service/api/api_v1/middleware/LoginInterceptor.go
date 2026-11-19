package middleware

import (
	"sun-panel/api/api_v1/common/apiReturn"
	"sun-panel/global"
	"sun-panel/models"

	"github.com/gin-gonic/gin"
)

func LoginInterceptor(c *gin.Context) {
	// 获得token
	cToken := c.GetHeader("token")

	// 没有token信息视为未登录
	if cToken == "" {
		apiReturn.ErrorByCode(c, 1000)
		c.Abort()
		return
	}

	// 1. 第一级：极速内存查找 (O(1)，绝大多数正常请求直接命中，不消耗任何 DB 性能)
	token := cToken
	if mappedToken, ok := global.CUserToken.Get(cToken); ok && mappedToken != "" {
		token = mappedToken
	}

	if userInfo, success := global.UserToken.Get(token); success && userInfo.ID != 0 {
		c.Set("userInfo", userInfo)
		return
	}

	// 2. 第二级：内存未命中时，从数据库 user_token 表兜底查询（例如重启后新加入的或边缘情况）
	mUserToken := models.UserToken{}
	if user, _, err := mUserToken.GetUserByToken(cToken); err == nil && user.ID != 0 && user.Status == 1 {
		user.Password = ""
		user.ReferralCode = ""
		user.Token = cToken
		// 自动回填进内存缓存
		global.CUserToken.SetDefault(cToken, cToken)
		global.UserToken.SetDefault(cToken, user)
		c.Set("userInfo", user)
		return
	}

	// 3. 兼容旧版本：尝试从 users 表查询 legacy token
	mUser := models.User{}
	if info, err := mUser.GetUserInfoByToken(cToken); err == nil && info.ID != 0 && info.Status == 1 {
		info.Password = ""
		info.ReferralCode = ""
		// 自动迁移入 user_token 表
		_ = mUserToken.AddToken(info.ID, cToken)
		global.CUserToken.SetDefault(cToken, cToken)
		global.UserToken.SetDefault(cToken, info)
		c.Set("userInfo", info)
		return
	}

	// 否则视为真正过期
	apiReturn.ErrorByCode(c, 1001)
	c.Abort()
	return
}

// 不验证缓存直接验证库省去没有缓存每次都要手动登录的问题
func LoginInterceptorDev(c *gin.Context) {

	// 获得token
	token := c.GetHeader("token")
	mUser := models.User{}

	// 去库中查询是否存在该用户；否则返回错误
	if info, err := mUser.GetUserInfoByToken(token); err != nil || info.ID == 0 {
		apiReturn.ErrorCode(c, 1001, global.Lang.Get("login.err_token_expire"), nil)
		c.Abort()
		return
	} else {
		// 通过
		// 设置当前用户信息
		c.Set("userInfo", info)
	}
}
