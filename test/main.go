package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	_ "gorm.io/driver/mysql"
	"gorm.io/gorm"
	_ "gorm.io/gorm"
	"net/http"
)

type User struct {
	Username string `form:"user" json:"UserName" binding:"required"`
	Password string `form:"password" json:"Password" binding:"required"`
	Phone    string `form:"phone" json:"PhoneNumber" binding:"required"`
}

//用于接收json的结构体
type Post struct {
	Username string `form:"user" json:"UserName" binding:"required"`
	Password string `form:"password" json:"Password" binding:"required"`
	Phone    string `form:"phone" json:"PhoneNumber" binding:"required"`
	Vcode    string `form:"vcode" json:"VerifyCode" binding:"required"`
}
type CheckVcode struct {
	Phone  string `form:"phone" json:"PhoneNumber" binding:"required"`
	Vcode  string
	Create string
}

func (User) TableName() string {
	return "user"
}
func (CheckVcode) TableName() string {
	return "chk"
}

//修改对应的用户名、密码和数据库，格式如下：
//dsn = "user:password@tcp(127.0.0.1:3306)/database?charset=utf8mb4&parseTime=True&loc=Local"
const dsn = "root:123456@tcp(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=True&loc=Local"

func main() {
	c1 := make(chan string, 100)
	c2 := make(chan string, 100)
	c3 := make(chan string, 100)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	db.AutoMigrate(&User{}, &CheckVcode{}) //如果数据库不存在表，则自动创表

	if err != nil {
		fmt.Println(err)
	}

	r := gin.Default()

	//配置静态资源文件
	r.LoadHTMLGlob("./static/**/*")
	r.Static("/static", "./static")

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login_username.html", nil)
		ip, _ := c.RemoteIP()
		fmt.Print(ip)
	}) //用户名登录的get实现

	r.GET("/login_phone", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login_phone.html", nil)
	}) //手机号登录的get实现

	r.GET("/index", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	}) //主页的get实现

	r.GET("/register", func(c *gin.Context) {
		c.HTML(http.StatusOK, "register.html", nil)
	}) //注册界面的get实现

	/*TODO：以下为简易版本，后面需要接入风控！*/
	r.POST("/getVCode", func(c *gin.Context) {
		var checkVcode CheckVcode
		c.ShouldBindJSON(&checkVcode)
		code := create()                  //获得随机验证码
		saveVcode(checkVcode.Phone, code) //存到数据库
		c.JSON(200, gin.H{
			"success": true,
			"code":    code,
		})
		return
	})

	r.POST("/", func(c *gin.Context) {
		//username := c.PostForm("UserName")
		//password := c.PostForm("Password")
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
		c.JSON(200, gin.H{
			"success": true,
			"msg":     "登录成功！",
		})
	}) //用户名登录的post提交处理

	r.POST("/login_phone", func(c *gin.Context) {
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
				"msg":     "验证码验证失败!",
			})
			return
		}
		//fmt.Print(vcode)
		c.JSON(200, gin.H{
			"success":    true,
			"msg":        "手机号登录成功!",
			"VerifyCode": json.Vcode,
		})
	}) //手机号登录的表单处理

	r.POST("/register", func(c *gin.Context) {
		//username := c.PostForm("UserName")
		//password := c.PostForm("Password")
		////password2 := c.PostForm("password2")
		//phone := c.PostForm("PhoneNumber")
		//vcode := c.PostForm("VerifyCode")
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
				"msg":     "用户已注册!",
			})
			return
		}

		save(json.Username, json.Phone, json.Password)

		//if !checkVcode(jvcode,json.Phone) {
		//	c.JSON(401,gin.H{
		//		"success":false,
		//		"msg":"验证码验证失败!",
		//	})
		//	c.Request.URL.Path, c.Request.Method = "/login_phone", "GET"
		//	r.HandleContext(c)
		//	return
		//}

		//fmt.Print(vcode)
		c.JSON(200, gin.H{
			"success":    true,
			"msg":        "注册成功!",
			"VerifyCode": json.Vcode,
			"status":     200,
		})
		//c.Request.URL.Path, c.Request.Method = "/", "GET"
		//r.HandleContext(c)
	}) //注册的表单处理

	r.POST("/logout", func(c *gin.Context) {
		//logout := c.PostForm("logout")
		//phone := c.PostForm("PhoneNumber")
		//password := c.PostForm("Password")
		//username := c.PostForm("UserName")
		var json Post
		//此处应当根据logout的值做分支处理，若为2则需要删除此用户信息
		c.ShouldBindJSON(&json)
		logout := "2"
		fmt.Print(logout)
		//logout == "1"的登出逻辑应该不用实现？
		if logout == "2" {
			if sign_username(json.Username, json.Password) {
				delete(json.Phone)
				c.JSON(200, gin.H{
					"success": 200,
					"msg":     "删除成功!",
				})
			}
		}
	}) //注册的表单处理

	r.Run(":8080")
}
