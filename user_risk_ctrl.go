package main

import (
	"github.com/gin-gonic/gin"
	"strconv"
	"time"
)

const limit = 5

func slide_bar(){
	print("实现滑动窗口 ！")
}//滑动条

func check_user(c *gin.Context){
	//对用户进行验证码验证
}//需要用手机验证码验证身份

func ban(c *gin.Context){
	//TODO 禁用该IP
}//禁止用户后续操作

func time_limit_ban(c *gin.Context) {
	//TODO 限时禁用IP
}//在一定时间内禁用IP

func risk_ctrl(cookie string) gin.HandlerFunc {
	return func (c *gin.Context){
		//c.Next()
		avgFreq, _, err := spy(cookie, c.ClientIP(), strconv.FormatInt(time.Now().UnixNano(),10), c.Request.Method)
		if err != -1 {
			if c.Request.Method == "GET" {
				time_limit_ban(c)
				return
			}
			if c.Request.Method == "POST" {
				ban(c)
				return
			}
		}
		if c.Request.Method == "POST" {
			if avgFreq < limit / 2 {
				slide_bar()
				return
			} else {
				check_user(c)
				return
			}
		}
	}
}

