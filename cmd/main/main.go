package main

import (
	"github.com/bklv-kirill/go-site-form-checker/pkg/config"
	"github.com/bklv-kirill/go-site-form-checker/pkg/services"
	"github.com/bklv-kirill/go-site-form-checker/pkg/storage"
	"log"
	"sync"
)

func main() {
	var cfg *config.Config = config.New()

	strg, err := storage.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	fs, err := strg.GetAllWithInputs()
	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup = sync.WaitGroup{}
	var ch chan struct{} = make(chan struct{}, cfg.MaxGoroutines)
	for _, f := range fs {
		wg.Add(1)

		ch <- struct{}{}
		go sfc.Check(&f, &wg, ch)
	}

	wg.Wait()
	log.Println("Выполнение скрипта завершено")
}
