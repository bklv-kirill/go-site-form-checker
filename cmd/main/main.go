package main

import (
	"github.com/bklv-kirill/go-site-form-checker/pkg/config"
	"github.com/bklv-kirill/go-site-form-checker/pkg/storage"
	"log"
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

	for _, f := range fs {
		f.Check(cfg)
	}

	log.Println("Выполнение скрипта завершено")
}
