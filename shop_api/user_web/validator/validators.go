package validator

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

// ValidateMobile 自定义验证器，用于验证手机号码是否合法
func ValidateMobile(fl validator.FieldLevel) bool {
	mobile := fl.Field().String()
	//使用正则表达式判断是否合法
	ok, _ := regexp.MatchString(`^1([38][0-9]|14[579]|5[^4]|16[6]|7[1-35-8]|9[189])\d{8}$`, mobile)
	if !ok {
		return false
	}
	return true
}
