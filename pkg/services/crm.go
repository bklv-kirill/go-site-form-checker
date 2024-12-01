package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/bklv-kirill/go-site-form-checker/pkg/config"
	"log"
	"net/http"
	"time"
)

type Crm struct {
	DebugMode  bool
	Url        string
	Token      string
	Attempts   int
	RetryDelay int
}

func NewCrm(cfg *config.Config) *Crm {
	return &Crm{
		DebugMode:  cfg.DebugMode,
		Url:        cfg.CrmUrl,
		Token:      cfg.CrmToken,
		Attempts:   cfg.CrmAttempts,
		RetryDelay: cfg.CrmRetryDelay,
	}
}

func (crm *Crm) CheckLeadByUuid(leadUuid string) error {
	var data map[string]string = map[string]string{
		"uuid": leadUuid,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, crm.Url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "SiteFormChecker")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", crm.Token))

	var client http.Client = http.Client{
		Timeout: 180 * time.Second,
	}

	for try := 1; try <= crm.Attempts; try++ {
		resp, err := client.Do(req)
		if err != nil {
			return err
		}

		if resp.StatusCode == http.StatusOK {
			return nil
		} else if try == crm.Attempts {
			return fmt.Errorf(resp.Status)
		} else if crm.DebugMode {
			log.Printf("Ошибка при проверке лида в CRM | uuid: %s | Повторная попытка...", leadUuid)
		}

		time.Sleep(time.Duration(crm.RetryDelay) * time.Second)
	}

	return fmt.Errorf("При проверке лида в CRM произошла неизвестная ошибка | uuid: %s", leadUuid)
}
