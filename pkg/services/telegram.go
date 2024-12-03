package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/bklv-kirill/go-site-form-checker/pkg/config"
	"net/http"
	"time"
)

const BotUrl string = "https://api.telegram.org/bot"

const (
	MethodSendMessage string = "sendMessage"
)

type Telegram struct {
	ChatId    string `json:"chat_id"`
	BotToken  string `json:"bot_token"`
	ParseMode string `json:"parse_mode"`
}

func NewTelegram(cfg *config.Config) *Telegram {
	if cfg.TelegramChatId == "" || cfg.TelegramBotToken == "" {
		return nil
	}

	return &Telegram{
		ChatId:    cfg.TelegramChatId,
		BotToken:  cfg.TelegramBotToken,
		ParseMode: cfg.TelegramParseMode,
	}
}

func (t *Telegram) SendMessage(text string) error {
	var data map[string]string = map[string]string{
		"chat_id":                  t.ChatId,
		"parse_mode":               t.ParseMode,
		"text":                     text,
		"disable_web_page_preview": "true",
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s/%s", BotUrl, t.BotToken, MethodSendMessage), bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	var client http.Client = http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("При отправке сообщения в телеграм произошла ошибка: %s | Сообщение: %s", resp.Status, text)
	}

	return nil
}
