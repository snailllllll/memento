package utils

import (
	"os"

	"github.com/joho/godotenv"
)

// 全局配置结构体
var Config struct {
	Token       string
	AdminUIN    string
	Text        string
	Addr        string
	DBURI       string
	DBUsername  string
	DBPassword  string
	LLMAPIKey   string
	Port        string
	InformGroup string
	APIHost     string
}

// LoadConfig 加载配置并初始化全局Config
func LoadConfig() error {
	_ = godotenv.Load() // 加载.env文件

	// 初始化全局配置
	Config.Token = os.Getenv("TOKEN")
	Config.AdminUIN = os.Getenv("ADMIN_UIN")
	Config.Text = os.Getenv("TEXT")
	Config.Addr = os.Getenv("ADDR")
	Config.DBURI = os.Getenv("DB_URI")
	Config.DBUsername = os.Getenv("DB_USERNAME")
	Config.DBPassword = os.Getenv("DB_PASSWORD")
	Config.LLMAPIKey = os.Getenv("LLM_APIKEY")
	Config.Port = os.Getenv("PORT")
	Config.InformGroup = os.Getenv("INFORM_GROUP")
	Config.APIHost = os.Getenv("API_HOST")

	return nil
}

// GetConfig 获取配置值（兼容旧版）
func GetConfig(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
