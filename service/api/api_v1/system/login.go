package system

import (
	"strconv"
	"strings"
	"sun-panel/api/api_v1/common/apiReturn"
	"sun-panel/api/api_v1/common/base"
	"sun-panel/global"
	"sun-panel/lib/cmn"
	"sun-panel/lib/cmn/systemSetting"
	"sun-panel/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LoginApi struct {
}

// 登录输入验证
type LoginLoginVerify struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required,max=50"`
	VCode    string `json:"vcode" validate:"max=6"`
	Email    string `json:"email"`
}

// @Summary 登录账号
// @Accept application/json
// @Produce application/json
// @Param LoginLoginVerify body LoginLoginVerify true "登陆验证信息"
// @Tags user
// @Router /login [post]
func (l LoginApi) Login(c *gin.Context) {
	param := LoginLoginVerify{}
	if err := c.ShouldBindJSON(&param); err != nil {
		apiReturn.ErrorParamFomat(c, err.Error())
		return
	}

	if errMsg, err := base.ValidateInputStruct(param); err != nil {
		apiReturn.ErrorParamFomat(c, errMsg)
		return
	}

	settings := systemSetting.ApplicationSetting{}
	global.SystemSetting.GetValueByInterface("system_application", &settings)

	mUser := models.User{}
	var (
		err  error
		info models.User
	)
	param.Username = strings.TrimSpace(param.Username)
	global.Logger.Infof("DEBUG LOGIN: Received Username=[%s], Password=[%s], PasswordLength=%d, TargetHash=[%s]", param.Username, param.Password, len(param.Password), cmn.PasswordEncryption(param.Password))

	// Check if user exists at all
	var userCheck models.User
	if err := global.Db.Where("username = ?", param.Username).First(&userCheck).Error; err != nil {
		global.Logger.Errorf("DEBUG LOGIN: Username [%s] not found in DB. Error: %v", param.Username, err)
	} else {
		global.Logger.Infof("DEBUG LOGIN: User [%s] found in DB. StoredHash=[%s]", param.Username, userCheck.Password)
	}

	if info, err = mUser.GetUserInfoByUsernameAndPassword(param.Username, cmn.PasswordEncryption(param.Password)); err != nil {
		// 未找到记录 账号或密码错误
		if err == gorm.ErrRecordNotFound {
			global.Logger.Warnf("DEBUG LOGIN: Login failed for [%s] - RECORD NOT FOUND with password index match.", param.Username)
			apiReturn.ErrorByCode(c, 1003)
			return
		} else {
			// 未知错误
			global.Logger.Errorf("DEBUG LOGIN: Database Error for [%s]: %v", param.Username, err)
			apiReturn.Error(c, err.Error())
			return
		}

	}

	global.Logger.Infof("DEBUG LOGIN: GetUserInfo SUCCESS. Found ID=[%d], Status=[%d]", info.ID, info.Status)

	// 停用或未激活
	if info.Status != 1 {
		global.Logger.Warnf("DEBUG LOGIN: User [%s] is DISABLED or INACTIVE. Status=[%d]. Returning 1004.", param.Username, info.Status)
		apiReturn.ErrorByCode(c, 1004)
		return
	}

	info.Password = ""
	info.ReferralCode = ""

	// 生成客户端唯一 Token
	cToken := uuid.NewString() + "-" + cmn.Md5(cmn.Md5("userId"+strconv.Itoa(int(info.ID))))

	// 1. 持久化到 user_token 数据库表，支持多设备同时在线与重启恢复
	mUserToken := models.UserToken{}
	if err := mUserToken.AddToken(info.ID, cToken); err != nil {
		global.Logger.Errorf("DEBUG LOGIN: Failed to insert user_token into DB: %v", err)
	}

	// 兼容老版本单字段
	_ = mUser.UpdateUserInfoByUserId(info.ID, map[string]interface{}{
		"token": cToken,
	})

	// 2. 存入内存缓存，保证后续接口 O(1) 极速直接命中
	global.CUserToken.SetDefault(cToken, cToken)
	global.UserToken.SetDefault(cToken, info)

	// 设置当前用户信息
	c.Set("userInfo", info)
	info.Token = cToken
	global.Logger.Infof("DEBUG LOGIN: ALL STEPS COMPLETE. Returning SuccessData for [%s]. Token=[%s...]", param.Username, cToken[:10])
	apiReturn.SuccessData(c, info)
}

// 安全退出
func (l *LoginApi) Logout(c *gin.Context) {
	cToken := c.GetHeader("token")
	if cToken != "" {
		// 1. 从内存缓存中删除
		global.CUserToken.Delete(cToken)
		global.UserToken.Delete(cToken)

		// 2. 从数据库 user_token 表中物理删除该设备的登录凭证
		mUserToken := models.UserToken{}
		_ = mUserToken.DeleteByToken(cToken)
	}
	apiReturn.Success(c)
}
