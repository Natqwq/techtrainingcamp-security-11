package main

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"math/rand"
	"strconv"
	"time"
)

func checkVcode(v Post) bool {
	phone := v.Phone
	var vcheck CheckVcode
	db, err := gorm.Open(mysql.Open(mysqlInfo), &gorm.Config{})
	if err != nil {
		fmt.Println(err)
	}

	//获取此用户最近的一条验证码
	db.Where("phone = ?", phone).Order("create_at DESC").Limit(1).First(&vcheck)
	if &vcheck == nil {
		return false
	} else if vcheck.HasExpired {
		return false
	} else if vcheck.Vcode != v.Vcode || phone != vcheck.Phone {
		return false
	}
	int64, _ := strconv.ParseInt(vcheck.Create, 10, 64)
	now := time.Now().Unix()
	//10分钟以内均有效
	if now-int64 > vcodeValid {
		return false
	}
	vcheck.HasExpired = true
	db.Model(&vcheck).Update("has_expired", true)
	return true
}

/*
*	保存手机_验证码实体类
 */
func saveVcode(phone string, code string) bool {
	db, err := gorm.Open(mysql.Open(mysqlInfo), &gorm.Config{})
	if err != nil {
		fmt.Println(err)
	}
	var checkVcode CheckVcode
	checkVcode.Vcode = code
	checkVcode.Phone = phone
	checkVcode.Create = strconv.FormatInt(time.Now().Unix(), 10)
	checkVcode.CreateAt = time.Now()
	fmt.Println(checkVcode)
	db.Create(&checkVcode)
	return true
}

//用于前端调用获取验证码，生成6位随机验证码，但前端如何调用及如果传给前端还未知
func create() string {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	vcode := fmt.Sprintf("%06v", rnd.Int31n(1000000))
	return vcode
}

/*
* 根据用户手机号获取用户用户近一个小时的获取验证码个数
 */
func vCodeSpy(post CheckVcode) int {
	db, err := gorm.Open(mysql.Open(mysqlInfo), &gorm.Config{})
	if err != nil {
		fmt.Println(err)
	}
	vcodeRecordNum := int64(0)
	db.Model(&post).Where("create_at >  date_sub(now(), interval 1 hour) and phone = ?", post.Phone).Count(&vcodeRecordNum)
	if vcodeRecordNum > 5 {
		return 1
	} else if vcodeRecordNum > 2 {
		return 2
	}
	return 0
}
