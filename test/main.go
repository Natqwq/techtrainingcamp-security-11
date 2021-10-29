package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

func main() {
	c1 := make(chan string, 100)
	c2 := make(chan string, 100)
	c3 := make(chan string, 100)
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
		fmt.Print(vcode)
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
		fmt.Print(username,password,password2,phone,vcode)
		c.Request.URL.Path, c.Request.Method = "/", "GET"
		r.HandleContext(c)
	})//注册的表单处理

	r.POST("/index", func(c *gin.Context){
		logout := c.PostForm("logout")
		//此处应当根据logout的值做分支处理，若为2则需要删除此用户信息
		fmt.Print(logout)
		c.Request.URL.Path, c.Request.Method = "/", "GET"
		r.HandleContext(c)
	})//注册的表单处理

	r.Run(":8080")
}