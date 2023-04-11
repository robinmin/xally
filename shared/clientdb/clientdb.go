package clientdb

import (
	"fmt"
	"unicode/utf8"

	"github.com/robinmin/xally/config"
	"github.com/robinmin/xally/shared/model"
	log "github.com/sirupsen/logrus"

	// "gorm.io/driver/sqlite" // // Sqlite driver based on GGO
	"github.com/glebarez/sqlite" // Pure go SQLite driver, checkout https://github.com/glebarez/sqlite for details
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type ClientDB struct {
	db *gorm.DB
}

func InitClientDB(db_name string, verbose bool) (*ClientDB, error) {
	var err error
	var db_cfg *gorm.Config

	if verbose {
		fmt.Println("Opening database : ", db_name)
		if config.MyConfig.DebugMode() {
			db_cfg = &gorm.Config{Logger: logger.Default.LogMode(logger.Info)}
		} else {
			db_cfg = &gorm.Config{}
		}
	} else {
		db_cfg = &gorm.Config{}
	}

	cdb := &ClientDB{}
	cdb.db, err = gorm.Open(sqlite.Open(db_name), db_cfg)
	if err != nil {
		log.Error("Failed to open DB plugin: ", db_name, err)
	} else {
		// if config.MyConfig.DebugMode() {
		if err = cdb.db.AutoMigrate(
			&OptionHistory{},
			&model.ConversationHistory{},
		); err != nil {
			log.Error(err)
		}
		// }
	}
	return cdb, err
}

func (cdb *ClientDB) LoadOptionHistory(role_name string) ([]string, error) {
	option_history := []string{}
	records := &[]OptionHistory{}

	cdb.db.Unscoped().Where("role = ?", role_name).Find(records)
	for _, record := range *records {
		option_history = append(option_history, record.Option)
	}
	return option_history, nil
}

func (cdb *ClientDB) AddOptionHistory(op_history *OptionHistory) bool {
	if op_history != nil {
		op_history.Option = TruncateStr(op_history.Option, 256)
		tx := cdb.db.Create(op_history)
		if tx.Error != nil {
			log.Error("Failed to add new option history")
			log.Error(tx.Error)
		} else {
			return true
		}
	}
	return false
}

func (cdb *ClientDB) AddChatHistory(chat_history *model.ConversationHistory) bool {
	if chat_history != nil {
		tx := cdb.db.Create(chat_history)
		if tx.Error != nil {
			log.Error("Failed to add new option history")
			log.Error(tx.Error)
		} else {
			return true
		}
	}
	return false
}

func TruncateStr(str string, maxLen int) string {
	if utf8.RuneCountInString(str) > maxLen {
		// 如果字符串长度超过最大长度，则截取前maxLen个字符
		return string([]rune(str)[:maxLen])
	}
	return str
}
