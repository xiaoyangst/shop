package main

import (
	"context"
	"crypto/sha512"
	"database/sql"
	"fmt"
	"github.com/anaskhan96/go-password-encoder"
	_ "github.com/go-sql-driver/mysql"
	"shop/user/model/gen"
	"strings"
	"time"
)

// 定义密码选项
var options = &password.Options{
	SaltLen:      16,         // 盐的长度
	Iterations:   100,        // 迭代次数
	KeyLen:       32,         // 生成的密钥长度
	HashFunction: sha512.New, // 使用 SHA-512 哈希函数
}

// 生成密码
func genPwd(pwd string) string {
	salt, encodedPwd := password.Encode(pwd, options)
	return fmt.Sprintf("pbkdf2_sha512$%s$%s", salt, encodedPwd) // 加密类型$盐$加密后的密码
}

// 验证密码
func verifyPwd(encodePwd string, srcPwd string) bool {
	pwdInfo := strings.Split(encodePwd, "$")
	fmt.Println(pwdInfo[1])
	fmt.Println(pwdInfo[2])
	check := password.Verify(srcPwd, pwdInfo[1], pwdInfo[2], options)
	return check
}

func main() {

	// 测试密码生成和验证
	pwd := "123456"
	actualPwd := genPwd(pwd)
	fmt.Println("生成的密码:", actualPwd)
	// 验证密码
	if verifyPwd(actualPwd, pwd) {
		fmt.Println("密码验证通过")
	} else {
		fmt.Println("密码验证失败")
	}

	// 连接 MySQL 数据库
	dsn := "root:root@tcp(192.168.1.102:3306)/shop?parseTime=true"
	dbConn, _ := sql.Open("mysql", dsn)
	defer dbConn.Close()

	// 创建 Queries 实例
	queries := model.New(dbConn)

	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	// 插入数据
	err := queries.CreateUser(ctx, model.CreateUserParams{
		Mobile:   "1301234568",
		Password: "123456",
		NikeName: "xy",
		Birthday: sql.NullTime{
			Time:  time.Now(),
			Valid: false,
		},
		Gender: model.UsersGenderOther,
		Role:   "user",
	})

	if err != nil {
		println(err.Error())
		return
	}

	// 执行查询
	users, err := queries.ListUsers(ctx)
	if err != nil {
		println(err.Error())
		return
	}
	// 打印结果
	for _, user := range users {
		println(user.Nikename)
	}

}
