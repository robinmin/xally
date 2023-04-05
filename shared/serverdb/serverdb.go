package serverdb

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/robinmin/xally/config"
	"github.com/robinmin/xally/shared/model"
)

const MaxTokenLifeSpan = 24 // in hours
const MaxExtendTimes = 31

var _db *gorm.DB

func InitServerDB(connection_str string, verbose bool) (*gorm.DB, error) {
	var err error
	var db_cfg *gorm.Config

	if verbose {
		if config.SvrConfig.DebugMode() {
			fmt.Println("Opening database connection : ", connection_str)
			db_cfg = &gorm.Config{Logger: logger.Default.LogMode(logger.Info)}
		} else {
			db_cfg = &gorm.Config{}
		}
	} else {
		db_cfg = &gorm.Config{}
	}

	_db, err = gorm.Open(mysql.New(mysql.Config{
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
		// 设置字符集为utf8mb4
		_db = _db.Set("gorm:table_options", "CHARSET=utf8mb4")

		if config.SvrConfig.DebugMode() {
			if err = _db.AutoMigrate(
				&AuthUser{},
				&UserToken{},
				&ProxyLog{},
			); err != nil {
				log.Error(err)
			}
		}
	}

	return _db, err
}

func GetDB() *gorm.DB {
	return _db.Session(&gorm.Session{NewDB: true})
}

// func (user *User) SaveUser() (int64, error) {
// 	tx := DB.Create(&user)
// 	if tx.Error != nil {
// 		return 0, tx.Error
// 	}
// 	return tx.RowsAffected, nil
// }

// func extractUserInfo(access_token string) *User {
// 	access_info, _ := utility.ExtractAccessInfo(config.SvrConfig.Server.AppToken, access_token)
// 	if access_info == nil {
// 		log.Error("Faield to extract access information from access token : " + access_token)
// 		return nil
// 	}

// 	user := &User{
// 		// Token:      access_token, // TODO: need to add replace logic here
// 		UserID:     access_info["uid"].(string),
// 		Username:   access_info["username"].(string),
// 		Email:      access_info["email"].(string),
// 		DeviceInfo: access_info["device_info"].(string),
// 		Status:     0, // by default, waiting for activate
// 		RegisterAt: time.Now(),
// 		ExpiredAt:  time.Now(), // by default, expieried immediately. waiting for activate
// 	}
// 	return user
// }

// func GetUserByToken(access_token string, include_expired bool) (*User, error) {
// 	user := extractUserInfo(access_token)
// 	if user == nil {
// 		msg := "Faield to extract user info from access from access token : " + access_token
// 		log.Error(msg)
// 		return nil, errors.New(msg)
// 	}

// 	var tmp_user User
// 	tx := DB.Where(&User{
// 		// Token:      user.Token,
// 		UserID:     user.UserID,
// 		Username:   user.Username,
// 		Email:      user.Email,
// 		DeviceInfo: user.DeviceInfo,
// 		Status:     1,
// 	}).First(&tmp_user)
// 	if tx.Error != nil {
// 		log.Error("Failed to query user by access token : " + access_token)
// 		return nil, tx.Error
// 	}

// 	if !include_expired && tmp_user.ExpiredAt.Before(time.Now()) {
// 		return nil, errors.New("Token expired")
// 	}

// 	return &tmp_user, nil
// }

// func ActivateUser(access_token string) error {
// 	tx := DB.Model(&User{}).Where("token = ?", access_token).Update("status", 1)
// 	if tx.Error != nil {
// 		log.Error("Failed to activate user")
// 		return tx.Error
// 	}
// 	return nil
// }

// func DeactivateUser(access_token string) error {
// 	tx := DB.Model(&User{}).Where("token = ?", access_token).Update("status", 0)
// 	if tx.Error != nil {
// 		log.Error("Failed to activate user")
// 		return tx.Error
// 	}
// 	return nil
// }

// /////////////////////////////////////////////////////////////////////////////
type WhiteList struct {
	// AvailableUserMap map[string]time.Time
	AvailableUserMap map[string]WhiteListUser
	Mutex            *sync.RWMutex
}

func (w *WhiteList) LoadWhiteList(interval int64) {
	// 第一次立即更新白名单
	w.updateAll()

	// 定时更新白名单
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	for {
		select {
		case <-ticker.C:
			w.updateAll()
		}
	}
}

func (w *WhiteList) updateAll() error {
	// 从数据库获取最新数据
	var valid_users []WhiteListUser
	err := GetDB().Model(&UserToken{}).Select(
		"user_tokens.user_id, user_tokens.token, user_tokens.expired_at",
	).Joins(
		"inner join auth_users on auth_users.ID = user_tokens.user_id",
	).Where(
		"user_tokens.token_type=? and auth_users.is_actived=1 and auth_users.is_verified=1 and now() between user_tokens.created_at and user_tokens.expired_at and now() between auth_users.created_at and auth_users.expired_at", "access",
	).Find(&valid_users).Error
	if err != nil {
		log.Error("Failed to get all auth_users on white list ")
		return err
	}

	// update the user list
	user_list := make(map[string]WhiteListUser)
	for _, tmp_user := range valid_users {
		user_list[tmp_user.Token] = tmp_user
	}

	// 加写锁，更新白名单
	w.Mutex.Lock()
	w.AvailableUserMap = user_list
	w.Mutex.Unlock()

	return nil
}

func (w *WhiteList) IsAccessTokenValid(access_token string) bool {
	if user, ok := w.AvailableUserMap[access_token]; !ok || user.ExpiredAt.Local().Before(time.Now()) {
		log.Error("Token is invalid or already expired!")
		return false
	}
	return true
}

func (w *WhiteList) GetUserInfoByToken(access_token string) *WhiteListUser {
	user, ok := w.AvailableUserMap[access_token]
	if !ok {
		log.Error("Token is invalid or already expired!")
		return nil
	}
	return &user
}

func (w *WhiteList) RefreshToken(old_access_token string) (string, error) {
	new_access_token, err := RefreshAccessToken(old_access_token)
	if err != nil {
		log.Error("Failed to get user id by access token " + old_access_token + " : " + err.Error())
		return "", err
	}

	// 从数据库获取最新数据
	var valid_user WhiteListUser
	err = GetDB().Model(&UserToken{}).Select(
		"user_tokens.user_id, user_tokens.token, user_tokens.expired_at",
	).Joins(
		"inner join auth_users on auth_users.id = user_tokens.user_id and auth_users.id = ? ", new_access_token.UserID,
	).Where(
		"user_tokens.token_type=? and auth_users.is_actived=1 and auth_users.is_verified=1 and now() between user_tokens.created_at and user_tokens.expired_at and now() between auth_users.created_at and auth_users.expired_at", "access",
	).First(&valid_user).Error
	if err != nil {
		log.Error("Failed to get current auth_user on white list ")
		return "", err
	}

	if new_access_token.Token != old_access_token {
		w.Mutex.Lock()
		delete(w.AvailableUserMap, old_access_token)
		w.AvailableUserMap[new_access_token.Token] = valid_user
		w.Mutex.Unlock()
	} else {
		w.Mutex.Lock()
		w.AvailableUserMap[new_access_token.Token] = valid_user
		w.Mutex.Unlock()
	}
	return new_access_token.Token, nil
}

func newUserToken(token_type string, user_id uint) (*UserToken, error) {
	token := &UserToken{
		TokenType:      token_type,
		Token:          uuid.New().String(),
		ExpiredAt:      time.Now().Add(MaxExtendTimes * MaxTokenLifeSpan * time.Hour),
		UserID:         user_id,
		ConsumeCounter: 0,
	}
	tx := GetDB().Create(&token)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return token, nil
}

func NewActiviationToken(user_id uint) (*UserToken, error) {
	return newUserToken("activation", user_id)
}

func NewAccessToken(user_id uint) (*UserToken, error) {
	return newUserToken("access", user_id)
}

func GetUserIDByActivationToken(token string) (uint, error) {
	var tmp_token UserToken
	tx := GetDB().Model(&UserToken{}).Where("token_type=? and token=? and consume_counter=0 and now() between created_at and expired_at", "activation", token).First(&tmp_token)
	if tx.Error != nil {
		log.Error("Invalid token or token already expired : " + token)
		return 0, tx.Error
	}

	return tmp_token.UserID, nil
}

func GetUserIDByAccessToken(token string) (uint, error) {
	var tmp_token UserToken
	tx := GetDB().Model(&UserToken{}).Where("token_type=? and token=? and now() between created_at and expired_at", "access", token).First(&tmp_token)
	if tx.Error != nil {
		tx.Rollback()
		log.Error("Invalid token or token already expired : " + token)
		return 0, tx.Error
	}

	return tmp_token.UserID, nil
}

func RefreshAccessToken(token string) (*UserToken, error) {
	var tmp_token UserToken
	tx := GetDB().Model(&UserToken{}).Where("token_type=? and token=? and now() between created_at and expired_at", "access", token).First(&tmp_token)
	if tx.Error != nil {
		tx.Rollback()
		log.Error("Invalid token or token already expired : " + token)
		return nil, tx.Error
	}

	if tmp_token.CreatedAt.Add(MaxExtendTimes * MaxTokenLifeSpan * time.Hour).Before(time.Now()) {
		// already reach the max lifetim of current token, force to expired
		tx = GetDB().Model(&UserToken{}).Where("token = ?", token).Update("expired_at", time.Now())
		if tx.Error != nil {
			log.Error("Failed to expire current access token")
			return nil, tx.Error
		}
		new_token, err := NewAccessToken(tmp_token.UserID)
		if tx.Error != nil {
			log.Error("Failed to generate new token")
			return nil, err
		}
		return new_token, nil
	} else {
		// extend the expiry date
		tmp_token.ExpiredAt = time.Now().Add(MaxExtendTimes * MaxTokenLifeSpan * time.Hour)
		tx = GetDB().Model(&UserToken{}).Where("token = ?", token).Update("expired_at", tmp_token.ExpiredAt)
		if tx.Error != nil {
			log.Error("Failed to extend the expiry date")
			return nil, tx.Error
		}
		return &tmp_token, nil
	}
}

// /////////////////////////////////////////////////////////////////////////////
// /////////////////////////////////////////////////////////////////////////////
// /////////////////////////////////////////////////////////////////////////////

func RegisterUser(user_info *model.UserInfo) (*AuthUser, error) {
	if user_info.Email == "" {
		return nil, errors.New("Email cannot be empty")
	}

	old_user, _ := getUserByEmail(user_info.Email)
	// if err != nil {
	// 	return nil, err
	// }

	// prepare user registery information
	user := &AuthUser{
		Username:   user_info.Username,
		Hostname:   user_info.Hostname,
		Email:      user_info.Email,
		DeviceInfo: user_info.DeviceInfo,
		Password:   user_info.Password,
		IsActived:  0,
		IsVerified: 0,

		RegisterAt: time.Now(),
		ExpiredAt:  time.Now(), // by default, expieried immediately. waiting for activate
	}

	if old_user != nil && old_user.ID != 0 {
		// update user information
		tx := GetDB().Model(&AuthUser{}).Where("id = ?", old_user.ID).Updates(user)
		if tx.Error != nil {
			log.Errorf("Failed to re-register existing user id : %v\n", old_user.ID)
			return nil, tx.Error
		}

		// then retrive user information again
		var err error
		old_user, err = getUserByEmail(user_info.Email)
		if err != nil {
			log.Errorf("Failed to retrive user information again : %v\n", old_user.ID)
			return nil, err
		}
		return old_user, nil
	}

	tx := GetDB().Model(&AuthUser{}).Create(user)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected != 1 {
		return nil, errors.New("Failed to create user : " + user_info.Email)
	}

	return getUserByEmail(user_info.Email)
}

func getUserByID(user_id uint) (*AuthUser, error) {
	var tmp_user AuthUser
	tx := GetDB().Model(&AuthUser{}).Where("id = ?", user_id).First(&tmp_user)
	if tx.Error != nil {
		log.Errorf("Failed to query user by user id : %d\n", user_id)
		return nil, tx.Error
	}

	return &tmp_user, nil
}

func getUserByEmail(email string) (*AuthUser, error) {
	var tmp_user AuthUser
	tx := GetDB().Model(&AuthUser{}).Where("email = ?", email).First(&tmp_user)
	if tx.Error != nil {
		log.Errorf("Failed to query user by email: %d\n", email)
		return nil, tx.Error
	}

	return &tmp_user, nil
}

func GetValidUser(user_id uint) (*AuthUser, error) {
	user, err := getUserByID(user_id)
	if err != nil {
		return nil, err
	}
	t := time.Now()
	if user.IsActived == 1 && user.IsVerified == 1 && user.ActivateAt.After(t) && user.ExpiredAt.Before(t) {
		return nil, errors.New("Invalid user id")
	}
	return user, nil
}

func ActiviateUser(user_id uint) (int64, error) {
	updated_user := AuthUser{
		IsActived:  1,
		IsVerified: 1,
		ActivateAt: time.Now(),
		ExpiredAt:  time.Now().Add(MaxExtendTimes * MaxTokenLifeSpan * time.Hour),
	}
	updated_user.UpdatedAt = time.Now()

	tx := GetDB().Model(&AuthUser{}).Where("id = ?", user_id).Updates(updated_user)
	if tx.Error != nil {
		log.Errorf("Failed to activate user id : %v\n", user_id)
		return 0, tx.Error
	}
	return tx.RowsAffected, nil
}

func DeactivateUser(user_id uint) (int64, error) {
	updated_user := AuthUser{
		IsActived:    0,
		DeactivateAt: time.Now(),
		ExpiredAt:    time.Now(),
	}
	updated_user.UpdatedAt = time.Now()

	tx := GetDB().Model(&AuthUser{}).Where("id = ?", user_id).Updates(updated_user)
	if tx.Error != nil {
		log.Errorf("Failed to deactivate user id : %v\n", user_id)
		return 0, tx.Error
	}
	return tx.RowsAffected, nil
}

func (user *AuthUser) VerifyUser() (*AuthUser, error) {
	var tmp_user AuthUser
	// load user information from database by keys
	tx := GetDB().Model(&AuthUser{}).Where(
		"username = ? and email=? device_info=? and is_actived=1 and is_verified=1 and now() between activate_at and expired_at",
		user.Username,
		user.Email,
		user.DeviceInfo,
	).First(&tmp_user)
	if tx.Error != nil {
		log.Error("Failed to load user information from database by keys")
		return nil, tx.Error
	}

	return &tmp_user, nil
}

func (user *AuthUser) Logout() error {
	// TODO: update white list
	return nil
}

type ProxyLog struct {
	ID                 uint      `gorm:"primary_key"`
	UserID             uint      `gorm:"not null"`                           // user id
	RemoteAddr         string    `gorm:"type:varchar(255)"`                  // remote ip address
	RequestTime        time.Time `gorm:"not null"`                           // 请求时间
	RequestMethod      string    `gorm:"not null"`                           // 请求方法
	RequestURL         string    `gorm:"not null"`                           // 请求URL
	RequestHeaders     string    `gorm:"type:text"`                          // 请求头
	RequestBody        string    `gorm:"type:longtext"`                      // 请求体
	ResponseStatusCode int       `gorm:"not null"`                           // 响应状态码
	ResponseHeaders    string    `gorm:"type:text"`                          // 响应头
	ResponseBody       string    `gorm:"type:longtext"`                      // 响应体
	CreatedAt          time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"` // 创建时间
}

func (plog *ProxyLog) RecordRequest() error {
	tx := GetDB().Model(&ProxyLog{}).Create(plog)
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}
