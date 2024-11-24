package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/bklv-kirill/go-site-form-checker/pkg/config"
	"net/http"
	"time"
)

const (
	MethodSendMessage string = "sendMessage"
)

type Telegram struct {
	ChatId    string `json:"chat_id"`
	Token     string `json:"token"`
	ParseMode string `json:"parse_mode"`
}

func New(cfg *config.Cfg) *Telegram {
	return &Telegram{
		ChatId:    cfg.TelegramChatId,
		Token:     cfg.TelegramToken,
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

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("https://api.telegram.org/bot%s/%s", t.Token, MethodSendMessage), bytes.NewBuffer(jsonData))
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
		return fmt.Errorf(resp.Status)
	}

	return nil
}
