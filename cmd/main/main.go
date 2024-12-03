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
		log.Fatal(err.Error())
	}

	fs, err := formSqlStrg.GetAllWithInputs()
	if err != nil {
		log.Fatal(err.Error())
	}

	var fsndr *services.FormSender = services.NewFormSender(cfg)
	var crm *services.Crm = services.NewCrm(cfg)
	var tg *services.Telegram = services.NewTelegram(cfg)

	var wg sync.WaitGroup = sync.WaitGroup{}
	var ch chan struct{} = make(chan struct{}, cfg.MaxGoroutines)

	for _, f := range fs {
		wg.Add(1)
		ch <- struct{}{}

		go func(wg *sync.WaitGroup, ch <-chan struct{}, f *form.Form) {
			defer func() {
				wg.Done()
				<-ch
			}()

			var resMsg string

			leadUuid, err := fsndr.SendForm(f)
			if err != nil {
				resMsg = err.Error()
			} else if leadUuid != "" {
				if crm != nil {
					if err := crm.CheckLeadByUuid(leadUuid); err != nil {
						resMsg = err.Error()
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
