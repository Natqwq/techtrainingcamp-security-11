/*
 * @Author: 1zwm
 * @Date: 2021-11-06 13:23:14
 * @LastEditTime: 2021-11-06 14:39:47
 * @LastEditors: Please set LastEditors
 * @Description: 基于ip实现限制http频率请求
 * 	method :
	var limiter = NewIPRateLimiter(1, 5)

	limiter := limiter.GetLimiter(r.RemoteAddr)
  	if !limiter.Allow() {    //令牌池满了返回false，若未满，则小号一个令牌，将请求传给下一个程序
   		return
  	}

 * @FilePath: /techtrainingcamp-security-11/test/ip_rate_limit.go
 */

 package main

import (
	"sync"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

//IPRateLimiter
type IPRateLimiter struct {
	ips map[string]*rate.Limiter
	mu  *sync.RWMutex
	r   rate.Limit
	b   int
}

//New IPRatelimiter
func NewIPLimiter(r rate.Limit, b int) *IPRateLimiter {
	i := &IPRateLimiter{
		ips: make(map[string]*(rate.Limiter)),
		mu:  &sync.RWMutex{},
		r:   r,
		b:   b,
	}
	return i
}

//AddIP创建了一个新的速率限制器，并将其添加到ips映射中，
//使用ip地址作为密钥
func (i *IPRateLimiter) AddIP(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter := rate.NewLimiter(i.r, i.b)

	i.ips[ip] = limiter

	return limiter
}

//GetLimiter返回所提供的IP地址的速率限制器
//否则AddIP将地址添加到映射中
func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.Lock()
	limiter, exists := i.ips[ip]

	if !exists {
		i.mu.Unlock()
		return i.AddIP(ip)
	}

	i.mu.Unlock()

	return limiter
}
