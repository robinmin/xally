package clientdb

import (
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type ClientDB struct {
	db *gorm.DB
}

func InitClientDB(db_name string, verbose bool) (*ClientDB, error) {
	var err error
	var db_cfg *gorm.Config

	// if verbose {
	// 	fmt.Println("Opening database : ", db_name)
	// 	db_cfg = &gorm.Config{Logger: logger.Default.LogMode(logger.Info)}
	// } else {
	db_cfg = &gorm.Config{}
	// }

	cdb := &ClientDB{}
	cdb.db, err = gorm.Open(sqlite.Open(db_name), db_cfg)
	if err != nil {
		log.Error("Failed to open DB plugin: ", db_name, err)
	} else {
		if err = cdb.db.AutoMigrate(&OptionHistory{}); err != nil {
			log.Error(err)
		}
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
		if len(op_history.Option) > 256 {
			op_history.Option = op_history.Option[0:256]
		}
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
