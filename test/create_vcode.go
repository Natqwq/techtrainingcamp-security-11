package main

import (
	"fmt"
	"math/rand"
	"time"
)
//用于前端调用获取验证码，生成6位随机验证码，但前端如何调用及如果传给前端还未知
func create()string{
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	vcode := fmt.Sprintf("%06v", rnd.Int31n(1000000))
	return vcode
}
