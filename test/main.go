package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	_ "gorm.io/driver/mysql"
	"gorm.io/gorm"
	_ "gorm.io/gorm"
	"net/http"
	"strconv"
	"time"
)

//用户表
type User struct {
	Username string `form:"user" json:"UserName" binding:"required"`
	Password string `form:"password" json:"Password" binding:"required"`
	Phone    string `form:"phone" json:"PhoneNumber" binding:"required"`
}

//post请求体
type Post struct {
	Username        string      `form:"user" json:"UserName" binding:"required"`
	Password        string      `form:"password" json:"Password" binding:"required"`
	Phone           string      `form:"phone" json:"PhoneNumber" binding:"required"`
	Vcode           string      `form:"vcode" json:"VerifyCode" binding:"required"`
	EnvironmentBase Environment `form:"environment" json:"Environment" binding:"required"`
	Logout          string      `form:"logout" json:"Logout" binding:"required"`
}

//验证码
type CheckVcode struct {
	Phone     string `form:"phone" json:"PhoneNumber" binding:"required"`
	Vcode     string
	Create    string
	Create_at time.Time
}

//用户最近一次登录
type Device struct {
	Username   string
	Deviceid   string
	Ip         string
	Logintime  time.Time
	Logouttime time.Time
}

//用户环境
type Environment struct {
	Deviceid string
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

//修改对应的用户名、密码和数据库，格式如下：
//dsn = "user:password@tcp(127.0.0.1:3306)/database?charset=utf8mb4&parseTime=True&loc=Local"
const dsn = "catchyou:123456@tcp(49.234.79.216:3306)/catchYou?charset=utf8mb4&parseTime=True&loc=Local"

func main() {
	c1 := make(chan string, 100)
	c2 := make(chan string, 100)
	c3 := make(chan string, 100)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	//如果数据库不存在表，则自动创表
	db.AutoMigrate(&User{}, &CheckVcode{}, &Device{})

	if err != nil {
		fmt.Println(err)
	}

	r := gin.Default()

	//配置静态资源文件
	r.LoadHTMLGlob("./static/**/*")
	r.Static("/static", "./static")

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login_username.html", nil)
		cookie, _ := c.Cookie("DeviceID")
		print(spy(cookie, c.ClientIP(), strconv.FormatInt(time.Now().UnixNano(), 10), c.Request.Method))
	}) //用户名登录的get实现

	r.GET("/login_phone", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login_phone.html", nil)
		cookie, _ := c.Cookie("DeviceID")
		print(spy(cookie, c.ClientIP(), strconv.FormatInt(time.Now().UnixNano(), 10), c.Request.Method))
	}) //手机号登录的get实现

	r.GET("/index", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
		cookie, _ := c.Cookie("DeviceID")
		print(spy(cookie, c.ClientIP(), strconv.FormatInt(time.Now().UnixNano(), 10), c.Request.Method))
	}) //主页的get实现

	r.GET("/register", func(c *gin.Context) {
		c.HTML(http.StatusOK, "register.html", nil)
		cookie, _ := c.Cookie("DeviceID")
		print(spy(cookie, c.ClientIP(), strconv.FormatInt(time.Now().UnixNano(), 10), c.Request.Method))
	}) //注册界面的get实现

	r.POST("/getVCode", func(c *gin.Context) {
		cookie, _ := c.Cookie("DeviceID")
		print(spy(cookie, c.ClientIP(), strconv.FormatInt(time.Now().UnixNano(), 10), c.Request.Method))

		var checkVcode CheckVcode
		c.ShouldBindJSON(&checkVcode)
		code := create()                  //获得随机验证码
		saveVcode(checkVcode.Phone, code) //存到数据库
		c.JSON(200, gin.H{
			"success": true,
			"msg":     "发送成功！",
		})
		return
	})

	r.POST("/", func(c *gin.Context) {
		cookie, _ := c.Cookie("DeviceID")
		print(spy(cookie, c.ClientIP(), strconv.FormatInt(time.Now().UnixNano(), 10), c.Request.Method))
		//此处需要统计用户IP和ID并返回此用户操作频次
		//此处需要根据用户操作情况决定是否进行安全防护
		//此处应实现用户名密码比对，并提取对应电话号码

		var json Post
		c.ShouldBindJSON(&json)
		c1 <- json.Username
		c2 <- json.Password
		print(c1, c2)
		if !sign_username(json.Username, json.Password) {
			c.JSON(401, gin.H{
				"success": false,
				"msg":     "用户名密码不正确!",
			})
			return
		}
		var device Device
		device.Username, device.Logintime = json.Username, time.Now()
		device.Deviceid = json.EnvironmentBase.Deviceid
		device.Ip = c.ClientIP()
		device.Logouttime = time.Now().Add(time.Minute * 1440)
		var nowDevice Device
		db.Where("username = ?", device.Username).First(&nowDevice)
		if nowDevice.Username != device.Username || nowDevice.Logouttime.Sub(device.Logintime) < 0 {
			db.Create(&device)
		} else {
			db.Where("username = ?", device.Username).Updates(Device{
				Username:   device.Username,
				Ip:         device.Ip,
				Deviceid:   device.Deviceid,
				Logintime:  device.Logintime,
				Logouttime: device.Logouttime,
			})
		}
		fmt.Println("device=", device)
		fmt.Println("ip= ", device.Ip)
		fmt.Println("deviceid= ", device.Deviceid)
		c.JSON(200, gin.H{
			"success": true,
			"msg":     "登录成功！",
		})
	}) //用户名登录的post提交处理

	r.POST("/login_phone", func(c *gin.Context) {
		cookie, _ := c.Cookie("DeviceID")
		print(spy(cookie, c.ClientIP(), strconv.FormatInt(time.Now().UnixNano(), 10), c.Request.Method))
		//此处需要统计用户IP和ID并返回此用户操作频次
		//此处需要根据用户操作情况决定是否进行安全防护
		//此处应实现电话号码和验证码比对，并提取对应用户名

		var json Post
		c.ShouldBindJSON(&json)
		c3 <- json.Phone
		if !sign_phone(json.Phone) {
			c.JSON(401, gin.H{
				"success": false,
				"msg":     "手机号未注册!",
			})
			return

		}
		//6位随机验证码生成
		if !checkVcode(json) {
			c.JSON(401, gin.H{
				"success": false,
				"msg":     "验证码错误或过期，请重试!",
			})
			return
		}
		c.JSON(200, gin.H{
			"success": true,
			"msg":     "手机号登录成功!",
		})
	}) //手机号登录的表单处理

	/*
	* 注册请求
	 */
	r.POST("/register", func(c *gin.Context) {
		cookie, _ := c.Cookie("DeviceID")
		print(spy(cookie, c.ClientIP(), strconv.FormatInt(time.Now().UnixNano(), 10), c.Request.Method))
		//此处需要统计用户IP和ID并返回此用户操作频次
		//此处需要根据用户操作情况决定是否进行安全防护
		//此处应当实现对新用户信息的存储

		var user User
		var json Post

		c.ShouldBindJSON(&json)

		fmt.Println(json)
		db.Where("username = ?", json.Username).First(&user)
		//fmt.Println(user," qwq")
		if user.Username == json.Username {
			c.JSON(401, gin.H{
				"success": false,
				"msg":     "该用户名已注册!",
			})
			return
		} else if !checkVcode(json) {
			c.JSON(401, gin.H{
				"success": false,
				"msg":     "验证码错误!",
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
	})

	/**
	* 登出请求
	**/
	r.POST("/logout", func(c *gin.Context) {
		cookie, _ := c.Cookie("DeviceID")
		print(spy(cookie, c.ClientIP(), strconv.FormatInt(time.Now().UnixNano(), 10), c.Request.Method))
		var json Post
		//此处应当根据logout的值做分支处理，若为2则需要删除此用户信息
		c.ShouldBindJSON(&json)
		logoutMethod := json.Logout

		if logoutMethod == "1" {
			//登出操作，更新数据库中对应设备的登出时间
			deviceId, _ := c.Cookie("DeviceID")
			db.Where("deviceid = ?", deviceId).UpdateColumns(Device{Logouttime: time.Now()})
			c.JSON(200, gin.H{
				"success": 200,
				"msg":     "登出成功!",
			})
		}
		if logoutMethod == "2" {
			//注销操作，删除数据库中账户
			if sign_username(json.Username, json.Password) {
				delete(json.Phone)
				c.JSON(200, gin.H{
					"success": 200,
					"msg":     "删除成功!",
				})
			}
		}
	})

	r.Run(":8080")

}
