package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
	"strconv"
	"time"
)

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

/*
*  风控逻辑实现
* isAllow :  返回是否通过风控
* spyMethod: 处理结果-为 0 即时通过， 为 1 既是 limitBan， 为 2 既是 需要slideBar
 */
func riskCtrl(c *gin.Context) (isAllow bool, spyMethod int) {
	spyMethod = 0
	id, _ := c.Cookie("DeviceID")
	method, _ := c.Request.Method, c.ClientIP()
	now := strconv.FormatInt(time.Now().UnixNano(), 10)
	limiter, avgFreq, normFreq, tleFreq, allow := spy(id, c.ClientIP(), now, method)
	fmt.Printf("%f %d %d", avgFreq, normFreq, tleFreq)
	if tleFreq != -1 {
		limitBan(id, limiter, tleFreq)
		spyMethod = 1
	} else if method == "GET" && (normFreq >= 20 || avgFreq >= 6) && tleFreq == -1 {
		spyMethod = 2
	} else if method == "POST" && (normFreq >= 8 || avgFreq >= 3) && tleFreq == -1 {
		spyMethod = 2
	}
	return allow, spyMethod
} // 风控实现
