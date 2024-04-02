package controller

import (
	"backend/model"
	"backend/pkg"

	// "log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type user struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Usertype string `json:"userType"`
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

// func GetAllUserInfo(c *gin.Context) {
// 	users, err := GetAllUsers()
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{
// 			"code":    http.StatusInternalServerError,
// 			"message": "failed to get all users: " + err.Error(),
// 		})
// 		return
// 	}

// 	var userInfos []gin.H
// 	for _, user := range users {
// 		userType, err := GetUserType(user.ID)
// 		if err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{
// 				"code":    http.StatusInternalServerError,
// 				"message": "failed to get user type for user " + user.ID + ": " + err.Error(),
// 			})
// 			return
// 		}

// 		username, err := pkg.GetUsername(user.ID)
// 		if err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{
// 				"code":    http.StatusInternalServerError,
// 				"message": "failed to get username for user " + user.ID + ": " + err.Error(),
// 			})
// 			return
// 		}

// 		userInfo := gin.H{
// 			"userType": userType,
// 			"userName": username,
// 		}
// 		userInfos = append(userInfos, userInfo)
// 	}

// 	c.JSON(http.StatusOK, gin.H{
// 		"code":      http.StatusOK,
// 		"message":   "get all users information success",
// 		"userInfos": userInfos,
// 	})
// }

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

	username, err := pkg.GetUsername(userID.(string))
	if err != nil {
		c.JSON(200, gin.H{
			"code":    200,
			"message": "get user name failed" + err.Error(),
		})
	}

	c.JSON(200, gin.H{
		"code":     200,
		"message":  "get user type success",
		"userType": userType,
		"userName": username,
	})
}
