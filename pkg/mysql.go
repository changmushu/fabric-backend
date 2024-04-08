package pkg

import (
	"backend/model"
	"database/sql"
	// "fmt"

	// "log"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

// 定义一个全局对象db
var db *sql.DB

func MysqlInit() (err error) {
	// 初始化数据库
	dsn := viper.GetString("mysql.user") + ":" + viper.GetString("mysql.password") + "@tcp(" + viper.GetString("mysql.host") + ":" + viper.GetString("mysql.port") + ")/"
	// 不会校验账号密码是否正确
	// 注意！！！这里不要使用:=，我们是给全局变量赋值，然后在main函数中使用全局变量db
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	db.SetConnMaxLifetime(time.Minute)
	// 创建数据库与表（如果不存在的话）
	db.Exec("CREATE DATABASE IF NOT EXISTS " + viper.GetString("mysql.db"))
	db.Exec("USE " + viper.GetString("mysql.db"))
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS users (user_id VARCHAR(50) PRIMARY KEY, username VARCHAR(50) UNIQUE NOT NULL, password VARCHAR(50) NOT NULL, realInfo VARCHAR(100), telephone VARCHAR(50), realname VARCHAR(50), card_id VARCHAR(50), Txid VARCHAR(255), isPass VARCHAR(255))")
	if err != nil {
		panic(err.Error())
	}
	// _, err = db.Exec("CREATE TABLE IF NOT EXISTS usersInfo (user_id VARCHAR(50) PRIMARY KEY, username VARCHAR(50) UNIQUE NOT NULL, password VARCHAR(50) NOT NULL, realInfo VARCHAR(100))")
	// if err != nil {
	// 	panic(err.Error())
	// }
	// 重新配置下数据库连接信息
	dsn = viper.GetString("mysql.user") + ":" + viper.GetString("mysql.password") + "@tcp(" + viper.GetString("mysql.host") + ":" + viper.GetString("mysql.port") + ")/" + viper.GetString("mysql.db")
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	// 尝试与数据库建立连接（校验dsn是否正确）
	err = db.Ping()
	if err != nil {
		panic(err.Error())
	}
	return nil
}

// 注册
func InsertUser(user *model.MysqlUser) (err error) {
	sqlStr := "select count(user_id) from users where username = ?"
	var count int64
	err = db.QueryRow(sqlStr, user.Username).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("用户名已存在")
	}
	sqlStr = "insert into users(user_id,username,password,realInfo) values(?,?,?,?)"
	_, err = db.Exec(sqlStr, user.UserID, user.Username, EncryptByMD5(user.Password), EncryptByMD5(user.RealInfo))
	if err != nil {
		return err
	}
	return nil
}

// 保存交易哈希
func UpdateTxID(user *model.MysqlUser) (err error) {
	sqlStr := "UPDATE users SET txid = ? WHERE user_id = ?"
	_, err = db.Exec(sqlStr, user.TxID)
	if err != nil {
		return err
	}
	return nil
}

// 更新用户
func UpdateUser(user *model.MysqlUser) (err error) {
	sqlStr := "UPDATE users SET telephone = ?, realname = ?, card_id = ?, isPass = ? WHERE user_id = ?"
	_, err = db.Exec(sqlStr, user.Telephone, user.Realname, user.CardID, user.IsPass, user.UserID)
	if err != nil {
		return err
	}
	return nil
}

// 登录
func Login(user *model.MysqlUser) (err error) {
	sqlStr := "select username,password from users where username = ?"
	var password string
	err = db.QueryRow(sqlStr, user.Username).Scan(&user.Username, &password)
	if err != nil {
		return err
	}
	if EncryptByMD5(user.Password) != password {
		return errors.New("密码错误")
	}
	return nil
}

// 获取用户ID
func GetUserID(username string) (userID string, err error) {
	sqlStr := "select user_id from users where username = ?"
	err = db.QueryRow(sqlStr, username).Scan(&userID)
	if err != nil {
		return "", err
	}
	return userID, nil
}

// 获取用户信息
func GetUsername(userID string) (username string, telephone string, realname string, card_id string, Txid string, isPass string, err error) {
	sqlStr := "SELECT username, telephone, realname, card_id, Txid, isPass FROM users WHERE user_id = ?"
	err = db.QueryRow(sqlStr, userID).Scan(&username, &telephone, &realname, &card_id, &Txid, &isPass)
	if err != nil {
		return "", "", "", "", "", "", err
	}
	return username, telephone, realname, card_id, Txid, isPass, nil
}

type User struct {
	UserName  string `json:"username"`
	UserID    string `json:"userid"`
	UserType  string `json:"usertype"`
	Telephone string `json:"telephone"`
	Realname  string `json:"realname"`
	CardID    string `json:"card_id"`
	TxID      string `json:"Txid"`
	IsPass    string `json:"isPass"`
}

// 获取全部用户信息
func GetAllUserNames() ([]User, error) {
	var users []User
	sqlStr := "SELECT username, user_id, telephone, realname, card_id, Txid, isPass FROM users"
	rows, err := db.Query(sqlStr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var user User
		if err := rows.Scan(&user.UserName, &user.UserID, &user.Telephone, &user.Realname, &user.CardID, &user.TxID, &user.IsPass); err != nil {
			return nil, err
		}
		// 在循环中调用链码查询
		userType, err := ChaincodeQuery("GetUserType", user.UserID)
		if err != nil {
			return nil, err
		}

		// 将查询到的用户类型添加到用户结构体中
		user.UserType = userType
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}
