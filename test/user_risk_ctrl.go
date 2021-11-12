package main

import (
	"github.com/gin-gonic/gin"
	"strconv"
	"time"
)

func slideBar(c *gin.Context) {

} //滑动条

func wait(c *gin.Context) {

} //等待一段时间重试

func ban(c *gin.Context) {

} //禁止用户后续操作

func timeLimitBan(c *gin.Context) {
	//TODO 限时禁用IP
} //在一定时间内禁用IP

func riskCtrl(cookie string) gin.HandlerFunc {
	return func(c *gin.Context) {
		//c.Next()
		avgFreq, _, err := spy(cookie, c.ClientIP(), strconv.FormatInt(time.Now().UnixNano(), 10), c.Request.Method)
		if err != -1 {
			if c.Request.Method == "GET" {
				timeLimitBan(c)
				return
			}
			if c.Request.Method == "POST" {
				ban(c)
				return
			}
		}
		if c.Request.Method == "POST" {
			if avgFreq < 20/2 {
				slideBar(c)
				return
			} else {
				//checkUser(c)
				return
			}
		}
	}
}
