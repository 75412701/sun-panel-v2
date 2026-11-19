package userToken

import (
	"sun-panel/global"
	"sun-panel/lib/cache"
	"sun-panel/models"

	"time"
)

func InitUserToken() cache.Cacher[models.User] {
	// 设置为0表示永不过期
	return global.NewCache[models.User](0*time.Second, 1*time.Hour, "UserToken")
}

// PreloadTokensFromDB 服务启动时将数据库中所有有效 Token 预热加载到内存缓存中
func PreloadTokensFromDB() {
	mToken := models.UserToken{}
	tokens, userMap, err := mToken.GetAllActiveUserTokens()
	if err != nil {
		global.Logger.Errorf("PreloadTokensFromDB error: %v", err)
		return
	}

	loadedCount := 0
	for _, t := range tokens {
		if u, exists := userMap[t.UserId]; exists && u.Status == 1 {
			u.Password = ""
			u.ReferralCode = ""
			u.Token = t.Token
			global.CUserToken.SetDefault(t.Token, t.Token)
			global.UserToken.SetDefault(t.Token, u)
			loadedCount++
		}
	}

	// 兼容老数据：如果旧版 users 表中存在未迁移的 token，自动迁移至 user_token 表
	var legacyUsers []models.User
	if err := global.Db.Where("token != '' AND status = 1").Find(&legacyUsers).Error; err == nil {
		for _, u := range legacyUsers {
			if u.Token != "" {
				if _, exists := global.UserToken.Get(u.Token); !exists {
					_ = mToken.AddToken(u.ID, u.Token)
					u.Password = ""
					u.ReferralCode = ""
					global.CUserToken.SetDefault(u.Token, u.Token)
					global.UserToken.SetDefault(u.Token, u)
					loadedCount++
				}
			}
		}
	}

	global.Logger.Infof("User tokens preloaded into memory cache: %d token(s)", loadedCount)
}

// func InitVerifyCodeCachePool() {
// 	global.VerifyCodeCachePool = cache.NewGoCache(10*time.Minute, 60*time.Second)
// }
