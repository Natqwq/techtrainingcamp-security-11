package main

import (
	"github.com/gomodule/redigo/redis"
	"golang.org/x/time/rate"
	"strconv"
	"sync"
	"time"
)

// Limiter RateLimiter
type Limiter struct {
	idsGet       map[string]*rate.Limiter
	idsPost      map[string]*rate.Limiter
	cnt          redis.Conn
	err          error
	mu           *sync.RWMutex
	r            rate.Limit
	b            int
	normFreqChan chan int
	fmt          string
	init         bool
}

func B2S(bs []uint8) string {
	var ba []byte
	for _, b := range bs {
		ba = append(ba, byte(b))
	}
	return string(ba)
} //byte数组转为String

var rc Limiter

func InitLimiter(r rate.Limit, b int) {
	rc.normFreqChan, rc.mu = make(chan int, chanBuf), &sync.RWMutex{}
	rc.idsGet, rc.idsPost, rc.init = make(map[string]*rate.Limiter), make(map[string]*rate.Limiter), true
	rc.cnt, rc.err = redis.Dial("tcp", redisIP)
	_, rc.err = rc.cnt.Do("AUTH", redisUser, redisPsd)
	_, rc.err = rc.cnt.Do("FLUSHALL")
	rc.fmt = time.RFC3339Nano
	rc.r, rc.b = r, b
	if rc.err != nil {
		panic(rc.err)
	}
	// defer rc.cnt.Close()
}

// AddID AddID创建了一个新的速率限制器，并将其添加到ips映射中，
//使用deviceid作为密钥
func (l *Limiter) AddID(id string, ip string, now string) *rate.Limiter {
	l.mu.Lock()
	limiter := rate.NewLimiter(l.r, l.b)
	// l.ids[id] = limiter
	_, l.err = l.cnt.Do("HSET", id, "ip", ip, "bgTime", now, "normFreq", 1, "tleFreq", 0)
	_, l.err = l.cnt.Do("EXPIRE", id, redisExpire)
	l.normFreqChan <- 1
	//插入记录，设置过期时间为1分钟，norm_freq表示一分钟内访问次数，tle_freq表示一分钟内被拒绝的访问次数
	l.mu.Unlock()
	return limiter
}

// GetLimiter GetLimiter返回所提供的IP地址的速率限制器.否则AddIP将地址添加到映射中
func (l *Limiter) GetLimiter(id string, ip string, now string, method string) *rate.Limiter {
	l.mu.Lock()
	var limiter *rate.Limiter
	exist, _ := l.cnt.Do("EXISTS", id)

	//根据键存在与否，获取或者生成相应的限制器
	if (int)(exist.(int64)) == 0 {
		l.mu.Unlock()
		if method == "GET" {
			l.idsGet[id] = l.AddID(id, ip, now)
			limiter, _ = l.idsGet[id]
		} else {
			l.idsPost[id] = l.AddID(id, ip, now)
			limiter, _ = l.idsPost[id]
		}
	} else {
		//获取请求数， 且生命周期重置为redisExpires
		normFreq, _ := l.cnt.Do("HINCRBY", id, "normFreq", 1)
		_, l.err = l.cnt.Do("EXPIRE", id, redisExpire)

		//根据访问方法，来分配相应的限制器
		if method == "GET" {
			limiter = l.idsGet[id]
		} else {
			limiter = l.idsPost[id]
		}
		l.normFreqChan <- int(normFreq.(int64))
		l.mu.Unlock()
	}
	return limiter
}

func spy(id string, ip string, now string, method string) (*rate.Limiter, float64, int, int, bool) {
	if !rc.init {
		InitLimiter(rLimiter, bLimiter)
	}
	if id == "" {
		id = ip
	}
	if method == "GET" {
		_, rc.err = rc.cnt.Do("SELECT", 0)
	} else {
		_, rc.err = rc.cnt.Do("SELECT", 1)
	}
	limiter := rc.GetLimiter(id, ip, now, method)
	normFreq := <-rc.normFreqChan

	//获得该该设备的第一次请求时间
	bgTime, _ := rc.cnt.Do("HGET", id, "bgTime")
	bgTime = B2S(bgTime.([]uint8))
	bgTime_, _ := strconv.ParseInt(bgTime.(string), 10, 64)

	//现在时间
	nowTime, _ := strconv.ParseInt(now, 10, 64)

	//平均速率
	avgFreq := 1.0
	//防止第一次访问时除以0
	if nowTime != bgTime_ {
		avgFreq = float64(normFreq) / (float64(nowTime-bgTime_) / 1e9)
	}
	if limiter.Allow() {
		tleFreq, _ := rc.cnt.Do("HINCRBY", id, "tleFreq", 0)
		return limiter, avgFreq, normFreq, int(tleFreq.(int64)), true
	} else {
		tleFreq, _ := rc.cnt.Do("HINCRBY", id, "tleFreq", 1)
		return limiter, avgFreq, normFreq, int(tleFreq.(int64)), false
	}
} //监视每个设备的的Get和Post操作频次
