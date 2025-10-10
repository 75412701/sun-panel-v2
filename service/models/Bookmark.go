package models

type Bookmark struct {
	BaseModel
	IconJson string `gorm:"type:varchar(1000)" json:"-"`
	Title    string `gorm:"type:varchar(50)" json:"title"`
	Url      string `gorm:"type:varchar(1000)" json:"url"`
	LanUrl   string `gorm:"type:varchar(1000)" json:"lanUrl"`
	Sort     int    `gorm:"type:int(11)" json:"sort"`
	id       int    `json:"itemIconGroupId"`
	UserId   uint   `json:"userId"`
	User     User   `json:"user"`
}
