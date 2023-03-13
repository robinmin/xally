package serverdb

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// var Token *token.Token

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
		log.Debug("TODO: migrate")
	}

	return DB, err
}

func (user *User) SaveUser() (int64, error) {
	var err error

	tx := DB.Create(&user)
	if tx.Error != nil {
		return 0, err
	}
	return tx.RowsAffected, nil
}

func (user *User) BeforeSave() error {
	//turn password into hash
	// hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	// if err != nil {
	// 	return err
	// }
	// user.Password = string(hashedPassword)

	// //remove spaces in username
	// user.Username = html.EscapeString(strings.TrimSpace(user.Username))
	return nil
}

func VerifyPassword(password, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func LoginCheck(username string, password string) (string, error) {
	// var err error

	// u := User{}
	// err = DB.Model(User{}).Where("username = ?", username).Take(&u).Error
	// if err != nil {
	// 	return "", err
	// }

	// err = VerifyPassword(password, u.Password)
	// if err != nil && err == bcrypt.ErrMismatchedHashAndPassword {
	// 	return "", err
	// }

	// token, err := Token.GenerateToken(u.ID)
	// if err != nil {
	// 	return "", err
	// }

	// return token, nil
	return "", nil
}

// import (
//     "database/sql"
//     "errors"
//     "time"

//     "github.com/google/uuid"
// )

// type Token struct {
//     ID        string    `json:"id"`
//     UserID    string    `json:"user_id"`
//     Token     string    `json:"-"`
//     CreatedAt time.Time `json:"created_at"`
//     ExpiresAt time.Time `json:"expires_at"`
// }

// func (t *Token) Create(db *sql.DB) error {
//     t.ID = uuid.New().String()
//     t.Token = uuid.New().String()

//     _, err := db.Exec("INSERT INTO tokens (id, user_id, token, created_at,
// expires_at) VALUES ($1, $2, $3, $4, $5)", t.ID, t.UserID, t.Token,
// time.Now(), time.Now().Add(time.Hour*24))
//     if err != nil {
//         return err
//     }

//     return nil
// }

// func GetToken(db *sql.DB, token string) (*Token, error) {
//     var t Token

//     err := db.QueryRow("SELECT id, user_id, token, created_at, expires_at
// FROM tokens WHERE token = $1", token).Scan(&t.ID, &t.UserID, &t.Token,
// &t.CreatedAt, &t.ExpiresAt)
//     if err != nil {
//         return nil, err
//     }

//     if t.ExpiresAt.Before(time.Now()) {
//         return nil, errors.New("Token has expired")
//     }

//     return &t, nil
// }

// func DeleteToken(db *sql.DB, token string) error {
//     _, err := db.Exec("DELETE FROM tokens WHERE token = $1", token)
//     if err != nil {
//         return err
//     }

//     return nil
// }

// import (
//     "database/sql"
//     "errors"
//     "time"

//     "golang.org/x/crypto/bcrypt"
// )

// type User struct {
//     ID        string    `json:"id"`
//     Username  string    `json:"username"`
//     Password  string    `json:"-"`
//     CreatedAt time.Time `json:"created_at"`
//     UpdatedAt time.Time `json:"updated_at"`
// }

// func (u *User) Create(db *sql.DB) error {
//     hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password),
// bcrypt.DefaultCost)
//     if err != nil {
//         return err
//     }

//     err = db.QueryRow("INSERT INTO users (username, password) VALUES ($1, $2)
// RETURNING id, created_at, updated_at", u.Username,
// hashedPassword).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
//     if err != nil {
//         return err
//     }

//     return nil
// }

// func GetUser(db *sql.DB, id string) (*User, error) {
//     var user User

//     err := db.QueryRow("SELECT id, username, password, created_at, updated_at
// FROM users WHERE id = $1", id).Scan(&user.ID, &user.Username,
// &user.Password, &user.CreatedAt, &user.UpdatedAt)
//     if err != nil {
//         return nil, err
//     }

//     return &user, nil
// }

// func (u *User) Update(db *sql.DB, id string) error {
//     if u.ID != id {
//         return errors.New("User ID in payload does not match ID in URL")
//     }

//     if u.Password != "" {
//         hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password),
// bcrypt.DefaultCost)
//         if err != nil {
//             return err
//         }
//         u.Password = string(hashedPassword)
//     }

//     _, err := db.Exec("UPDATE users SET username = $1, password = $2,
// updated_at = $3 WHERE id = $4", u.Username, u.Password, time.Now(), id)
//     if err != nil {
//         return err
//     }

//     return nil
// }

// func DeleteUser(db *sql.DB, id string) error {
//     _, err := db.Exec("DELETE FROM users WHERE id = $1", id)
//     if err != nil {
//         return err
//     }

//     return nil
// }
