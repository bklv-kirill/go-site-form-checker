package sfc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/bklv-kirill/go-site-form-checker/pkg/config"
	form "github.com/bklv-kirill/go-site-form-checker/pkg/models"
	"github.com/bklv-kirill/go-site-form-checker/pkg/telegram"
	"github.com/chromedp/chromedp"
	"github.com/google/uuid"
	"log"
	"net/http"
	"sync"
	"time"
)

var cfg *config.Config = config.New()

func Check(f *form.Form, wg *sync.WaitGroup, ch <-chan struct{}) {
	defer func() {
		wg.Done()
		<-ch
	}()

	var tg *telegram.Telegram = telegram.New(cfg)

	var genMsg string = fmt.Sprintf("Название: %s | Ссылка: %s\n", f.Name, f.Url)

	ctx, cancel := chromedp.NewRemoteAllocator(context.Background(), "http://127.0.0.1:9222")
	defer cancel()

	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 3*60*time.Second)
	defer cancel()

	if err := goToSite(&ctx, f.Url); err != nil {
		if err := tg.SendMessage(fmt.Sprintf("%s | %s", err.Error(), genMsg)); err != nil {
			log.Printf("%v+\n", err)
		}

		return
	}

	if err := chromedpWait(&ctx, 5); err != nil {
		if err := tg.SendMessage(fmt.Sprintf("%s | %s", err.Error(), genMsg)); err != nil {
			log.Printf("%v+\n", err)
		}

		return
	}

	if err := waitElemForClick(&ctx, f.ElemForClick, 30); err != nil {
		if err := tg.SendMessage(fmt.Sprintf("%s | %s", err.Error(), genMsg)); err != nil {
			log.Printf("%v+\n", err)
		}

		return
	}

	if err := evaluateClickOnElem(&ctx, f.ElemForClick); err != nil {
		if err := tg.SendMessage(fmt.Sprintf("%s | %s", err.Error(), genMsg)); err != nil {
			log.Printf("%v+\n", err)
		}

		return
	}

	if err := waitExpElem(&ctx, f.ExpElem, 30); err != nil {
		if err := tg.SendMessage(fmt.Sprintf("%s | %s", err.Error(), genMsg)); err != nil {
			log.Printf("%v+\n", err)
		}

		return
	}

	leadUuid, err := fillForm(&ctx, f)
	if err != nil {
		if err := tg.SendMessage(fmt.Sprintf("%s | %s", err.Error(), genMsg)); err != nil {
			log.Printf("%v+\n", err)
		}

		return
	}

	if err := sendForm(&ctx, f); err != nil {
		if err := tg.SendMessage(fmt.Sprintf("%s | %s", err.Error(), genMsg)); err != nil {
			log.Printf("%v+\n", err)
		}

		return
	}

	if err := waitResElem(&ctx, f.ResElem, 30); err != nil {
		if err := tg.SendMessage(fmt.Sprintf("%s | %s", err.Error(), genMsg)); err != nil {
			log.Printf("%v+\n", err)
		}

		return
	}

	if leadUuid != "" {
		if err := checkInCRM(leadUuid); err != nil {
			if err := tg.SendMessage(fmt.Sprintf("%s | %s", err.Error(), genMsg)); err != nil {
				log.Printf("%v+\n", err)
			}

			return
		}
	}
}

func goToSite(ctx *context.Context, url string) error {
	if cfg.DebugMode {
		log.Printf("Переход на сайт %s\n", url)
	}
	if err := chromedp.Run(*ctx, chromedp.Navigate(url)); err != nil {
		return fmt.Errorf("Ошибка при переходе на сайт: %v", err)
	}

	return nil
}

func chromedpWait(ctx *context.Context, sec int) error {
	if cfg.DebugMode {
		log.Printf("Ожидание: %d секунд\n", sec)
	}
	if err := chromedp.Run(*ctx, chromedp.Sleep(time.Duration(sec)*time.Second)); err != nil {
		return fmt.Errorf("Ошибка при ожидании: %v", err)
	}

	return nil
}

func waitElemForClick(ctx *context.Context, elemForClick string, sec int) error {
	if cfg.DebugMode {
		log.Printf("Ожидание появления элемента для нажатия | %d секунд\n", sec)
	}
	wCtx, wCancel := context.WithTimeout(*ctx, time.Duration(sec)*time.Second)
	defer wCancel()
	if err := chromedp.Run(wCtx, chromedp.WaitVisible(elemForClick, chromedp.ByQuery)); err != nil {
		return fmt.Errorf("Ошибка при ожидании элемента для нажатия: %v", err)
	}

	return nil
}

func evaluateClickOnElem(ctx *context.Context, elemForClick string) error {
	if cfg.DebugMode {
		log.Printf("Имитация клика на элемент для нажатия\n")
	}
	if err := chromedp.Run(*ctx, chromedp.Evaluate(fmt.Sprintf("document.querySelector('%s').click()", elemForClick), nil)); err != nil {
		return fmt.Errorf("Ошибка при клике на элемент для нажатия: %v", err)
	}

	return nil
}

func waitExpElem(ctx *context.Context, expElem string, sec int) error {
	if cfg.DebugMode {
		log.Printf("Ожидание появления элемента для взаимодействия | %d секунд\n", sec)
	}
	wCtx, wCancel := context.WithTimeout(*ctx, time.Duration(sec)*time.Second)
	defer wCancel()
	if err := chromedp.Run(wCtx, chromedp.WaitVisible(expElem, chromedp.ByQuery)); err != nil {
		return fmt.Errorf("Ошибка при ожидании элемента для взаимодействия: %v", err)
	}

	return nil
}

func fillForm(ctx *context.Context, f *form.Form) (string, error) {
	if cfg.DebugMode {
		log.Printf("Заполнение формы\n")
	}
	var leadUuid string
	for _, i := range f.Inputs {
		if i.ForUuid {
			leadUuid = fmt.Sprintf("SFC - %s", uuid.New())
			i.Value = leadUuid
		}

		if err := chromedp.Run(*ctx, chromedp.SendKeys(fmt.Sprintf("%s %s", f.ExpElem, i.Selector), i.Value)); err != nil {
			return leadUuid, fmt.Errorf("Ошибка при заполнении формы: %v", err)
		}
	}

	return leadUuid, nil
}

func sendForm(ctx *context.Context, f *form.Form) error {
	if cfg.DebugMode {
		log.Printf("Отправка формы\n")
	}
	if err := chromedp.Run(*ctx, chromedp.Evaluate(fmt.Sprintf("document.querySelector('%s %s').click()", f.ExpElem, f.SubmitElem), nil)); err != nil {
		return fmt.Errorf("Ошибка при отправке формы: %v", err)
	}

	return nil
}

func waitResElem(ctx *context.Context, resElem string, sec int) error {
	if cfg.DebugMode {
		log.Printf("Ожидание появления результирующего элемента | %d секунд\n", sec)
	}
	wCtx, wCancel := context.WithTimeout(*ctx, time.Duration(sec)*time.Second)
	defer wCancel()
	if err := chromedp.Run(wCtx, chromedp.WaitVisible(resElem, chromedp.ByQuery)); err != nil {
		return fmt.Errorf("Ошибка при ожидании появления результирующего элемента: %v", err)
	}

	return nil
}

func checkInCRM(leadUuid string) error {
	if cfg.DebugMode {
		log.Printf("Проверка лида в CRM\n")
	}

	var data map[string]string = map[string]string{
		"uuid": leadUuid,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, cfg.CrmUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "SiteFormChecker")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cfg.CrmToken))

	var client http.Client = http.Client{
		Timeout: 10 * time.Second,
	}

	var att int = 0

	var clsr func() error
	clsr = func() error {
		time.Sleep(10 * time.Second)

		att++

		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			if cfg.DebugMode {
				log.Printf("Ошибка при проверке лида в CRM | Попытка %d из %d\n", att, cfg.CrmAttempts)
			}

			if att >= cfg.CrmAttempts {
				return fmt.Errorf(resp.Status)
			}

			return clsr()
		}

		return nil
	}

	return clsr()
}
