package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	_ "gorm.io/driver/mysql"
	"gorm.io/gorm"
	_ "gorm.io/gorm"
	"math/rand"
	"net/http"
	"time"
)
type User struct{
	Username string
	Password string
	Phone string
}
func (User) TableName() string{
	return "user"
}
//修改对应的用户名、密码和数据库，格式如下：
//dsn = "user:password@tcp(127.0.0.1:3306)/database?charset=utf8mb4&parseTime=True&loc=Local"
const dsn = "root:123456@tcp(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=True&loc=Local"
func main() {
	c1 := make(chan string, 100)
	c2 := make(chan string, 100)
	c3 := make(chan string, 100)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err!= nil {
		fmt.Println(err)
	}

	r := gin.Default()
	r.LoadHTMLFiles("./login_username.html", "./index.html", "./login_phone.html", "./register.html")

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login_username.html", nil)
		ip, _ := c.RemoteIP()
		fmt.Print(ip)
	}) //用户名登录的get实现

	r.GET("/login_phone", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login_phone.html", nil)
	})//手机号登录的get实现

	r.GET("/index", func(c *gin.Context){
		c.HTML(http.StatusOK, "index.html", gin.H{
			"Name":     <-c1,
			"Phone":    <-c2,
			"Password": <-c3,
		})
	})//主页的get实现

	r.GET("/register", func(c *gin.Context){
		c.HTML(http.StatusOK, "register.html", nil)
	})//注册界面的get实现

	r.POST("/", func(c *gin.Context) {
		username := c.PostForm("username")
		password := c.PostForm("password")
		//此处需要统计用户IP和ID并返回此用户操作频次
		//此处需要根据用户操作情况决定是否进行安全防护
		//此处应实现用户名密码比对，并提取对应电话号码
		c1<-username
		c2<-password

		if !sign_username(username,password) {
			c.JSON(401,gin.H{
				"success":false,
				"msg":"用户名密码不正确!",
			})
			c.Request.URL.Path, c.Request.Method = "/", "GET"
			r.HandleContext(c)
			return
		}
		c.Request.URL.Path, c.Request.Method = "/index", "GET"
		r.HandleContext(c)
	})//用户名登录的post提交处理

	r.POST("/login_phone", func(c *gin.Context) {
		phone := c.PostForm("phone")
		vcode := c.PostForm("vcode")
		//此处需要统计用户IP和ID并返回此用户操作频次
		//此处需要根据用户操作情况决定是否进行安全防护
		//此处应实现电话号码和验证码比对，并提取对应用户名
		c3<-phone

		if !sign_phone(phone) {

			c.JSON(401,gin.H{
				"success":false,
				"msg":"手机号未注册!",
			})
			c.Request.URL.Path, c.Request.Method = "/login_phone", "GET"
			r.HandleContext(c)
			return
		}
		//6位随机验证码生成
		rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
		vcode = fmt.Sprintf("%06v", rnd.Int31n(1000000))
		fmt.Print(vcode)
		c.JSON(200,gin.H{
			"success":true,
			"msg":"手机号登录成功!",
			"vcode":vcode,
		})
		c.Request.URL.Path, c.Request.Method = "/index", "GET"
		r.HandleContext(c)
	})//手机号登录的表单处理

	r.POST("/register", func(c *gin.Context){
		username := c.PostForm("username")
		password := c.PostForm("password")
		password2 := c.PostForm("password2")
		phone := c.PostForm("phone")
		vcode := c.PostForm("vcode")
		//此处需要统计用户IP和ID并返回此用户操作频次
		//此处需要根据用户操作情况决定是否进行安全防护
		//此处应当实现对新用户信息的存储

		var user User

		db.Where("username = ?",username).First(&user)
		//fmt.Println(user," qwq")
		if user.Username == username {
			c.JSON(401,gin.H{
				"success":false,
				"msg":"用户已注册!",

			})
		c.Request.URL.Path, c.Request.Method = "/register", "GET"
		r.HandleContext(c)
		return
		}
		if password2 == "" {
			c.JSON(401,gin.H{
				"success":false,
				"msg":"请输入确认密码!",
			})
			c.Request.URL.Path, c.Request.Method = "/register", "GET"
			r.HandleContext(c)
			return
		}
		if password != password2 {
			c.JSON(401,gin.H{
				"success":false,
				"msg":"确认密码与原密码不匹配!",
			})
			c.Request.URL.Path, c.Request.Method = "/register", "GET"
			r.HandleContext(c)
			return
		}
		save(username,phone,password)
		fmt.Println(username,password,password2,phone,vcode)
		//6位随机验证码生成
		rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
		vcode = fmt.Sprintf("%06v", rnd.Int31n(1000000))

		fmt.Print(vcode)
		c.JSON(200,gin.H{
			"success":true,
			"msg":"注册成功!",
			"vcode":vcode,
		})
		c.Request.URL.Path, c.Request.Method = "/", "GET"
		r.HandleContext(c)
	})//注册的表单处理

	r.POST("/index", func(c *gin.Context){
		logout := c.PostForm("logout")
		phone := c.PostForm("phone")
		password := c.PostForm("password")
		username := c.PostForm("username")

		//此处应当根据logout的值做分支处理，若为2则需要删除此用户信息
		fmt.Print(logout)
		//logout == "1"的登出逻辑应该不用实现？
		if logout == "2" {
			if sign_username(username,password) {
				delete(phone)
				c.JSON(200,gin.H{
					"success":200,
					"msg":"删除成功!",
				})
			}
		}
		c.Request.URL.Path, c.Request.Method = "/", "GET"
		r.HandleContext(c)
	})//注册的表单处理

	r.Run(":8080")
}