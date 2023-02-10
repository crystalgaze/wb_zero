package main

import (
	"fmt"
	"os"
	"os/signal"
	"wb_zero/api"
	"wb_zero/configs"
	"wb_zero/internal/store"
	"wb_zero/internal/stream"
)

func main() {
	configs.ConfigSetup()
	dbObject := store.NewStore()
	csh := store.NewCache(dbObject)
	sh := stream.NewStream(dbObject)
	myApi := api.NewApi(csh)
	signalChan := make(chan os.Signal, 1)
	cleanupDone := make(chan bool)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		for range signalChan {
			fmt.Printf("\nОтписка и закрытие соединения\n\n")
			csh.Finish()
			sh.Finish()
			myApi.Finish()
			cleanupDone <- true
		}
	}()
	<-cleanupDone
}
