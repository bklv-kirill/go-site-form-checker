package main

import (
	"fmt"
	"github.com/bklv-kirill/go-site-form-checker/pkg/config"
	"github.com/bklv-kirill/go-site-form-checker/pkg/models"
	formRepo "github.com/bklv-kirill/go-site-form-checker/pkg/repo/form"
	"github.com/bklv-kirill/go-site-form-checker/pkg/services"
	"log"
	"sync"
)

func main() {
	var cfg *config.Config = config.New()

	formSqlRepo, err := formRepo.NewSqlRepo(cfg)
	if err != nil {
		log.Fatal(err.Error())
	}

	fs, err := formSqlRepo.GetAllWithInputs()
	if err != nil {
		log.Fatal(err.Error())
	}

	var formSender *services.FormSender = services.NewFormSender(cfg)
	var crm *services.Crm = services.NewCrm(cfg)
	var tg *services.Telegram = services.NewTelegram(cfg)

	var wg sync.WaitGroup = sync.WaitGroup{}
	var ch chan struct{} = make(chan struct{}, cfg.MaxGoroutines)

	for _, f := range fs {
		wg.Add(1)
		ch <- struct{}{}

		go func(wg *sync.WaitGroup, ch <-chan struct{}, f *models.Form) {
			defer func() {
				wg.Done()
				<-ch
			}()

			var resMsg string

			leadUuid, err := formSender.SendForm(f)
			if err != nil {
				resMsg = err.Error()
			} else if leadUuid != "" {
				if crm != nil {
					if err := crm.CheckLeadByUuid(leadUuid); err != nil {
						resMsg = err.Error()
					} else {
						resMsg = fmt.Sprintf("Форма успешно проверена | %s", f.GetPrevMsg())
					}
				} else {
					resMsg = fmt.Sprintf("Форма успешно отправленна | %s", f.GetPrevMsg())
				}
			} else {
				resMsg = fmt.Sprintf("Форма успешно отправленна (без uuid) | %s", f.GetPrevMsg())
			}

			if tg != nil {
				if err := tg.SendMessage(resMsg); err != nil {
					log.Println(err.Error())
				}
			} else {
				log.Println(resMsg)
			}
		}(&wg, ch, &f)
	}

	wg.Wait()
}
