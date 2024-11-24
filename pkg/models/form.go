package form

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/bklv-kirill/go-site-form-checker/pkg/config"
	"github.com/bklv-kirill/go-site-form-checker/pkg/telegram"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"github.com/google/uuid"
	"log"
	"net/http"
	"sync"
	"time"
)

type Form struct {
	Id           int     `db:"id" json:"id"`
	Name         string  `db:"name" json:"name"`
	Url          string  `db:"url" json:"url"`
	ElemForClick string  `db:"element_for_click" json:"element_for_click"`
	ExpElem      string  `db:"expected_element" json:"expected_element"`
	SubmitElem   string  `db:"submit_element" json:"submit_element"`
	ResElem      string  `db:"result_element" json:"result_element"`
	Inputs       []Input `db:"inputs" json:"inputs"`
	CreatedAt    string  `db:"created_at" json:"created_at"`
	UpdatedAt    string  `db:"updated_at" json:"updated_at"`
}

type Input struct {
	Id        int    `db:"id" json:"id"`
	FormId    int    `db:"form_id" json:"form_id"`
	Selector  string `db:"selector" json:"selector"`
	Value     string `db:"value" json:"value"`
	Uuid      bool   `db:"uuid" json:"uuid"`
	CreatedAt string `db:"created_at" json:"created_at"`
	UpdatedAt string `db:"updated_at" json:"updated_at"`
}

func (f *Form) Check(wg *sync.WaitGroup, ch chan struct{}, cfg *config.Cfg) {
	defer func() {
		wg.Done()
		<-ch
	}()

	var tg *telegram.Telegram = telegram.New(cfg)

	var genMsg string = fmt.Sprintf("Название: %s | Ссылка: %s\n", f.Name, f.Url)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	ctx, cancel = chromedp.NewRemoteAllocator(ctx, "http://127.0.0.1:9222")
	defer cancel()

	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	if cfg.DebugMode {
		chromedp.ListenTarget(ctx, func(ev interface{}) {
			if msg, err := ev.(*runtime.EventConsoleAPICalled); err {
				for _, arg := range msg.Args {
					if arg.Value != nil {
						log.Printf("Сonsole [%s]: %s\n", msg.Type, arg.Value)
					} else if arg.Preview != nil && arg.Preview.Description != "" {
						log.Printf("Сonsole [%s]: %s\n", msg.Type, arg.Preview.Description)
					}
				}
			}
		})
	}

	if cfg.DebugMode {
		log.Printf("Переход на сайт %s\n", f.Url)
	}
	if err := chromedp.Run(ctx,
		chromedp.Navigate(f.Url),
	); err != nil {
		if err = tg.SendMessage(fmt.Sprintf("Ошибка при переходе на сайт: %v | %s", err, genMsg)); err != nil {
			log.Println(err)
		}

		return
	}

	if cfg.DebugMode {
		log.Printf("Ожидание: 5 секунд")
	}
	if err := chromedp.Run(ctx,
		chromedp.Sleep(5*time.Second),
	); err != nil {
		if err = tg.SendMessage(fmt.Sprintf("Ошибка при ожидании: %v | %s", err, genMsg)); err != nil {
			log.Println(err)
		}

		return
	}

	if cfg.DebugMode {
		log.Println("Ожидание появления элемента для нажатия")
	}
	wCtx, wCancel := context.WithTimeout(ctx, 15*time.Second)
	defer wCancel()
	if err := chromedp.Run(wCtx,
		chromedp.WaitVisible(f.ElemForClick, chromedp.ByQuery),
	); err != nil {
		if err = tg.SendMessage(fmt.Sprintf("Ошибка при ожидании элемента для нажатия: %v | %s", err, genMsg)); err != nil {
			log.Println(err)
		}

		return
	}

	if cfg.DebugMode {
		log.Println("Имитация клика на элемент для нажатия")
	}
	if err := chromedp.Run(ctx,
		chromedp.Evaluate(fmt.Sprintf("document.querySelector('%s').click()", f.ElemForClick), nil),
	); err != nil {
		if err = tg.SendMessage(fmt.Sprintf("Ошибка при клике на элемент для нажатия: %v | %s", err, genMsg)); err != nil {
			log.Println(err)
		}

		return
	}

	if cfg.DebugMode {
		log.Println("Ожидание появления элемента для взаимодействия")
	}
	wCtx, wCancel = context.WithTimeout(ctx, 15*time.Second)
	defer wCancel()
	if err := chromedp.Run(wCtx,
		chromedp.WaitVisible(f.ExpElem, chromedp.ByQuery),
	); err != nil {
		if err = tg.SendMessage(fmt.Sprintf("Ошибка при ожидании элемента для взаимодействия: %v | %s", err, genMsg)); err != nil {
			log.Println(err)
		}

		return
	}

	if cfg.DebugMode {
		log.Println("Заполнение формы")
	}
	var leadUuid string
	for _, i := range f.Inputs {
		if i.Uuid {
			leadUuid = fmt.Sprintf("SFC - %s", uuid.New())
			i.Value = leadUuid
		}

		if err := chromedp.Run(ctx,
			chromedp.SendKeys(fmt.Sprintf("%s %s", f.ExpElem, i.Selector), i.Value),
		); err != nil {
			if err = tg.SendMessage(fmt.Sprintf("Ошибка при заполнении формы: %v | %s", err, genMsg)); err != nil {
				log.Println(err)
			}

			return
		}
	}

	if cfg.DebugMode {
		log.Println("Проверка формы")
		for _, i := range f.Inputs {
			var val string
			if err := chromedp.Run(ctx,
				chromedp.Value(fmt.Sprintf("%s %s", f.ExpElem, i.Selector), &val, chromedp.ByQuery),
			); err != nil {
				if err = tg.SendMessage(fmt.Sprintf("Ошибка при проверке формы: %v | %s", err, genMsg)); err != nil {
					log.Println(err)
				}

				return
			}

			log.Printf("%s: %s\n", i.Selector, val)
		}
	}

	if cfg.DebugMode {
		log.Println("Отправка формы")
	}
	if err := chromedp.Run(ctx,
		chromedp.Evaluate(fmt.Sprintf("document.querySelector('%s %s').click()", f.ExpElem, f.SubmitElem), nil),
	); err != nil {
		if err = tg.SendMessage(fmt.Sprintf("Ошибка при отправке формы: %v | %s", err, genMsg)); err != nil {
			log.Println(err)
		}

		return
	}

	if cfg.DebugMode {
		log.Println("Ожидание появления результирующего элемента")
	}
	wCtx, wCancel = context.WithTimeout(ctx, 15*time.Second)
	defer wCancel()
	if err := chromedp.Run(wCtx,
		chromedp.WaitVisible(f.ResElem, chromedp.ByQuery),
	); err != nil {
		if err = tg.SendMessage(fmt.Sprintf("Ошибка при ожидании появления результирующего элемента: %v | %s", err, genMsg)); err != nil {
			log.Println(err)
		}

		return
	}

	if cfg.DebugMode {
		log.Println("Проверка лида в CRM")
	}
	if leadUuid != "" {
		time.Sleep(10 * time.Second)

		if err := checkInCRM(leadUuid, cfg); err != nil {
			if err = tg.SendMessage(fmt.Sprintf("Ошибка при проверке лида в CRM: %v | %s", err, genMsg)); err != nil {
				log.Println(err)
			}

			return
		}
	}

	if err := tg.SendMessage(fmt.Sprintf("Проверка завершена успушно. Название: %s | Ссылка: %s \n", f.Name, f.Url)); err != nil {
		log.Println(err)
	}
}

func checkInCRM(leadUuid string, cfg *config.Cfg) error {
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

			time.Sleep(5 * time.Second)

			return clsr()
		}

		return nil
	}

	return clsr()
}
