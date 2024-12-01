package services

import (
	"context"
	"fmt"
	"github.com/bklv-kirill/go-site-form-checker/pkg/config"
	"github.com/bklv-kirill/go-site-form-checker/pkg/models"
	"github.com/chromedp/chromedp"
	"github.com/google/uuid"
	"log"
	"time"
)

type FormSender struct {
	DebugMode           bool   `json:"debug_mode"`
	RemoteBrowserSchema string `json:"remote_browser_schema"`
	RemoteBrowserUrl    string `json:"remote_browser_url"`
	RemoteBrowserPort   string `json:"remote_browser_port"`
	Attempts            int    `json:"attempts"`
	Timeout             int    `json:"timeout"`
	RetryDelay          int    `json:"retry_delay"`
}

func NewFormSender(cfg *config.Config) *FormSender {
	return &FormSender{
		DebugMode:           cfg.DebugMode,
		RemoteBrowserSchema: cfg.RemoteBrowserSchema,
		RemoteBrowserUrl:    cfg.RemoteBrowserUrl,
		RemoteBrowserPort:   cfg.RemoteBrowserPort,
		Attempts:            cfg.SendFormAttempts,
		Timeout:             cfg.SendFormTimeout,
		RetryDelay:          cfg.SendFormRetryDelay,
	}
}

func (fs *FormSender) SendForm(f *form.Form) (string, error) {
	var genMsg string = fmt.Sprintf("Название: %s | Ссылка: %s\n", f.Name, f.Url)

	for att := 1; att <= fs.Attempts; att++ {
		leadUuid, err := fs.execSendForm(f)
		if err == nil {
			return leadUuid, nil
		}

		if att == fs.Attempts {
			return "", fmt.Errorf("%s %s", err.Error(), genMsg)
		}

		if fs.DebugMode {
			log.Printf("Ошибка при отправки формы | %s | %s | Повторная попытка...", err.Error(), genMsg)
		}

		time.Sleep(time.Duration(fs.RetryDelay) * time.Second)
	}

	return "", fmt.Errorf("При отправке формы произошла неизвестная ошибка | %s", genMsg)
}

func (fs *FormSender) execSendForm(f *form.Form) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(fs.Timeout)*time.Second)
	defer cancel()

	ctx, cancel = chromedp.NewRemoteAllocator(
		ctx,
		fmt.Sprintf("%s://%s:%s", fs.RemoteBrowserSchema, fs.RemoteBrowserUrl, fs.RemoteBrowserPort),
	)
	defer cancel()

	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	if err := fs.goToSite(ctx, f.Url); err != nil {
		return "", err
	}

	if err := fs.chromedpWait(ctx, 5); err != nil {
		return "", err
	}

	if err := fs.waitElemForClick(ctx, f.ElemForClick); err != nil {
		return "", err
	}

	if err := fs.evaluateClickOnElem(ctx, f.ElemForClick); err != nil {
		return "", err
	}

	if err := fs.waitExpElem(ctx, f.ExpElem); err != nil {
		return "", err
	}

	leadUuid, err := fs.fillForm(ctx, f)
	if err != nil {
		return "", err
	}

	if err := fs.sendForm(ctx, f); err != nil {
		return "", err
	}

	if err := fs.waitResElem(ctx, f.ResElem); err != nil {
		return "", err
	}

	return leadUuid, nil
}

func (fs *FormSender) goToSite(ctx context.Context, url string) error {
	if fs.DebugMode {
		log.Printf("Переход на сайт %s\n", url)
	}
	if err := chromedp.Run(ctx, chromedp.Navigate(url)); err != nil {
		return fmt.Errorf("Ошибка при переходе на сайт: %s", err.Error())
	}

	return nil
}

func (fs *FormSender) chromedpWait(ctx context.Context, sec int) error {
	if fs.DebugMode {
		log.Printf("Ожидание: %d секунд\n", sec)
	}
	if err := chromedp.Run(ctx, chromedp.Sleep(time.Duration(sec)*time.Second)); err != nil {
		return fmt.Errorf("Ошибка при ожидании: %s", err.Error())
	}

	return nil
}

func (fs *FormSender) waitElemForClick(ctx context.Context, elemForClick string) error {
	if fs.DebugMode {
		log.Println("Ожидание появления элемента для нажатия")
	}
	if err := chromedp.Run(ctx, chromedp.WaitVisible(elemForClick, chromedp.ByQuery)); err != nil {
		return fmt.Errorf("Ошибка при ожидании элемента для нажатия: %s", err.Error())
	}

	return nil
}

func (fs *FormSender) evaluateClickOnElem(ctx context.Context, elemForClick string) error {
	if fs.DebugMode {
		log.Printf("Имитация клика на элемент для нажатия\n")
	}
	if err := chromedp.Run(ctx, chromedp.Evaluate(fmt.Sprintf("document.querySelector('%s').click()", elemForClick), nil)); err != nil {
		return fmt.Errorf("Ошибка при клике на элемент для нажатия: %s", err.Error())
	}

	return nil
}

func (fs *FormSender) waitExpElem(ctx context.Context, expElem string) error {
	if fs.DebugMode {
		log.Println("Ожидание появления элемента для взаимодействия")
	}
	if err := chromedp.Run(ctx, chromedp.WaitVisible(expElem, chromedp.ByQuery)); err != nil {
		return fmt.Errorf("Ошибка при ожидании элемента для взаимодействия: %s", err.Error())
	}

	return nil
}

func (fs *FormSender) fillForm(ctx context.Context, f *form.Form) (string, error) {
	if fs.DebugMode {
		log.Println("Заполнение формы")
	}
	var leadUuid string
	for _, i := range f.Inputs {
		if i.ForUuid {
			leadUuid = fmt.Sprintf("SFC - %s", uuid.New())
			i.Value = leadUuid
		}

		if err := chromedp.Run(ctx, chromedp.SendKeys(fmt.Sprintf("%s %s", f.ExpElem, i.Selector), i.Value)); err != nil {
			return leadUuid, fmt.Errorf("Ошибка при заполнении формы: %s", err.Error())
		}
	}

	return leadUuid, nil
}

func (fs *FormSender) sendForm(ctx context.Context, f *form.Form) error {
	if fs.DebugMode {
		log.Println("Отправка формы")
	}
	if err := chromedp.Run(ctx, chromedp.Evaluate(fmt.Sprintf("document.querySelector('%s %s').click()", f.ExpElem, f.SubmitElem), nil)); err != nil {
		return fmt.Errorf("Ошибка при отправке формы: %s", err.Error())
	}

	return nil
}

func (fs *FormSender) waitResElem(ctx context.Context, resElem string) error {
	if fs.DebugMode {
		log.Println("Ожидание появления результирующего элемента")
	}
	if err := chromedp.Run(ctx, chromedp.WaitVisible(resElem, chromedp.ByQuery)); err != nil {
		return fmt.Errorf("Ошибка при ожидании появления результирующего элемента: %s", err.Error())
	}

	return nil
}
