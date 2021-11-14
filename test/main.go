package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"gorm.io/driver/mysql"
	_ "gorm.io/driver/mysql"
	"gorm.io/gorm"
	_ "gorm.io/gorm"
	"net/http"
	"time"
)

// User 用户表
type User struct {
	ID       uint
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
	ID         uint
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
	//打开数据库连接对象
	db, err := gorm.Open(mysql.Open(mysqlInfo), &gorm.Config{})

	//如果数据库不存在表，则自动创表
	_ = db.AutoMigrate(&User{}, &CheckVcode{}, &Device{})

	//打开redis连接对象
	rdb, _ := redis.Dial("tcp", redisIP)
	_, _ = rdb.Do("AUTH", redisPsd)

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
		isAllow, spyMethod := riskCtrl(c)

		//获得浏览器中的JSESSIONID
		JSESSIONID, err := c.Cookie("JSESSIONID")
		auth, _ := redis.Bool(rdb.Do("Exists", JSESSIONID))
		if err != nil || !auth {
			c.JSON(http.StatusUnauthorized, gin.H{
				"msg": "未登录！",
			})
			return
		}
		if isAllow {
			//更新用户redis中该连请求的状态
			rdb.Do("EXPIRE", JSESSIONID, 60*60)
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
			//是否通过验证码认证
			spyCode := vCodeSpy(checkVcode)
			if spyCode == 0 || spyCode == 2 {
				code := create() //获得随机验证码
				saveVcode(checkVcode.Phone, code)
				//存到数据库
				if spyCode == 2 {
					spyMethod = 2
				}
				c.JSON(200, gin.H{
					"success":   true,
					"msg":       "发送成功！",
					"spyMethod": spyMethod,
				})
			} else {
				c.JSON(http.StatusBadRequest, gin.H{
					"msg":       "验证码发送次数过多，请稍后再试！",
					"spyMethod": spyMethod,
				})
			}
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
			//生成JSESSIONID保存在redis中
			u1, _ := uuid.NewUUID()
			http.SetCookie(c.Writer, &http.Cookie{
				Name:  "JSESSIONID",
				Value: u1.String(),
			})
			rdb.Do("SET", u1.String(), json.Username)
			rdb.Do("EXPIRE", u1.String(), 60*60)
			c.JSON(200, gin.H{
				"success": true,
				"msg":     "登录成功！",
			})
		} else {
			print(1)
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
			//生成JSESSIONID保存在redis中
			u1, _ := uuid.NewUUID()
			http.SetCookie(c.Writer, &http.Cookie{
				Name:  "JSESSIONID",
				Value: u1.String(),
			})
			//c.SetCookie("JSESSIONID", u1.String(), , "/", "localhost", false, false)
			rdb.Do("SET", u1, json.Phone)
			rdb.Do("EXPIRE", u1.String(), 60*60)
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
		JSESSIONID, _ := c.Cookie("JSESSIONID")
		if isAllow {
			var json Post
			//此处应当根据logout的值做分支处理，若为2则需要删除此用户信息
			_ = c.ShouldBindJSON(&json)
			//删除redis中的JSESSIONID
			userAuth, err := redis.String(rdb.Do("GET", JSESSIONID))
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{
					"msg": "请先登录！",
				})
				return
			}

			logoutMethod := json.Logout
			if logoutMethod == "1" {
				c.JSON(200, gin.H{
					"success":   200,
					"msg":       "登出成功!",
					"spyMethod": spyMethod,
				})
			}
			if logoutMethod == "2" {
				//注销操作，删除数据库中账户
				if err != nil {
					c.JSON(500, gin.H{
						"success":   500,
						"msg":       "删除错误!",
						"spyMethod": spyMethod,
					})
					return
				}
				delete(userAuth)
				c.JSON(200, gin.H{
					"success":   200,
					"msg":       "删除成功!",
					"spyMethod": spyMethod,
				})
			}
			rdb.Do("DEL", JSESSIONID)
		} else {
			c.JSON(http.StatusLocked, gin.H{
				"message": "Too many requests, page locked!",
			})
		}
	})

	_ = r.Run(":8080")

}
