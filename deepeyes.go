package main

import (
	"context"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cgoder/deepeyes/api"
	gss "github.com/cgoder/deepeyes/gss"
	log "github.com/sirupsen/logrus"
)

func main() {

	svr := gss.NewService()

	log.WithFields(log.Fields{
		"module": "main",
		"func":   "main",
	}).Info("Server CORE start")

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
	}()

	// pprof debug
	go gss.DebugRuntime()

	// http service
	go api.HTTPAPIServer(svr)

	// go RTSPServer()

	// auto stream
	go svr.ChannelRunAll(ctx)

	signalChanel := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(signalChanel, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-signalChanel
		log.WithFields(log.Fields{
			"module": "main",
			"func":   "main",
		}).Info("Server receive signal", sig)
		done <- true
	}()
	log.WithFields(log.Fields{
		"module": "main",
		"func":   "main",
	}).Info("Server start success a wait signals")

	<-done

	svr.ProgramStopAll()
	time.Sleep(2 * time.Second)
	log.WithFields(log.Fields{
		"module": "main",
		"func":   "main",
	}).Info("Server stop working by signal")
}
