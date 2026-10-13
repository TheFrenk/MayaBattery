package main

import (
	"errors"
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/sstallion/go-hid"

	"maya-battery/internal/app"
	"maya-battery/internal/config"
	"maya-battery/internal/singleton"
)

func setupLogging() {
	dir := config.Dir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return
	}
	f, err := os.OpenFile(filepath.Join(dir, "maya.log"),
		os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	log.SetOutput(f)
}

func main() {
	console := flag.Bool("console", false, "писать логи в консоль, а не в файл")
	flag.Parse()
	if !*console {
		setupLogging()
	}

	release, err := singleton.Acquire(`Local\MayaBattery`)
	if err != nil {
		if errors.Is(err, singleton.ErrAlreadyRunning) {
			return
		}
		log.Fatal(err)
	}
	defer release()

	if err := hid.Init(); err != nil {
		log.Fatal("HID init: ", err)
	}
	defer hid.Exit()

	app.New(config.Load()).Run()
}
