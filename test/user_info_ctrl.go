package main

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func signPhone(phone string) bool {
	db, err := gorm.Open(mysql.Open(mysqlInfo), &gorm.Config{})
	if err != nil {
		fmt.Println(err)
	}
	if phone == "" {
		return false
	}
	var user User
	db.Where("phone = ?", phone).First(&user)
	if user.Phone != phone {
		return false //手机号未找到指定用户
	}
	return true
} //手机号登录信息比对

func signUsername(username string, password string) bool {
	db, err := gorm.Open(mysql.Open(mysqlInfo), &gorm.Config{})
	if err != nil {
		fmt.Println(err)
	}
	if username == "" || password == "" {
		return false
	}
	var user User
	db.Where("username = ?", username).First(&user)
	if user.Username != username || decode(user.Password) != password {
		return false //用户名与密码不匹配
	}

	return true
} //用户名登录信息比对

func save(username string, phone string, password string) {
	db, err := gorm.Open(mysql.Open(mysqlInfo), &gorm.Config{})
	if err != nil {
		fmt.Println(err)
	}
	var user User
	user.Username = username
	user.Password = encode(password)
	user.Phone = phone
	fmt.Println(user)
	db.Create(&user)
} //存储新用户信息

func delete(username string) {
	db, err := gorm.Open(mysql.Open(mysqlInfo), &gorm.Config{})
	if err != nil {
		fmt.Println(err)
	}
	var user User
	db.Where("username = ?", username).First(&user)
	db.Delete(&user)
} //删除用户
