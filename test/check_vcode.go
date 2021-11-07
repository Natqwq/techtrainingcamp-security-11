package main

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"strconv"
	"time"
)

func checkVcode(v Post) bool{
	phone:=v.Phone
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Println(err)
	}
	var vcheck CheckVcode
	db.Where("vcode = ?",v.Vcode).First(&vcheck)

	if vcheck.Vcode!=v.Vcode ||phone!= vcheck.Phone  {
		return false
	}
	int64,_:=strconv.ParseInt(vcheck.Create,10,64)
	now:=time.Now().Unix()
	if now-int64>60 {
		return false
	}
	return true
}
