package serverdb

import (
	"errors"
	"fmt"
	"time"

	"github.com/robinmin/xally/config"
	"github.com/robinmin/xally/shared/utility"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitServerDB(connection_str string, verbose bool) (*gorm.DB, error) {
	var err error
	var db_cfg *gorm.Config

	if verbose {
		fmt.Println("Opening database connection : ", connection_str)
		db_cfg = &gorm.Config{Logger: logger.Default.LogMode(logger.Info)}
	} else {
		db_cfg = &gorm.Config{}
	}

	DB, err = gorm.Open(mysql.New(mysql.Config{
		DSN:                       connection_str, // DSN data source name
		DefaultStringSize:         256,            // string 类型字段的默认长度
		DisableDatetimePrecision:  true,           // 禁用 datetime 精度，MySQL 5.6 之前的数据库不支持
		DontSupportRenameIndex:    true,           // 重命名索引时采用删除并新建的方式，MySQL 5.7 之前的数据库和 MariaDB 不支持重命名索引
		DontSupportRenameColumn:   true,           // 用 `change` 重命名列，MySQL 8 之前的数据库和 MariaDB 不支持重命名列
		SkipInitializeWithVersion: false,          // 根据当前 MySQL 版本自动配置
	}), db_cfg)
	if err != nil {
		log.Error("Failed to connect to database: ", err.Error())
	} else {
		if err = DB.AutoMigrate(&User{}); err != nil {
			log.Error(err)
		}
		// log.Debug("TODO: migrate")
	}

	return DB, err
}

func (user *User) SaveUser() (int64, error) {
	tx := DB.Create(&user)
	if tx.Error != nil {
		return 0, tx.Error
	}
	return tx.RowsAffected, nil
}

func extractUserInfo(access_token string) *User {
	access_info, _ := utility.ExtractAccessInfo(config.SvrConfig.Server.APITokenSecret, access_token)
	if access_info == nil {
		log.Error("Faield to extract access information from access token : " + access_token)
		return nil
	}

	user := &User{
		Token:      access_token,
		UserID:     access_info["uid"].(string),
		Username:   access_info["username"].(string),
		Email:      access_info["email"].(string),
		DeviceInfo: access_info["device_info"].(string),
		Status:     0, // by default, waiting for activiate
		RegisterAt: time.Now(),
		ExpiredAt:  time.Now(), // by default, expieried immediately. waiting for activiate
	}
	return user
}
func RegisterUser(access_token string) (*User, error) {
	user := extractUserInfo(access_token)
	if user == nil {
		msg := "Faield to extract user info from access from access token : " + access_token
		log.Error(msg)
		return nil, errors.New(msg)
	}

	rows, err := user.SaveUser()
	if err != nil {
		msg := "Faield to add new user by token : " + access_token
		log.Error(msg)
		return nil, err
	}
	log.Printf("# of User has been added : %v", rows)

	return user, nil
}

func GetUserByToken(access_token string) (*User, error) {
	user := extractUserInfo(access_token)
	if user == nil {
		msg := "Faield to extract user info from access from access token : " + access_token
		log.Error(msg)
		return nil, errors.New(msg)
	}

	var tmp_user User
	tx := DB.Where(&User{
		// Token:      user.Token,
		UserID:     user.UserID,
		Username:   user.Username,
		Email:      user.Email,
		DeviceInfo: user.DeviceInfo,
		Status:     1,
	}).First(&tmp_user)
	if tx.Error != nil {
		log.Error("Failed to query user by access token : " + access_token)
		return nil, tx.Error
	}

	if tmp_user.ExpiredAt.Before(time.Now()) {
		return nil, errors.New("Token has expired")
	}
	return &tmp_user, nil
}

func GetAllUsers() (map[string]time.Time, error) {
	var valid_users []User
	tx := DB.Where("status = 1 and expired_at > ?", time.Now()).Find(&valid_users)
	if tx.Error != nil {
		log.Error("Failed to get all token list ")
		return nil, tx.Error
	}

	// update the user list
	user_list := make(map[string]time.Time)
	for _, tmp_user := range valid_users {
		user_list[tmp_user.Token] = tmp_user.ExpiredAt
	}

	return user_list, nil
}

func ActiviateUser(access_token string) error {
	tx := DB.Model(&User{}).Where("token = ?", access_token).Update("status", 1)
	if tx.Error != nil {
		log.Error("Failed to activiate user")
		return tx.Error
	}
	return nil
}

func DeactiviateUser(access_token string) error {
	tx := DB.Model(&User{}).Where("token = ?", access_token).Update("status", 0)
	if tx.Error != nil {
		log.Error("Failed to activiate user")
		return tx.Error
	}
	return nil
}
