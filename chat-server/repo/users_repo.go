package repo

import (
	"chat-ai/chat-server/config"
	_const "chat-ai/chat-server/const"
	mylog "chat-ai/chat-server/log"
	"chat-ai/chat-server/model"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"time"
)

var dsn string
var MyDB *gorm.DB

func getDSN() string {
	var (
		dbUser     = config.GlobalConf.Database.Username
		dbPassword = config.GlobalConf.Database.Password
		dbHost     = config.GlobalConf.Database.Host
		dbPort     = config.GlobalConf.Database.Port
		dbName     = config.GlobalConf.Database.Database
	)
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", dbUser, dbPassword, dbHost, dbPort, dbName)
}

func GetDB() error {
	dsn = getDSN()
	var err error
	MyDB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		mylog.Logger.Errorf("failed to connect to database: %v", err)
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	return nil
}

func QueryUserByEmailAndStatus(email string, user *model.User) *gorm.DB {
	return MyDB.Where("email = ? and status = 0", email).First(user)
}

func QueryUserByPhoneOrEmail(email string, user *model.User) *gorm.DB {
	result := MyDB.Where("email = ?", email).First(user)
	return result
}

func InsertUser(user *model.User) *gorm.DB {
	return MyDB.Create(user)
}

func QueryById(id int, user *model.User) *gorm.DB {
	return MyDB.Where("id = ? ", id).First(user)
}

func UpdateVIPExpireDate(userID uint64, rechargeTime int) error {
	// Start a transaction to ensure atomicity
	tx := MyDB.Begin()

	// Get the user from the database
	user := model.User{}
	if err := tx.Where("id = ? FOR UPDATE", userID).First(&user).Error; err != nil {
		tx.Rollback()
		return err
	}
	// Calculate the new VIP expire date
	now := time.Now().Unix()
	if user.VipExpireDate > now {
		user.VipExpireDate += int64(rechargeTime)
	} else {
		user.VipExpireDate = now + int64(rechargeTime)
	}

	// Update the user's VIP expire date
	if err := tx.Model(&user).
		Update("vip_expire_date", user.VipExpireDate).
		Update("vip_status", _const.USER_VIP_STATUS).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

// 分页查询所有用户信息
func GetAllUsers(pageNum int, pageSize int) ([]*model.User, int64, error) {
	var users []*model.User
	var total int64
	offset := (pageNum - 1) * pageSize
	err := MyDB.Offset(offset).Limit(pageSize).Find(&users).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, 0, err
	}
	// 查询总记录数
	MyDB.Model(&model.User{}).Count(&total)
	return users, total, nil
}
