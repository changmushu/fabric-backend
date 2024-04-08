package controller

import (
	"backend/model"
	"backend/pkg"
	"fmt"

	// "log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type user struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	Usertype  string `json:"userType"`
	Telephone string `json:"telephone"`
	Realname  string `json:"realname"`
	CardID    string `json:"cardID"`
	TxID      string `json:"txid"`
	IsPass    string `json:"isPass"`
}

type res struct {
	Username  string `json:"username"`
	Usertype  string `json:"userType"`
	Telephone string `json:"telephone"`
	Realname  string `json:"realname"`
	CardID    string `json:"cardID"`
	TxID      string `json:"txid"`
	IsPass    string `json:"isPass"`
}

func Register(c *gin.Context) {
	// 将用户信息存入mysql数据库
	var userJson user
	var user model.MysqlUser
	//将表单数据转为JSON数据
	if err := c.ShouldBindJSON(&userJson); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	//赋值
	user.UserID = pkg.GenerateID()
	user.Username = userJson.Username
	user.Password = userJson.Password
	user.RealInfo = pkg.EncryptByMD5(userJson.Username)
	user.Telephone = "--"
	user.Realname = "--"
	user.CardID = "--"
	err := pkg.InsertUser(&user)
	// log.Println(user)
	if err != nil {
		// log.Println("Error inserting user into MySQL database:", err)
		// 返回错误信息给客户端或者记录日志
		c.JSON(200, gin.H{
			"code":    500,
			"message": "注册失败!该用户已存在！",
		})
		return
	}
	// 将用户信息存入区块链
	// userID string, userType string, realInfoHash string
	// 将post请求的参数封装成一个数组args
	var args []string
	var TxIDUser model.MysqlUser
	args = append(args, user.UserID)
	args = append(args, userJson.Usertype)
	args = append(args, user.RealInfo)
	res, err := pkg.ChaincodeInvoke("RegisterUser", args)
	if err != nil {
		c.JSON(200, gin.H{
			"code":    500,
			"message": "注册失败:" + err.Error(),
		})
		return
	}
	TxIDUser.TxID = res
	err = pkg.UpdateTxID(&TxIDUser)
	c.JSON(200, gin.H{
		"code":    200,
		"message": "注册成功！",
		"txid":    res,
	})
}

func Login(c *gin.Context) {
	var userJson user
	var user model.MysqlUser

	if err := c.ShouldBindJSON(&userJson); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user.Username = userJson.Username
	user.Password = userJson.Password
	// 获取用户ID
	var err error
	user.UserID, err = pkg.GetUserID(user.Username)
	if err != nil {
		c.JSON(200, gin.H{
			"code":    200,
			"message": "没有找到该用户",
		})
		return
	}
	userType, err := GetUserType(user.UserID)
	if err != nil {
		c.JSON(200, gin.H{
			"code":    500,
			"message": "登陆失败:" + err.Error(),
		})
		return
	}
	err = pkg.Login(&user)
	if err != nil {
		c.JSON(200, gin.H{
			"code":    500,
			"message": "登陆失败:" + err.Error(),
		})
		return
	}

	// 生成jwt
	jwt, err := pkg.GenToken(user.UserID, userType)
	if err != nil {
		c.JSON(200, gin.H{
			"code":    500,
			"message": "登陆失败:" + err.Error(),
		})
		return
	}
	c.JSON(200, gin.H{
		"code":    200,
		"message": "登陆成功！",
		"token":   jwt,
	})
}

func Logout(c *gin.Context) {
	c.JSON(200, gin.H{
		"code":    200,
		"message": "登出成功！",
	})
}

// 获取用户类型
func GetUserType(userID string) (string, error) {
	userType, err := pkg.ChaincodeQuery("GetUserType", userID)
	if err != nil {
		return "", err
	}
	return userType, nil
}

// 获取所有用户
func GetAllUsers(c *gin.Context) {
	allUsers, err := pkg.GetAllUserNames()
	if err != nil {
		c.JSON(200, gin.H{
			"code":    500,
			"message": "获取失败:" + err.Error(),
		})
	}
	c.JSON(200, gin.H{
		"code": 200,
		// "message": "成功！",
		"users": allUsers,
	})
}

// 更新用户信息
func UpdateUserInfo(c *gin.Context) {
	// 将用户信息存入mysql数据库
	var userJson user
	var user model.MysqlUser
	// 将表单数据转为JSON数据
	if err := c.ShouldBindJSON(&userJson); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 赋值
	var err error
	user.UserID, err = pkg.GetUserID(userJson.Username)
	user.Telephone = userJson.Telephone
	user.Realname = userJson.Realname
	user.CardID = userJson.CardID
	user.IsPass = userJson.IsPass
	fmt.Print(user)
	err = pkg.UpdateUser(&user)
	// log.Println(user)
	if err != nil {
		// log.Println("Error inserting user into MySQL database:", err)
		// 返回错误信息给客户端或者记录日志
		c.JSON(200, gin.H{
			"code":    500,
			"message": "更新失败!",
		})
		return
	}
	c.JSON(200, gin.H{
		"code":    200,
		"message": "更新成功！",
	})
}

// 获取用户信息
func GetInfo(c *gin.Context) {

	userID, exist := c.Get("userID")
	if !exist {
		c.JSON(200, gin.H{
			"code":    200,
			"message": "get user type failed",
		})
	}

	userType, err := GetUserType(userID.(string))
	if err != nil {
		c.JSON(200, gin.H{
			"code":    200,
			"message": "get user type failed" + err.Error(),
		})
	}

	username, telephone, realname, cardID, txID, isPass, err := pkg.GetUsername(userID.(string))
	if err != nil {
		c.JSON(200, gin.H{
			"code":    200,
			"message": "获取信息失败：" + err.Error(),
		})
	}

	var res res

	res.Username = username
	res.Telephone = telephone
	res.Realname = realname
	res.CardID = cardID
	res.TxID = txID
	res.Usertype = userType
	res.IsPass = isPass

	c.JSON(200, gin.H{
		"code":    200,
		"message": "success",
		"result":  res,
	})
}
