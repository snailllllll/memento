package utils

import (
	"math/rand"
	"time"
)

// GenerateRandomNumber 生成指定范围内的随机整数
// min: 最小值（包含）
// max: 最大值（包含）
// 返回值: [min, max]范围内的随机整数
func GenerateRandomNumber(min, max int) int {
	// 确保随机数种子只初始化一次
	rand.Seed(time.Now().UnixNano())

	// 处理边界情况
	if min >= max {
		return min
	}

	// 生成随机数
	return rand.Intn(max-min+1) + min
}

// GenerateRandomString 生成指定长度的随机字符串
// length: 字符串长度
// 返回值: 随机字符串
func GenerateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)

	rand.Seed(time.Now().UnixNano())
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}

	return string(result)
}

// GenerateRandomDigits 生成指定长度的随机数字字符串
// length: 数字字符串长度
// 返回值: 随机数字字符串
func GenerateRandomDigits(length int) string {
	const digits = "0123456789"
	result := make([]byte, length)

	rand.Seed(time.Now().UnixNano())
	for i := range result {
		result[i] = digits[rand.Intn(len(digits))]
	}

	return string(result)
}
