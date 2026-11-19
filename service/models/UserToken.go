package models

// 用户Token表 - 用于持久化多设备登录凭证
type UserToken struct {
	BaseModelNoUpdated
	UserId uint   `gorm:"index;type:int(11)" json:"userId"`
	Token  string `gorm:"uniqueIndex;type:varchar(64)" json:"token"`
}

// GetUserByToken 根据 Token 查询用户与 Token 记录
func (m *UserToken) GetUserByToken(token string) (User, UserToken, error) {
	var userToken UserToken
	if err := Db.Where("token = ?", token).First(&userToken).Error; err != nil {
		return User{}, UserToken{}, err
	}

	var user User
	if err := Db.Where("id = ? AND status = 1", userToken.UserId).First(&user).Error; err != nil {
		return User{}, userToken, err
	}

	return user, userToken, nil
}

// AddToken 添加 Token 记录
func (m *UserToken) AddToken(userId uint, token string) error {
	record := UserToken{
		UserId: userId,
		Token:  token,
	}
	return Db.Create(&record).Error
}

// DeleteByToken 删除指定 Token (单设备退出)
func (m *UserToken) DeleteByToken(token string) error {
	return Db.Where("token = ?", token).Delete(&UserToken{}).Error
}

// DeleteByUserId 删除指定用户的所有 Token (修改密码、停用或删除账号)
func (m *UserToken) DeleteByUserId(userId uint) error {
	return Db.Where("user_id = ?", userId).Delete(&UserToken{}).Error
}

// GetAllActiveUserTokens 获取所有有效用户的 Token 列表 (用于服务启动时预热内存)
func (m *UserToken) GetAllActiveUserTokens() ([]UserToken, map[uint]User, error) {
	var tokens []UserToken
	if err := Db.Find(&tokens).Error; err != nil {
		return nil, nil, err
	}

	userIdMap := make(map[uint]bool)
	for _, t := range tokens {
		userIdMap[t.UserId] = true
	}

	var userIds []uint
	for uid := range userIdMap {
		userIds = append(userIds, uid)
	}

	userMap := make(map[uint]User)
	if len(userIds) > 0 {
		var users []User
		if err := Db.Where("id IN ? AND status = 1", userIds).Find(&users).Error; err == nil {
			for _, u := range users {
				userMap[u.ID] = u
			}
		}
	}

	return tokens, userMap, nil
}
