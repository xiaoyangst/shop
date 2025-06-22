package global

import (
	"crypto/sha512"
	"database/sql"
	"fmt"
	"github.com/anaskhan96/go-password-encoder"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"strings"
)

var (
	DbConn *sql.DB
)

func connectMysql() {
	dsn := "root:root@tcp(192.168.0.101:3306)/shop?parseTime=true"

	var err error
	DbConn, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("failed to open DB: %v", err)
	}

	if err = DbConn.Ping(); err != nil {
		log.Fatalf("failed to ping DB: %v", err)
	}
}

// 定义密码选项
var options = &password.Options{
	SaltLen:      16,         // 盐的长度
	Iterations:   100,        // 迭代次数
	KeyLen:       32,         // 生成的密钥长度
	HashFunction: sha512.New, // 使用 SHA-512 哈希函数
}

// GenPwd 生成密码
func GenPwd(pwd string) string {
	salt, encodedPwd := password.Encode(pwd, options)
	return fmt.Sprintf("pbkdf2_sha512$%s$%s", salt, encodedPwd) // 加密类型$盐$加密后的密码
}

// VerifyPwd 验证密码
func VerifyPwd(srcPwd string, encodePwd string) bool {
	pwdInfo := strings.Split(encodePwd, "$")
	if len(pwdInfo) < 3 {
		return false
	}
	check := password.Verify(srcPwd, pwdInfo[1], pwdInfo[2], options)
	return check
}

func init() {
	connectMysql()
}
