package main

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"strconv"
	"time"
)

func checkVcode(v Post) bool {
	phone := v.Phone
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Println(err)
	}
	var vcheck CheckVcode
	//TODO；这里应该是根据手机号进行查找
	db.Where("vcode = ?", v.Vcode).First(&vcheck)

	if vcheck.Vcode != v.Vcode || phone != vcheck.Phone {
		return false
	}
	int64, _ := strconv.ParseInt(vcheck.Create, 10, 64)
	now := time.Now().Unix()
	//10分钟以内均有效
	if now-int64 > 60*10 {
		return false
	}
	return true
}

/*
*	保存手机_验证码实体类
 */
func saveVcode(phone string, code string) bool {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Println(err)
	}
	var checkVcode CheckVcode
	checkVcode.Vcode = code
	checkVcode.Phone = phone
	checkVcode.Create = strconv.FormatInt(time.Now().Unix(), 10)
	checkVcode.Create_at = time.Now()
	fmt.Println(checkVcode)
	db.Create(&checkVcode)
	return true
}
