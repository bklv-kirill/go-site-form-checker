package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Файл .env не найден: %s", err)
	}
}

type Config struct {
	DbCon  string
	DbHost string
	DbPort string
	DbUser string
	DbPass string
	DbName string

	CrmUrl      string
	CrmToken    string
	CrmAttempts int

	TelegramChatId    string
	TelegramBotToken  string
	TelegramParseMode string

	DebugMode     bool
	MaxGoroutines int
}

func New() *Config {
	return &Config{
		DbCon:  getEnvAsString("DB_CON", "mysql"),
		DbHost: getEnvAsString("DB_HOST", "127.0.0.1"),
		DbPort: getEnvAsString("DB_PORT", "3306"),
		DbUser: getEnvAsString("DB_USER", "root"),
		DbPass: getEnvAsString("DB_PASS", ""),
		DbName: getEnvAsString("DB_NAME", "site_form_checker"),

		CrmUrl:      getEnvAsString("CRM_URL", ""),
		CrmToken:    getEnvAsString("CRM_TOKEN", ""),
		CrmAttempts: getEnvAsInt("CRM_ATTEMPTS", 10),

		TelegramChatId:    getEnvAsString("TELEGRAM_CHAT_ID", ""),
		TelegramBotToken:  getEnvAsString("TELEGRAM_BOT_TOKEN", ""),
		TelegramParseMode: getEnvAsString("TELEGRAM_PARSE_MODE", "html"),

		DebugMode:     getEnvAsBool("DEBUG_MODE", true),
		MaxGoroutines: getEnvAsInt("MAX_GOROUTINES", 10),
	}
}

func getEnvAsString(key string, def string) string {
	if str, exists := os.LookupEnv(key); exists {
		return str
	}

	return def
}

func getEnvAsInt(key string, def int) int {
	var str string = getEnvAsString(key, "")

	if val, err := strconv.Atoi(str); err == nil {
		return val
	}

	return def
}

func getEnvAsBool(key string, def bool) bool {
	var str string = getEnvAsString(key, "")

	if val, err := strconv.ParseBool(str); err == nil {
		return val
	}

	return def
}
