package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
	"net/http"
	"strconv"
	"time"
)

func slideBar(c *gin.Context) {
	c.JSON(http.StatusTooManyRequests, gin.H{
		"slideBar": true,
		"message":  "SlideBar Needed",
	})
} //滑动条

func limitBan(id string, lim *rate.Limiter, level int) {
	if level <= 3 {
		lim.SetBurst(4)
		_, rc.err = rc.cnt.Do("EXPIRE", id, 300)
	} else if level <= 8 {
		lim.SetBurst(2)
		_, rc.err = rc.cnt.Do("EXPIRE", id, 600)
	} else if level <= 15 {
		lim.SetBurst(1)
		_, rc.err = rc.cnt.Do("EXPIRE", id, 1200)
	} else {
		lim.SetLimit(0)
		lim.SetBurst(0)
		_, rc.err = rc.cnt.Do("EXPIRE", id, 3600)
	}
} //分级禁止

func riskCtrl(c *gin.Context) bool {
	id, _ := c.Cookie("DeviceID")
	method, _ := c.Request.Method, c.ClientIP()
	now := strconv.FormatInt(time.Now().UnixNano(), 10)
	limiter, avgFreq, normFreq, tleFreq, allow := spy(id, c.ClientIP(), now, method)
	fmt.Printf("%f %d %d", avgFreq, normFreq, tleFreq)
	if tleFreq != -1 {
		limitBan(id, limiter, tleFreq)
	} else if method == "GET" && (normFreq >= 20 || avgFreq >= 6) && tleFreq == -1 {
		slideBar(c)
	} else if method == "POST" && (normFreq >= 8 || avgFreq >= 3) && tleFreq == -1 {
		slideBar(c)
	}
	return allow
} // 风控实现
