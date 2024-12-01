package main

import (
	"fmt"
	"github.com/bklv-kirill/go-site-form-checker/pkg/config"
	"github.com/bklv-kirill/go-site-form-checker/pkg/models"
	"github.com/bklv-kirill/go-site-form-checker/pkg/services"
	"github.com/bklv-kirill/go-site-form-checker/pkg/storage/form"
	"log"
	"sync"
)

func main() {
	var cfg *config.Config = config.New()

	formSqlStrg, err := formStorage.NewFormSqlStorage(cfg)
	if err != nil {
		log.Fatal(err)
	}

	fs, err := formSqlStrg.GetAllWithInputs()
	if err != nil {
		log.Fatal(err)
	}

	var fsndr *services.FormSender = services.NewFormSender(cfg)
	var crm *services.Crm = services.NewCrm(cfg)
	var tg *services.Telegram = services.NewTelegram(cfg)

	var wg sync.WaitGroup = sync.WaitGroup{}
	var ch chan struct{} = make(chan struct{}, cfg.MaxGoroutines)

	for _, f := range fs {
		wg.Add(1)

		ch <- struct{}{}

		go func(wg *sync.WaitGroup, ch chan struct{}, f *form.Form) {
			defer func() {
				wg.Done()
				<-ch
			}()

			leadUuid, err := fsndr.SendForm(f)
			if err != nil {
				if err := tg.SendMessage(err.Error()); err != nil {
					log.Println(err)
				}

				return
			}

			if err := crm.CheckLeadByUuid(leadUuid); err != nil {
				if err := tg.SendMessage(err.Error()); err != nil {
					log.Println(err)
				}

				return
			}

			if err := tg.SendMessage(fmt.Sprintf("Проверка формы успешно завершена | Название: %s | Ссылка: %s", f.Name, f.Url)); err != nil {
				log.Println(err)
			}
		}(&wg, ch, &f)
	}

	wg.Wait()
}
