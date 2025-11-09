package main

import (
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/alvin0319/go-stargate-client/config"
	"github.com/alvin0319/go-stargate-client/protocol"
	"github.com/alvin0319/go-stargate-client/server"
)

type CustomHandler struct {
	log *slog.Logger
}

func (h *CustomHandler) Handle(w *protocol.Wrapper) error {
	h.log.Info("received packet", "id", w.P.ID(), "response", w.Response, "responseID", w.ResponseID)
	return nil
}

func main() {
	conf, err := config.Read()
	if err != nil {
		panic(err)
	}
	slog.SetLogLoggerLevel(slog.LevelDebug)
	log := slog.Default()
	log.Info("starting stargate server", "host", conf.Host, "port", conf.Port)
	l, err := server.Listen(conf.Host+":"+strconv.Itoa(conf.Port), conf.Password)
	if err != nil {
		panic(err)
	}
	defer l.Close()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Info("received shutdown signal, closing server...")
		l.Close()
		os.Exit(0)
	}()

	for {
		c := l.Accept()
		c.Handler(&CustomHandler{log: log})
		log.Info("accepted session", "addr", c.RemoteAddr())
	}
}
