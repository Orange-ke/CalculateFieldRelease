package main

import (
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"lz/conf"
	"lz/server"
	"net/http"
	"os"
)

func initLog(mode string) {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	// Only log the debug severity or above.
	if mode == "debug" {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
}

func main() {
	conf.Init()
	initLog(conf.AppConfig.Mode)

	var upgrader = websocket.Upgrader{
		ReadBufferSize:  conf.AppConfig.ReadBufferSize,
		WriteBufferSize: conf.AppConfig.WriteBufferSize,
	}

	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}
	s := server.NewServer(conf.AppConfig.Port, upgrader)
	s.Serve()
}
