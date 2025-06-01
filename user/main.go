package main

import (
	"context"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"shop/user/model/gen"
	"time"
)

func main() {

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
