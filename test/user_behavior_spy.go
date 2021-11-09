package main

import (
	"github.com/gomodule/redigo/redis"
	"golang.org/x/time/rate"
	"sync"
)

//RateLimiter
type Limiter struct {
	ids map[string]*rate.Limiter
	cnt redis.Conn
	err error
	mu  *sync.RWMutex
	r   rate.Limit
	b   int
	normFreqChan chan int
	init bool
}

var rc Limiter
func InitLimiter(r rate.Limit, b int){
	rc.normFreqChan,rc.mu = make(chan int,100), &sync.RWMutex{}
	rc.ids, rc.init = make(map[string]*rate.Limiter), true
	rc.cnt,rc.err = redis.Dial("tcp","127.0.0.1:6379")
	rc.r, rc.b = r, b
	if rc.err != nil {
		panic(rc.err)
	}
	// defer rc.cnt.Close()
}


//AddID创建了一个新的速率限制器，并将其添加到ids映射中，
//使用deviceid作为主键
func (l *Limiter) AddID(id string, ip string, now string) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()
	limiter := rate.NewLimiter(l.r, l.b)
	// l.ids[id] = limiter
	_, err := rc.cnt.Do("HSET",id,"ip",ip,"time",now,"normFreq",0,"tleFreq",0)
	_, err = rc.cnt.Do("EXPIRE", id, 3600)
	rc.normFreqChan<-0
	if err != nil {
		return nil
	}//插入记录，设置过期时间为1个小时，norm_freq表示一小时内访问次数，tle_freq表示一小时内被拒绝的访问次数
	return limiter
}

//GetLimiter返回所提供的IP地址的速率限制器
//否则AddID将id添加到映射中
func (l *Limiter) GetLimiter(id string, ip string, now string) *rate.Limiter {
	l.mu.Lock()
	limiter, exists := l.ids[id]
	if !exists {
		l.mu.Unlock()
		return l.AddID(id, ip, now)
	}else {
		normFreq, err := rc.cnt.Do("HINCRBY",id, "normFreq",1)
		_, err = rc.cnt.Do("EXPIRE", id, 3600)
		if err != nil {
			return nil
		}//普通访问计数器加一
		rc.normFreqChan<-normFreq.(int)
		l.mu.Unlock()
		return limiter
	}
}

func spy(id string, ip string, now string) (int,int){
	if !rc.init {
		InitLimiter(1, 5)
	}
	if id=="" {//如果没有获取到ID则使用ip代替
		id=ip
	}
	limiter := rc.GetLimiter(id,ip,now)
	normFreq := <-rc.normFreqChan
	if limiter.Allow(){
		return normFreq,-1
	}else {
		tleFreq, err := rc.cnt.Do("HINCRBY", id, "tleFreq", 1)
		if err != nil {
			return -1, -1
		}
		return normFreq,tleFreq.(int)
	}
}//监视每个用户的操作频次
