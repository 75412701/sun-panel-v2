package database

import (
	"log"
	"os"
	"path"
	"sun-panel/lib/cmn"
	"sun-panel/models"
	"time"

	"gorm.io/driver/mysql"
	_ "gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	_ "gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

const (
	MYSQL  = "mysql"
	SQLITE = "sqlite"
)

type DbClient interface {
	Connect() (db *gorm.DB, err error)
}

type MySQLConfig struct {
	Username    string
	Password    string
	Host        string
	Port        string
	Database    string
	WaitTimeout int
}

type SQLiteConfig struct {
	Filename string
}

func DbInit(dbClient DbClient) (db *gorm.DB, dbErr error) {
	db, dbErr = dbClient.Connect()
	if dbErr != nil {
		return
	}
	return
}

// Connect mysql连接
func (d *MySQLConfig) Connect() (db *gorm.DB, err error) {
	dsn := d.Username + ":" + d.Password + "@tcp(" + d.Host + ":" + d.Port + ")/" + d.Database + "?charset=utf8mb4&parseTime=True&loc=Local"
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: GetLogger(),
		NamingStrategy: schema.NamingStrategy{
			// TablePrefix:   "blog_",
			SingularTable: true,
		},
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	sqlDb, _ := db.DB()
	sqlDb.SetMaxIdleConns(10)  // SetMaxIdleConns 设置空闲连接池中连接的最大数量
	sqlDb.SetMaxOpenConns(100) // SetMaxOpenConns 设置打开数据库连接的最大数量。
	wait_timeout := d.WaitTimeout
	sqlDb.SetConnMaxLifetime(time.Duration(wait_timeout * int(time.Second))) // SetConnMaxLifetime 设置了连接可复用的最大时间。
	return
}

// Connect sqllite3连接
func (d *SQLiteConfig) Connect() (db *gorm.DB, err error) {
	filePath := d.Filename
	exists := false
	if exists, err = cmn.PathExists(path.Dir(filePath)); err != nil {
		return
	} else {

		// 创建文件夹
		if !exists {
			if err = os.MkdirAll(path.Dir(filePath), 0700); err != nil {
				return
			}
		}

		db, err = gorm.Open(sqlite.Open(filePath), &gorm.Config{
			Logger: GetLogger(),
			NamingStrategy: schema.NamingStrategy{
				SingularTable: true,
			},
		})
	}

	return
}

// 日志
func GetLogger() logger.Interface {
	return logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer（日志输出的目标，前缀和日志包含的内容——译者注）
		logger.Config{
			SlowThreshold:             time.Second, // 慢 SQL 阈值
			LogLevel:                  logger.Warn, // 日志级别
			IgnoreRecordNotFoundError: true,        // 忽略ErrRecordNotFound（记录未找到）错误
			Colorful:                  true,        // 彩色打印
		},
	)

}

// 创建数据库
func CreateDatabase(driver string, db *gorm.DB) error {

	// mysql特殊处理
	if driver == MYSQL {
		db = db.Set("gorm:table_options", "ENGINE=InnoDB")
	}

	// 创建数据表
	err := db.AutoMigrate(
		&models.User{},
		&models.UserToken{},
		&models.SystemSetting{},
		&models.ItemIcon{},
		&models.UserConfig{},
		&models.File{},
		&models.ItemIconGroup{},
		&models.ModuleConfig{},
		&models.Bookmark{},
		&models.Notepad{},
		&models.SearchEngine{},
	)

	return err
}

// 初始化一个用户,一个用户都没有的时候创建一个
func NotFoundAndCreateUser(db *gorm.DB) error {
	fUser := models.User{}
	if err := db.First(&fUser).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return err
		}
		username := "admin"
		fUser.Mail = username
		fUser.Username = username
		fUser.Name = username
		fUser.Status = 1
		fUser.Role = 1
		fUser.Password = cmn.PasswordEncryption("123456")

		if errCreate := db.Create(&fUser).Error; errCreate != nil {
			return errCreate
		}
	}

	return nil
}

// 初始化搜索引擎数据
func NotFoundAndCreateSearchEngines(db *gorm.DB) error {
	var count int64
	// 检查是否已经有数据（包括软删除的）
	// 注意：这里我们使用Unscoped来检查，如果用户删除了所有数据，也不应该重复初始化
	// 但是用户需求是：默认加上，后续允许删除。如果用户全删了，软删除还在，Unscoped能查到，就不会再加了。
	// 如果用户硬删除了（物理删除），则会重新添加。
	if err := db.Unscoped().Model(&models.SearchEngine{}).Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	// 默认搜索引擎数据
	// 图标数据太长，这里简化处理，或者需要从前端获取？
	// 这里使用 base.go 中定义的空字符串或简单描述，具体图标由前端负责渲染默认值，或者后端存入长SVG字符串
	// 不过既然前端代码里有 createDefaultEngines，后端其实不需要存图标？
	// 不对，前端代码 createDefaultEngines 是调用后端接口添加的。
	// 如果后端预埋数据，需要提供完整的 IconSrc。
	// 为了简化，这里先存入基本的URL和Title，IconSrc 留空或存简单值，前端如果发现IconSrc为空可以显示默认图标？
	// 前端代码逻辑： <NAvatar :src="state.currentSearchEngine?.iconSrc || defaultSearchEngineList[0]?.iconSrc" ... />
	// 前端 createDefaultEngines 是把 SVG 内容传给后端的。
	// 这里我们直接尽量去读取项目里的 svg 文件内容？太复杂。
	// 我们只插入基础数据，图标留空，让用户自己编辑更好？或者只插入Google/Baidu/Bing的标题和URL
	// 实际上前端有 createDefaultEngines 逻辑，如果后端返回空列表，前端会初始化。
	// 但用户特别要求"修改这个表的逻辑,默认加上"。
	// 意味着后端初始化时就要有。

	// 定义默认数据
	defaultEngines := []models.SearchEngine{
		{
			Title:   "Google",
			Url:     "https://www.google.com/search?q=%s",
			Sort:    1,
			UserId:  1,                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               // 默认给管理员
			IconSrc: "data:image/x-icon;base64,iVBORw0KGgoAAAANSUhEUgAAACAAAAAgCAYAAABzenr0AAAACXBIWXMAAAsSAAALEgHS3X78AAAE3ElEQVRYhc2XW2xUVRSGv30uU9rTMhUaCuVixQSMpLYmPIhYwAcTIyKFROI1KQmkxBgj8UEf5MEYiSYmPvhAfCASY02MUaFEHjSRQYtGNDAqIhKxQJVeodPLMJ2Zc/by4cz91tZ4+5Od2Wf2Xuv/z9lrr722EhFmiTagA9gEbCwz5wQQAg4D4Vl5FZGZWqeIhGXuCKdsK/qvNNgsIqG/QFyIUMrXnAR0iEjkbyBPI5LyWcSlpDgGOoG3S67XV71wshd+CMPwEAiAAShoXIy0tsL6dtS6u8qt+E7gUO4fhQI6gI+LzM6ehjf2w9AQJC1QgCi/KRMwQZmIYSLKRBYvwXj2KVTLmlIituEHaZGAZvzIDeZN/+BVOPwhTAX850QgS44CZWWaoBDTRgzbd77lfoxdTxYKGMffUZfA/35pHCoif+8F+Pxdvx/wSr1NPpSR/3z2HERvFM4KkrMMaYtOCvf2p8/A5W5ouAHz4+AkoDoJpguGB6QEpT6GAKJUVsstKzBe2QdOTSmpG1OcmSUIA62Z4eEeCD0BV+pgnoahOhhwwBCwVsL23bCuHZxaf340Cl9/g+5+H0YjcOvKSuRpfA+0KRFpA87kDX2xCgYHIWpDNOALmLLhju2w4/kscSGiUeTIMdTWzTORp3GnhR/5WYwdhUA/NBhQKzDtQZUHddvhkZcru3Mc1GMPz4Y4jQ4LP7dnETkK8wywACVgu2DUwQOvzcXxbLHJojD4pnuz/WoBF1j2KATyN0ga310p41pAkc0xCqEpqFhcr3JnbbSK7OJ9oAyUpcAEAsDSLWVfYXd3OQGFGVaxpx262vP/Ldi4KelCZk+LoWDeirICykKpbEs/l0CRABGd7aeMhNLGcxJTxk/xFxAN4uJpl6R4TIuLG79U0b+gEbxU06TSUqpVRlEMEGhGJy4zJRpLmVzXHtWRIzTUbSjpoOueNInK++35EQbGZ+THwi+jMjtB17bjXesDDKLaT7f9I+9Q37QPy6wvcrCnXeWQ+5ichu5v08IkM7725iLzEwZ+DZeB1G9lVAv9nseg1vziac7EJ/nk1z0zv04KB89EGXVjRO0pkkYCVyVJGgnWFsdyyCDnbAaw6zsYtldwUQtXtTCshUEx+HL0Mw5e2MuUO1mR/MBPf7A//DvjNcNMVI8QC0zhGR4PtZQM5MMlD6OJiRDHz91LRBTjYtAnDtd1DWPikAi0squ5i/sa1hO0nIynnoGLvHn+Ar0DMQKx5VhuNbZbg5MIYno2x3fXsTSYJyJzGEGJMizUt5fjVw8wLjaXJciEOEyKw5DXSJwASapI6lpqzflEkh6SDEJ8CYiJii3HdB0CroMdX0BXy0Je2lAUPzuBQ7kVUYiCtNx9oYuPhnoYkIV4YjGmb2JSGnBReFKFxsbVVSgE8Rz09BJQGuJNWLEmPMOlrXYZJ7fdXkh+gtQZlJsHOvHLpQweX/UWmxp3ABCXQHZAzEzXUG6mr6yJTN+1x3lw+SKObV5dSD6e4vJtZlOUnoqc5rmfX+e32AQxCSIYCCag0NryT01dhXjVSHIBtbKUF29bx9OrVxW6ggpFaRqdlCnLj42c4uhImN6x8/RPXwMUIgYKYb5Zz93BNrY0ruHBRS0E7apSLorK8v/8YvK/vZr9a5fTUjFQDv/I9fxPxUx0d1WRkbMAAAAASUVORK5CYII=", // 暂时留空，前端可能有默认图标逻辑，或者用户自行上传
		},
		{
			Title:   "Baidu",
			Url:     "https://www.baidu.com/s?wd=%s",
			Sort:    2,
			UserId:  1,
			IconSrc: "data:image/png;base64,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",
		},
		{
			Title:   "Bing",
			Url:     "https://www.bing.com/search?q=%s",
			Sort:    3,
			UserId:  1,
			IconSrc: "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAACAAAAAgCAYAAABzenr0AAAAAXNSR0IArs4c6QAAAc1JREFUWEftld9ZwjAUxc+NC8QNZAJ1AwoD6AbqABacAJjAUhfADfRdKBvYDWQDswC9fmla/knKrcjHC3ltcvPLueeeEo686Mj34wRwUkCsgO4mTWTzW0DdALgozJuCOAWpgYmC2V8MvRNAdxMNRg/M3coLiKMCxNQBqQTIL8+yKUCXwqIpFAUmCsQQ1QCd5Hnt5YwpzmgAILWXuLbwPYC7BSBxZIbtJyGwPwd0N7lAxl/LQjwwcbu/rbAOx32AeotvToWpBMKrgA4/RoByLyMamWHwUFVQP04SEJpuDw9N3K72TFGsAmDyCeAq3yd4UdGOpKg7M3GrsacCEy4LmLglm5aMv+ucycX1UepwCQBF5xJnr56RQO8CsAZ0gVO/BamJW9d7tmAcAdRxnsLUvLSCShPWNG1Zy98CN+OlqWy3+iYObAb8WjpMegAvR1RRQxrN1UEUrqiQN4xHIPVqZzxPSTslc+4tx0+m1uoLBFGcq+DGUbz8am2WEI5XZpPO+cG3bEwvgiiXy9sysQKrG4totn22P6ZSkRnA71DqLW/LZiQLIHYqIFa92FgX4t8BLEcdiIMAbIPwJePBANYh/L/ygwJI/HMCOClwdAV+ACbgyyFwwhgnAAAAAElFTkSuQmCC",
		},
	}

	// 批量插入
	if err := db.Create(&defaultEngines).Error; err != nil {
		return err
	}

	return nil
}
