package main

import (
	"github.com/bklv-kirill/go-site-form-checker/pkg/config"
	"github.com/bklv-kirill/go-site-form-checker/pkg/storage"
	"log"
	"sync"
)

func main() {
	var cfg *config.Cfg = config.New()

	strg, err := storage.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	fs, err := strg.GetAllWithInputs()
	if err != nil {
		log.Fatal(err)
	}

	var ch chan struct{} = make(chan struct{}, cfg.MaxGoroutines)
	var wg sync.WaitGroup = sync.WaitGroup{}
	for _, f := range fs {
		wg.Add(1)

		ch <- struct{}{}
		go f.Check(&wg, ch, cfg)
	}

	wg.Wait()
	log.Println("Выполнение скрипта завершено")
}
