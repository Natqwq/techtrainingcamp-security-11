package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	_ "gorm.io/driver/mysql"
	"gorm.io/gorm"
	_ "gorm.io/gorm"
	"net/http"
	"time"
)

// User 用户表
type User struct {
	Username string `form:"user" json:"UserName" binding:"required"`
	Password string `form:"password" json:"Password" binding:"required"`
	Phone    string `form:"phone" json:"PhoneNumber" binding:"required"`
}

// Post post请求体
type Post struct {
	Username        string      `form:"user" json:"UserName" binding:"required"`
	Password        string      `form:"password" json:"Password" binding:"required"`
	Phone           string      `form:"phone" json:"PhoneNumber" binding:"required"`
	Vcode           string      `form:"vcode" json:"VerifyCode" binding:"required"`
	EnvironmentBase Environment `form:"environment" json:"Environment" binding:"required"`
	Logout          string      `form:"logout" json:"Logout" binding:"required"`
}

// CheckVcode 验证码
type CheckVcode struct {
	Phone      string `form:"phone" json:"PhoneNumber" binding:"required"`
	Vcode      string
	Create     string
	CreateAt   time.Time
	HasExpired bool
}

// Device 用户最近一次登录
type Device struct {
	Username   string
	DeviceID   string
	Ip         string
	LoginTime  time.Time
	LogoutTime time.Time
}

// Environment 用户环境
type Environment struct {
	DeviceID string
}

func (User) TableName() string {
	return "user"
}
func (CheckVcode) TableName() string {
	return "chk"
}
func (Device) TableName() string {
	return "device"
}

func main() {
	c1 := make(chan string, chanBuf)
	c2 := make(chan string, chanBuf)
	c3 := make(chan string, chanBuf)
	db, err := gorm.Open(mysql.Open(mysqlInfo), &gorm.Config{})

	//如果数据库不存在表，则自动创表
	_ = db.AutoMigrate(&User{}, &CheckVcode{}, &Device{})

	if err != nil {
		fmt.Println(err)
	}

	r := gin.Default()

	//配置静态资源文件
	r.LoadHTMLGlob("./static/**/*")
	r.Static("/static", "./static")

	r.GET("/", func(c *gin.Context) {
		isAllow, spyMethod := riskCtrl(c)
		if isAllow {
			c.HTML(http.StatusOK, "login_username.html", gin.H{
				"spyMethod": spyMethod,
			})
		} else {
			c.JSON(http.StatusLocked, gin.H{
				"message": "Too many requests, page locked!",
			})
		}
	}) //用户名登录的get实现

	r.GET("/login_phone", func(c *gin.Context) {
		isAllow, spyMethod := riskCtrl(c)
		if isAllow {
			c.HTML(http.StatusOK, "login_phone.html", gin.H{
				"spyMethod": spyMethod,
			})
		} else {
			c.JSON(http.StatusLocked, gin.H{
				"message": "Too many requests, page locked!",
			})
		}
	}) //手机号登录的get实现

	r.GET("/index", func(c *gin.Context) {
		cookie, _ := c.Cookie("DeviceID")
		var device Device
		db.Where("device_id = ?", cookie).First(&device)
		if cookie == "" || device.LogoutTime.Sub(time.Now()) <= 0 || device.DeviceID != cookie {
			c.HTML(403, "index.html", nil)
			return
		}
		isAllow, spyMethod := riskCtrl(c)
		if isAllow {
			c.HTML(http.StatusOK, "index.html", gin.H{
				"spyMethod": spyMethod,
			})
		} else {
			c.JSON(http.StatusLocked, gin.H{
				"message": "Too many requests, page locked!",
			})
		}
	}) //主页的get实现

	r.GET("/register", func(c *gin.Context) {
		isAllow, spyMethod := riskCtrl(c)
		if isAllow {
			c.HTML(http.StatusOK, "register.html", gin.H{
				"spyMethod": spyMethod,
			})
		} else {
			c.JSON(http.StatusLocked, gin.H{
				"message": "Too many requests, page locked!",
			})
		}
	}) //注册界面的get实现

	r.POST("/getVCode", func(c *gin.Context) {
		isAllow, spyMethod := riskCtrl(c)
		if isAllow {
			var checkVcode CheckVcode
			_ = c.ShouldBindJSON(&checkVcode)
			code := create()                  //获得随机验证码
			saveVcode(checkVcode.Phone, code) //存到数据库
			c.JSON(200, gin.H{
				"success":   true,
				"msg":       "发送成功！",
				"spyMethod": spyMethod,
			})
			return
		} else {
			c.JSON(http.StatusLocked, gin.H{
				"message": "Too many requests, page locked!",
			})
		}
	})

	r.POST("/", func(c *gin.Context) {
		isAllow, spyMethod := riskCtrl(c)
		if isAllow {

			var json Post
			_ = c.ShouldBindJSON(&json)
			c1 <- json.Username
			c2 <- json.Password
			print(c1, c2)
			if !signUsername(json.Username, json.Password) {
				c.JSON(401, gin.H{
					"success":   false,
					"msg":       "用户名密码不正确!",
					"spyMethod": spyMethod,
				})
				return
			}
			var device Device
			device.Username, device.LoginTime = json.Username, time.Now()
			device.DeviceID = json.EnvironmentBase.DeviceID
			device.Ip = c.ClientIP()
			device.LogoutTime = time.Now().Add(time.Minute * 1440)
			var nowDevice Device
			db.Where("username = ?", device.Username).First(&nowDevice)
			if nowDevice.Username != device.Username || nowDevice.LogoutTime.Sub(device.LoginTime) < 0 {
				db.Create(&device)
			} else {
				db.Where("username = ?", device.Username).Updates(Device{
					Username:   device.Username,
					Ip:         device.Ip,
					DeviceID:   device.DeviceID,
					LoginTime:  device.LoginTime,
					LogoutTime: device.LogoutTime,
				})
			}
			fmt.Println("device=", device)
			fmt.Println("ip= ", device.Ip)
			fmt.Println("DeviceID= ", device.DeviceID)
			c.JSON(200, gin.H{
				"success": true,
				"msg":     "登录成功！",
			})
		} else {
			c.JSON(http.StatusLocked, gin.H{
				"message": "Too many requests, page locked!",
			})
		}
	}) //用户名登录的post提交处理

	r.POST("/login_phone", func(c *gin.Context) {
		isAllow, spyMethod := riskCtrl(c)
		if isAllow {
			var json Post
			_ = c.ShouldBindJSON(&json)
			c3 <- json.Phone
			if !signPhone(json.Phone) {
				c.JSON(401, gin.H{
					"success":   false,
					"msg":       "手机号未注册!",
					"spyMethod": spyMethod,
				})
				return
			}
			//6位随机验证码生成
			if !checkVcode(json) {
				c.JSON(401, gin.H{
					"success":   false,
					"msg":       "验证码错误或过期，请重试!",
					"spyMethod": spyMethod,
				})
				return
			}
			c.JSON(200, gin.H{
				"success": true,
				"msg":     "手机号登录成功!",
			})
		} else {
			c.JSON(http.StatusLocked, gin.H{
				"message": "Too many requests, page locked!",
			})
		}
	}) //手机号登录的表单处理

	/*
	* 注册请求
	 */
	r.POST("/register", func(c *gin.Context) {
		isAllow, spyMethod := riskCtrl(c)
		if isAllow {
			//此处需要统计用户IP和ID并返回此用户操作频次
			//此处需要根据用户操作情况决定是否进行安全防护
			//此处应当实现对新用户信息的存储

			var user User
			var json Post

			_ = c.ShouldBindJSON(&json)

			fmt.Println(json)
			db.Where("username = ?", json.Username).First(&user)
			//fmt.Println(user," qwq")
			if user.Username == json.Username {
				c.JSON(401, gin.H{
					"success":   false,
					"msg":       "该用户名已注册!",
					"spyMethod": spyMethod,
				})
				return
			} else if !checkVcode(json) {
				c.JSON(401, gin.H{
					"success":   false,
					"msg":       "验证码错误!",
					"spyMethod": spyMethod,
				})
				return
			} else {
				save(json.Username, json.Phone, json.Password)
				println("保存用户成功")
				c.JSON(200, gin.H{
					"success": true,
					"msg":     "注册成功!",
				})
			}
		} else {
			c.JSON(http.StatusLocked, gin.H{
				"message": "Too many requests, page locked!",
			})
		}
	})

	/**
	* 登出请求
	**/
	r.POST("/logout", func(c *gin.Context) {
		isAllow, spyMethod := riskCtrl(c)
		if isAllow {
			var json Post
			//此处应当根据logout的值做分支处理，若为2则需要删除此用户信息
			_ = c.ShouldBindJSON(&json)
			logoutMethod := json.Logout

			if logoutMethod == "1" {
				//登出操作，更新数据库中对应设备的登出时间
				DeviceID, _ := c.Cookie("DeviceID")
				db.Where("device_id = ?", DeviceID).UpdateColumns(Device{LogoutTime: time.Now()})
				c.JSON(200, gin.H{
					"success":   200,
					"msg":       "登出成功!",
					"spyMethod": spyMethod,
				})
			}
			if logoutMethod == "2" {
				//注销操作，删除数据库中账户
				if signUsername(json.Username, json.Password) {
					delete(json.Phone)
					c.JSON(200, gin.H{
						"success":   200,
						"msg":       "删除成功!",
						"spyMethod": spyMethod,
					})
				}
			}
		} else {
			c.JSON(http.StatusLocked, gin.H{
				"message": "Too many requests, page locked!",
			})
		}
	})

	_ = r.Run(":8080")

}
