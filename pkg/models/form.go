package form

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/bklv-kirill/go-site-form-checker/pkg/config"
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

	var genMsg string = fmt.Sprintf("Название: %s | Ссылка: %s\n", f.Name, f.Url)

	log.Printf("Проверка запущена. %s", genMsg)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	ctx, cancel = chromedp.NewRemoteAllocator(ctx, "http://127.0.0.1:9222")
	defer cancel()

	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	if cfg.DebugMode {
		chromedp.ListenTarget(ctx, func(ev interface{}) {
			if msg, ok := ev.(*runtime.EventConsoleAPICalled); ok {
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
	var err error = chromedp.Run(ctx,
		chromedp.Navigate(f.Url),
	)
	if err != nil {
		log.Printf("Ошибка при переходе на сайт: %v | %s", err, genMsg)
		return
	}

	if cfg.DebugMode {
		log.Printf("Ожидание: 5 секунд")
	}
	err = chromedp.Run(ctx,
		chromedp.Sleep(5*time.Second),
	)
	if err != nil {
		log.Printf("Ошибка при ожидании: %v | %s", err, genMsg)
		return
	}

	if cfg.DebugMode {
		log.Println("Ожидание появления элемента для нажатия")
	}
	wCtx, wCancel := context.WithTimeout(ctx, 15*time.Second)
	defer wCancel()
	err = chromedp.Run(wCtx,
		chromedp.WaitVisible(f.ElemForClick, chromedp.ByQuery),
	)
	if err != nil {
		log.Printf("Ошибка при ожидании элемента для нажатия: %v | %s", err, genMsg)
		return
	}

	if cfg.DebugMode {
		log.Println("Имитация клика на элемент для нажатия")
	}
	err = chromedp.Run(ctx,
		chromedp.Evaluate(fmt.Sprintf("document.querySelector('%s').click()", f.ElemForClick), nil),
	)
	if err != nil {
		log.Printf("Ошибка при клике на элемент для нажатия: %v | %s", err, genMsg)
		return
	}

	if cfg.DebugMode {
		log.Println("Ожидание появления элемента для взаимодействия")
	}
	wCtx, wCancel = context.WithTimeout(ctx, 15*time.Second)
	defer wCancel()
	err = chromedp.Run(wCtx,
		chromedp.WaitVisible(f.ExpElem, chromedp.ByQuery),
	)
	if err != nil {
		log.Printf("Ошибка при ожидании элемента для взаимодействия: %v | %s", err, genMsg)
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

		err = chromedp.Run(ctx,
			chromedp.SendKeys(fmt.Sprintf("%s %s", f.ExpElem, i.Selector), i.Value),
		)
		if err != nil {
			log.Printf("Ошибка при заполнении формы: %v | %s", err, genMsg)
			return
		}
	}

	if cfg.DebugMode {
		log.Println("Проверка формы")
		for _, i := range f.Inputs {
			var val string
			err = chromedp.Run(ctx,
				chromedp.Value(fmt.Sprintf("%s %s", f.ExpElem, i.Selector), &val, chromedp.ByQuery),
			)
			if err != nil {
				log.Printf("Ошибка при проверке формы: %v | %s", err, genMsg)
				return
			}

			log.Printf("%s: %s\n", i.Selector, val)
		}
	}

	if cfg.DebugMode {
		log.Println("Отправка формы")
	}
	err = chromedp.Run(ctx,
		chromedp.Evaluate(fmt.Sprintf("document.querySelector('%s %s').click()", f.ExpElem, f.SubmitElem), nil),
	)
	if err != nil {
		log.Printf("Ошибка при отправке формы: %v | %s", err, genMsg)
		return
	}

	if cfg.DebugMode {
		log.Println("Ожидание появления результирующего элемента")
	}
	wCtx, wCancel = context.WithTimeout(ctx, 15*time.Second)
	defer wCancel()
	err = chromedp.Run(wCtx,
		chromedp.WaitVisible(f.ResElem, chromedp.ByQuery),
	)
	if err != nil {
		log.Printf("Ошибка при ожидании появления результирующего элемента: %v | %s", err, genMsg)
		return
	}

	if cfg.DebugMode == false && leadUuid != "" {
		time.Sleep(10 * time.Second)

		if err = checkInCRM(leadUuid, cfg); err != nil {
			log.Printf("Ошибка при проверке лида в CRM: %v | %s", err, genMsg)
			return
		}
	}

	// ------------------------------------
	//log.Println("Получение содержимого <body>")
	//var body string
	//err = chromedp.Run(ctx,
	//	chromedp.OuterHTML("body", &body, chromedp.ByQuery),
	//)
	//if err != nil {
	//	log.Fatalf("Ошибка при получении содержимого <body>: %v | %s", err, genMsg)
	//	return
	//}
	//log.Printf("Содержимое <body>:\n%s\n", body)
	//------------------------------------

	log.Printf("Проверка завершена успушно. Название: %s | Ссылка: %s \n", f.Name, f.Url)
}

func checkInCRM(leadUuid string, cfg *config.Cfg) error {
	var data map[string]string = map[string]string{
		"uuid": leadUuid,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", cfg.CrmUrl, bytes.NewBuffer(jsonData))
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
